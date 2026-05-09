package handlers

import (
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/metrics"
	"golang.org/x/time/rate"
)

// RateLimiter — token-bucket per-IP. Используется как Fiber middleware.
//
// Дизайн-выбор:
//   - golang.org/x/time/rate — стандарт-де-факто, lock-free, не требует Redis;
//     trade-off — счётчики не разделяются между инстансами Gateway. Для одного
//     инстанса (наш случай) допустимо. Под кластер пришлось бы взять
//     redis_rate либо token-bucket в Lua.
//   - 100 req/min с burst 20 — комфортно для UI (пользователь не успевает
//     накликать), но достаточно жёстко для botnet'а на /api/v1/auth/login.
//   - sweep раз в 5 мин чистит мапу от мёртвых IP, чтобы не лизалась память.
type RateLimiter struct {
	rps     rate.Limit
	burst   int
	mu      sync.Mutex
	buckets map[string]*ipBucket
}

type ipBucket struct {
	limiter *rate.Limiter
	seenAt  time.Time
}

// NewRateLimiter возвращает ratelimiter. perMin = разрешённых запросов в минуту,
// burst = насколько единомоментный всплеск разрешён.
func NewRateLimiter(perMin, burst int) *RateLimiter {
	rl := &RateLimiter{
		rps:     rate.Limit(float64(perMin) / 60.0),
		burst:   burst,
		buckets: make(map[string]*ipBucket),
	}
	go rl.sweep()
	return rl
}

// Middleware возвращает Fiber-handler. Извлекает IP из X-Forwarded-For (если
// есть) или RemoteIP, поднимает per-IP bucket и проверяет .Allow().
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := clientIP(c)
		bucket := rl.bucketFor(ip)
		if !bucket.Allow() {
			route := c.Route().Path
			if route == "" {
				route = c.Path()
			}
			metrics.RateLimitThrottled.WithLabelValues(route).Inc()
			// Retry-After в секундах: обратно к rate (если 100/min, то 1 token = 0.6s).
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

func (rl *RateLimiter) bucketFor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[ip]
	if !ok {
		b = &ipBucket{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.buckets[ip] = b
	}
	b.seenAt = time.Now()
	return b.limiter
}

// sweep периодически чистит buckets неактивных IP, чтобы не было утечки.
func (rl *RateLimiter) sweep() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		rl.mu.Lock()
		for ip, b := range rl.buckets {
			if b.seenAt.Before(cutoff) {
				delete(rl.buckets, ip)
			}
		}
		rl.mu.Unlock()
	}
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
