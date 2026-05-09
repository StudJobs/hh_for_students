// Package metrics — Prometheus-инструментация Company-сервиса.
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

const serviceLabel = "company"

var (
	registry = prometheus.NewRegistry()

	grpcDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "grpc_server_handling_seconds",
		Help:    "Histogram of response latency (seconds) for gRPC server method handling.",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
	}, []string{"service", "grpc_method", "code"})

	grpcHandled = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_server_handled_total",
		Help: "Total number of RPCs completed on the server.",
	}, []string{"service", "grpc_method", "code"})
)

func init() {
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		grpcDuration,
		grpcHandled,
	)
}

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
		grpcDuration.WithLabelValues(serviceLabel, info.FullMethod, code).Observe(time.Since(start).Seconds())
		grpcHandled.WithLabelValues(serviceLabel, info.FullMethod, code).Inc()
		return resp, err
	}
}

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
