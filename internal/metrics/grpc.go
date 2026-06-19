package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	grpcRequestsTotal = promauto.With(defaultRegistry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_grpc_requests_total",
			Help: "Total number of gRPC or Connect RPC requests.",
		},
		[]string{"service", "method", "status_code"},
	)

	grpcRequestDuration = promauto.With(defaultRegistry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pylon_grpc_request_duration_seconds",
			Help:    "Duration of gRPC or Connect RPC requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)
)

func ObserveGRPCRequest(service, method string, statusCode int, duration time.Duration) {
	grpcRequestsTotal.WithLabelValues(
		normalizeLabel(service, "unknown"),
		normalizeLabel(method, "unknown"),
		strconv.Itoa(statusCode),
	).Inc()

	grpcRequestDuration.WithLabelValues(
		normalizeLabel(service, "unknown"),
		normalizeLabel(method, "unknown"),
	).Observe(duration.Seconds())
}
