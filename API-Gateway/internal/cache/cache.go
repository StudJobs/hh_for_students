// Package cache реализует cache-aside поверх Redis для GET-эндпоинтов API-Gateway.
//
// Whitelist-подход: кэшируем только маршруты, явно перечисленные в Cacheable.
// Это безопаснее «кэшировать всё подряд» — owner-specific эндпоинты (`/users/me`,
// `/user/achievements/`, `/expert/queue`, `/tasks/my-submissions`) сознательно
// исключены, чтобы пользователи не получали данные друг друга.
//
// Инвалидация — по prefix-pattern: при POST/PATCH/DELETE на `/api/v1/<resource>`
// чистим все ключи `gw:GET:/api/v1/<resource>*` через `SCAN + UNLINK`.
//
// Trade-off: SCAN+UNLINK не атомарен и пройдёт по всем ключам, но 256 MB Redis
// с allkeys-lru держит ~50-100k ключей — обход за <100 мс на типичной нагрузке.
// Альтернатива (теги через MGET по indirect-keys) сложнее и не нужна на демо.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "gw:GET:"
)

// Entry — то, что кладём в Redis. Тело JSON + content-type + status code.
type Entry struct {
	Status      int               `json:"s"`
	ContentType string            `json:"ct"`
	Body        []byte            `json:"b"`
	Headers     map[string]string `json:"h,omitempty"`
}

type Client struct {
	rdb *redis.Client
	ttl time.Duration
}

// New возвращает клиент. addr пустой — отключённый клиент (no-op), это позволяет
// gateway работать без Redis для локального dev.
func New(addr string, ttl time.Duration) *Client {
	if addr == "" {
		return &Client{}
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		PoolSize:     50,
	})
	return &Client{rdb: rdb, ttl: ttl}
}

// Enabled true если есть рабочий клиент.
func (c *Client) Enabled() bool { return c != nil && c.rdb != nil }

// Ping — health-check. Используется при старте gateway, чтобы не выйти из строя
// при недоступном Redis (просто отключаем кэш и логируем warn).
func (c *Client) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return fmt.Errorf("redis client not configured")
	}
	return c.rdb.Ping(ctx).Err()
}

// Key возвращает ключ для GET-запроса с учётом query-string.
// Auth-зависимые маршруты ДОЛЖНЫ быть исключены на уровне whitelist —
// сюда попадают только публично-кэшируемые URL'ы.
func Key(path, rawQuery string) string {
	if rawQuery == "" {
		return keyPrefix + path
	}
	return keyPrefix + path + "?" + rawQuery
}

// Get возвращает Entry, или nil если не найдено / Redis недоступен.
func (c *Client) Get(ctx context.Context, key string) *Entry {
	if !c.Enabled() {
		return nil
	}
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil // miss или ошибка сети — возвращаем nil как «cache miss»
	}
	var e Entry
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil
	}
	return &e
}

// Set кладёт Entry с TTL. Ошибки игнорируем — cache не критичен.
func (c *Client) Set(ctx context.Context, key string, e *Entry) {
	if !c.Enabled() {
		return
	}
	raw, err := json.Marshal(e)
	if err != nil {
		return
	}
	_ = c.rdb.Set(ctx, key, raw, c.ttl).Err()
}

// InvalidatePrefix чистит все ключи с указанным префиксом через SCAN.
// Используется при write-операциях (POST/PATCH/DELETE) на ресурс.
//
// `resource` ожидается без `gw:GET:` — функция добавляет его сама. Пример:
// `InvalidatePrefix(ctx, "/api/v1/tasks")` вычистит `/api/v1/tasks`,
// `/api/v1/tasks?status=open`, `/api/v1/tasks/abc-123` и т.д.
func (c *Client) InvalidatePrefix(ctx context.Context, resource string) {
	if !c.Enabled() {
		return
	}
	pattern := keyPrefix + resource + "*"
	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	var batch []string
	for iter.Next(ctx) {
		batch = append(batch, iter.Val())
		if len(batch) >= 100 {
			_ = c.rdb.Unlink(ctx, batch...).Err()
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		_ = c.rdb.Unlink(ctx, batch...).Err()
	}
}

// IsCacheableRoute возвращает true для GET-маршрутов, которые безопасно кэшировать
// (не зависят от авторизованного пользователя).
//
// Маршруты с :id (`/api/v1/users/abc-123`) тоже кэшируемые — они read-only и
// одинаковы для всех. Owner-specific (`/users/me`, `/user/achievements/`) — нет.
func IsCacheableRoute(path string) bool {
	for _, allow := range allowedPrefixes {
		if strings.HasPrefix(path, allow) {
			return true
		}
	}
	return false
}

// InvalidationsForWrite возвращает префиксы, которые нужно почистить при
// write-операции на данный путь. Например, POST `/api/v1/hr/tasks/` чистит
// `/api/v1/tasks` (студенты сразу видят новую задачу) и `/api/v1/hr/tasks` (HR
// видит свою же).
func InvalidationsForWrite(path string) []string {
	var out []string
	for prefix, related := range writeInvalidations {
		if strings.HasPrefix(path, prefix) {
			out = append(out, related...)
		}
	}
	return out
}

// allowedPrefixes — публичные read-only ресурсы.
//
// Намеренно НЕ кэшируем: /users/me, /user/achievements/, /expert/queue,
// /tasks/my-submissions, /hr/me, /hr/tasks, /hr/vacancy, /company/me — все они
// owner-specific и требуют авторизованного контекста.
var allowedPrefixes = []string{
	"/api/v1/skills/popular",
	"/api/v1/skills/search",
	"/api/v1/skills/bulk",
	"/api/v1/users",   // SUFFIX-проверка: исключаем /users/me, см. ShouldExclude
	"/api/v1/vacancy", // только GET-листинги, write-маршруты в /hr/vacancy
	"/api/v1/tasks",   // /tasks/my-submissions исключаем через ShouldExclude
	"/api/v1/company",
}

// ShouldExclude — финальный фильтр. Маршруты в whitelist могут содержать
// owner-specific под-ресурсы, и их надо явно вычислить.
func ShouldExclude(path string) bool {
	exclusions := []string{
		"/api/v1/users/me",
		"/api/v1/user/", // /user/achievements/...
		"/api/v1/tasks/my-submissions",
		"/api/v1/tasks/mine",
		"/api/v1/chat/",
		"/api/v1/expert/",
		"/api/v1/hr/",
		"/api/v1/auth/",
		"/api/v1/company/me",
		"/api/v1/company/membership/",
		"/api/v1/company/members",
	}
	for _, ex := range exclusions {
		if strings.HasPrefix(path, ex) {
			return true
		}
	}
	return false
}

// writeInvalidations — карта prefix → список префиксов для чистки.
// Написана так, чтобы read-маршруты увидели свежие данные после write.
var writeInvalidations = map[string][]string{
	"/api/v1/hr/vacancy": {"/api/v1/vacancy"},
	"/api/v1/hr/tasks":   {"/api/v1/tasks"},
	"/api/v1/users/edit": {"/api/v1/users"},
	"/api/v1/company":    {"/api/v1/company"},
}
