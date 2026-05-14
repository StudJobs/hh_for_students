// Package metrics экспонирует Prometheus-метрики API-Gateway. HTTP-латентность через
// Fiber-middleware, плюс счётчики для Phase 3 (cache) и Phase 4 (rate limit).
package metrics

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serviceLabel = "api-gateway"

var (
	registry = prometheus.NewRegistry()

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Histogram of HTTP request latency at the gateway edge.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"service", "method", "route", "status"})

	// CacheHits / CacheMisses — заполняются Phase 3 (Redis cache-aside middleware).
	CacheHits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_cache_hits_total",
		Help: "Number of GET responses served from Redis cache.",
	}, []string{"route"})

	CacheMisses = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_cache_misses_total",
		Help: "Number of GET responses that missed Redis cache and went upstream.",
	}, []string{"route"})

	// RateLimitThrottled — заполняется Phase 4 (token-bucket middleware).
	RateLimitThrottled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_ratelimit_throttled_total",
		Help: "Number of requests rejected with 429 by rate limiter.",
	}, []string{"route"})
)

func init() {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		httpDuration,
		CacheHits,
		CacheMisses,
		RateLimitThrottled,
	)
}

// HTTPMiddleware returns a Fiber middleware that records latency / status code per route.
// `route` берём из c.Route().Path — это маршрут с :id-плейсхолдерами, не конкретный URL.
// Это критично: иначе кардинальность лейблов взорвётся (каждый UUID = отдельный лейбл).
//
// `method` снимаем ДО c.Next() и копируем в свежую строку: c.Method() в Fiber возвращает
// unsafe-string поверх shared fasthttp-буфера. Если читать после c.Next() (когда контекст
// уже мог быть переиспользован для следующего запроса), получаем мусор вроде "POSTH".
// Это и был баг B2 — Prometheus падал, потому что метка `method` рандомно склеивалась.
func HTTPMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		method := normalizeMethod(string([]byte(c.Method())))

		err := c.Next()
		dur := time.Since(start).Seconds()

		route := c.Route().Path
		if route == "" {
			route = "unknown"
		}
		status := strconv.Itoa(c.Response().StatusCode())
		httpDuration.WithLabelValues(serviceLabel, method, route, status).Observe(dur)
		return err
	}
}

// normalizeMethod ограничивает кардинальность лейбла method стандартными HTTP-методами.
// Любые странные значения (например, мусор из shared-buffer'а fasthttp) превращаются в OTHER —
// это страхует /metrics от взрыва уникальных серий и от ошибок Prometheus типа
// "collected metric ... was collected before with the same name and label values".
func normalizeMethod(m string) string {
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return m
	}
	return "OTHER"
}

// ServeMetrics запускает /metrics на отдельном порту в фоне.
func ServeMetrics(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("metrics endpoint listening on %s/metrics", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("metrics server error: %v", err)
		}
	}()
}
