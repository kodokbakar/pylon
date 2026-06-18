package observability

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsRegistry = prometheus.NewRegistry()
	metricsOnce     sync.Once

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pylon_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	grpcRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_grpc_requests_total",
			Help: "Total number of gRPC requests.",
		},
		[]string{"service", "method", "status"},
	)

	grpcRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pylon_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)

	kafkaMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_kafka_messages_total",
			Help: "Total number of Kafka messages.",
		},
		[]string{"topic", "event_type", "status"},
	)

	wsConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "pylon_ws_connections_active",
			Help: "Number of active WebSocket connections.",
		},
	)
)

func InitMetrics() {
	metricsOnce.Do(func() {
		metricsRegistry.MustRegister(
			httpRequestsTotal,
			httpRequestDuration,
			grpcRequestsTotal,
			grpcRequestDuration,
			kafkaMessagesTotal,
			wsConnectionsActive,
		)
	})
}

func MetricsHandler() http.Handler {
	InitMetrics()

	return promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
}
