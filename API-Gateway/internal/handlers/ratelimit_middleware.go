package handlers

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/metrics"
	"golang.org/x/time/rate"
)

// RateLimiter — token-bucket с per-user (или per-IP fallback) ключом.
// Используется как Fiber middleware.
//
// Дизайн-выбор:
//   - golang.org/x/time/rate — стандарт-де-факто, lock-free, не требует Redis;
//     trade-off — счётчики не разделяются между инстансами Gateway. Для одного
//     инстанса (наш случай) допустимо. Под кластер пришлось бы взять redis_rate
//     либо token-bucket в Lua.
//   - Ключ: если есть валидный по форме JWT с непустым sub — используем
//     "u:<userID>" (читаем payload без проверки подписи; bucket-ключ не должен
//     быть доверенным — даже если злоумышленник подделает токен, он получит
//     отдельный bucket с тем же лимитом, защита auth-middleware'а на следующем
//     шаге отбросит запрос). Иначе — "ip:<addr>" для public-роутов.
//     Per-user важен для локалки и SPA с многими вкладками: вкладки одного юзера
//     ходят с одного IP, и per-IP bucket бы их объединял в общий лимит.
//   - sweep раз в 5 мин чистит мапу от мёртвых ключей, чтобы не лизалась память.
type RateLimiter struct {
	rps     rate.Limit
	burst   int
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	limiter *rate.Limiter
	seenAt  time.Time
}

// NewRateLimiter возвращает ratelimiter. perMin = разрешённых запросов в минуту,
// burst = насколько единомоментный всплеск разрешён.
func NewRateLimiter(perMin, burst int) *RateLimiter {
	rl := &RateLimiter{
		rps:     rate.Limit(float64(perMin) / 60.0),
		burst:   burst,
		buckets: make(map[string]*bucket),
	}
	go rl.sweep()
	return rl
}

// Middleware возвращает Fiber-handler. Извлекает ключ (user или IP), поднимает
// bucket для него и проверяет .Allow().
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := rateLimitKey(c)
		bucket := rl.bucketFor(key)
		if !bucket.Allow() {
			route := c.Route().Path
			if route == "" {
				route = c.Path()
			}
			metrics.RateLimitThrottled.WithLabelValues(route).Inc()
			// Retry-After в секундах: обратно к rate (если 600/min, то 1 token = 0.1s → округлим до 1s).
			retryAfter := int(60.0 / float64(rl.rps*60))
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Set("Retry-After", strconv.Itoa(retryAfter))
			c.Set("X-RateLimit-Limit", strconv.FormatFloat(float64(rl.rps)*60, 'f', 0, 64))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "too many requests",
				"retry_after": retryAfter,
			})
		}
		return c.Next()
	}
}

func (rl *RateLimiter) bucketFor(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.buckets[key] = b
	}
	b.seenAt = time.Now()
	return b.limiter
}

// sweep периодически чистит buckets неактивных ключей, чтобы не было утечки.
func (rl *RateLimiter) sweep() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		rl.mu.Lock()
		for k, b := range rl.buckets {
			if b.seenAt.Before(cutoff) {
				delete(rl.buckets, k)
			}
		}
		rl.mu.Unlock()
	}
}

// rateLimitKey возвращает ключ bucket'а: user-id из JWT или IP-фоллбек.
// JWT не верифицируется криптографически — на этом этапе нам нужен только
// stable identifier, и auth-middleware на следующем шаге отвергнет невалидный
// токен (так что подделкой обойти лимит нельзя без обхода auth, что отдельная
// граница защиты).
func rateLimitKey(c *fiber.Ctx) string {
	if userID := userIDFromAuthHeader(c.Get("Authorization")); userID != "" {
		return "u:" + userID
	}
	return "ip:" + clientIP(c)
}

// userIDFromAuthHeader парсит JWT payload (без проверки подписи) и достаёт sub.
// Используется ТОЛЬКО для определения rate-limit-ключа.
func userIDFromAuthHeader(header string) string {
	if header == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	token := header[len(prefix):]
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Sub    string `json:"sub"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	if claims.Sub != "" {
		return claims.Sub
	}
	return claims.UserID
}

// clientIP возвращает IP клиента, учитывая X-Forwarded-For (от reverse proxy).
// Берём первое значение из CSV, чтобы избежать spoofing промежуточными прокси.
func clientIP(c *fiber.Ctx) string {
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		// Первое значение — клиент, остальные — прокси.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return trimSpace(xff[:i])
			}
		}
		return trimSpace(xff)
	}
	if real := c.Get("X-Real-IP"); real != "" {
		return trimSpace(real)
	}
	return c.IP()
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
