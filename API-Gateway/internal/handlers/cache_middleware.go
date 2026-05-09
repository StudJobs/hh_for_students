package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/studjobs/hh_for_students/api-gateway/internal/cache"
	"github.com/studjobs/hh_for_students/api-gateway/internal/metrics"
)

// CacheMiddleware реализует cache-aside поверх Redis.
//
// Логика:
//   1. GET-запрос на cacheable-маршрут → check Redis → hit → вернуть из кэша
//   2. miss → c.Next() → если 200 OK, положить body в Redis с TTL
//   3. POST/PATCH/PUT/DELETE → c.Next() → если 2xx, инвалидировать связанные prefix'ы
//
// Лейбл `route` для метрик берётся из c.Route().Path (без UUID), как в
// metrics.HTTPMiddleware — тот же паттерн.
func CacheMiddleware(cli *cache.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		method := c.Method()

		// Read-path: только GET, только whitelist.
		if method == fiber.MethodGet && cache.IsCacheableRoute(path) && !cache.ShouldExclude(path) {
			key := cache.Key(path, string(c.Request().URI().QueryString()))
			route := c.Route().Path
			if route == "" {
				route = path
			}
			if entry := cli.Get(c.Context(), key); entry != nil {
				// HIT — отдаём ответ как есть, не вызываем c.Next() (хендлер не нужен).
				metrics.CacheHits.WithLabelValues(route).Inc()
				if entry.ContentType != "" {
					c.Set(fiber.HeaderContentType, entry.ContentType)
				}
				c.Set("X-Cache", "HIT")
				return c.Status(entry.Status).Send(entry.Body)
			}
			// MISS — пускаем дальше, потом сохраняем.
			metrics.CacheMisses.WithLabelValues(route).Inc()
			if err := c.Next(); err != nil {
				return err
			}
			status := c.Response().StatusCode()
			if status == fiber.StatusOK {
				body := append([]byte(nil), c.Response().Body()...) // copy, body re-used
				cli.Set(c.Context(), key, &cache.Entry{
					Status:      status,
					ContentType: string(c.Response().Header.ContentType()),
					Body:        body,
				})
			}
			c.Set("X-Cache", "MISS")
			return nil
		}

		// Write-path: POST/PATCH/PUT/DELETE — пропускаем хендлер, потом инвал.
		if method != fiber.MethodGet {
			if err := c.Next(); err != nil {
				return err
			}
			status := c.Response().StatusCode()
			if status >= 200 && status < 300 {
				for _, prefix := range cache.InvalidationsForWrite(path) {
					cli.InvalidatePrefix(c.Context(), prefix)
				}
			}
			return nil
		}

		// Не cacheable GET — просто пропускаем.
		return c.Next()
	}
}
