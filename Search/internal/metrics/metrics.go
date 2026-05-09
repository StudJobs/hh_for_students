// Package metrics обвешивает Search-сервис Prometheus-метриками.
// Экспонируется на отдельном HTTP-порту (см. cmd/main.go), чтобы не пересекаться с
// gRPC-портом 50057. Имя меток совместимо с дашбордом StudJobs · Overview.
package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serviceLabel = "search"

var (
	registry = prometheus.NewRegistry()

	grpcDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_server_handling_seconds",
		Help:    "Histogram of response latency (seconds) for gRPC server method handling.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 14), // 1ms..16s
	}, []string{"service", "grpc_method", "code"})

	grpcHandled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_server_handled_total",
		Help: "Total number of RPCs completed on the server, regardless of success or failure.",
	}, []string{"service", "grpc_method", "code"})

	// ESQueryDuration — кастомная метрика поверх gRPC: показывает время самого ES-запроса
	// без накладных расходов на сериализацию protobuf. Op = search|index|reindex.
	ESQueryDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "search_es_query_duration_seconds",
		Help:    "Histogram of Elasticsearch query duration (seconds) by operation type.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"op", "index"})
)

func init() {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		grpcDuration,
		grpcHandled,
		ESQueryDuration,
	)
}

// UnaryInterceptor returns a gRPC unary interceptor that records latency + status code.
func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := codes.OK.String()
		if err != nil {
			if s, ok := status.FromError(err); ok {
				code = s.Code().String()
			} else {
				code = codes.Unknown.String()
			}
		}
		dur := time.Since(start).Seconds()
		grpcDuration.WithLabelValues(serviceLabel, info.FullMethod, code).Observe(dur)
		grpcHandled.WithLabelValues(serviceLabel, info.FullMethod, code).Inc()
		return resp, err
	}
}

// ServeMetrics запускает HTTP-эндпоинт /metrics в фоне.
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

// ObserveES — хелпер для замера длительности ES-запросов.
// Использование: defer metrics.ObserveES("search", "profiles")(time.Now())
func ObserveES(op, index string) func(time.Time) {
	return func(start time.Time) {
		ESQueryDuration.WithLabelValues(op, index).Observe(time.Since(start).Seconds())
	}
}
