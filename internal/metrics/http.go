package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.With(defaultRegistry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "pylon_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status_code"},
	)

	httpRequestDuration = promauto.With(defaultRegistry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pylon_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = promauto.With(defaultRegistry).NewGauge(
		prometheus.GaugeOpts{
			Name: "pylon_http_requests_in_flight",
			Help: "Number of HTTP requests currently in progress.",
		},
	)
)

func ObserveHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(
		normalizeLabel(method, "UNKNOWN"),
		normalizeLabel(path, "unknown"),
		strconv.Itoa(statusCode),
	).Inc()

	httpRequestDuration.WithLabelValues(
		normalizeLabel(method, "UNKNOWN"),
		normalizeLabel(path, "unknown"),
	).Observe(duration.Seconds())
}

func IncHTTPRequestsInFlight() {
	httpRequestsInFlight.Inc()
}

func DecHTTPRequestsInFlight() {
	httpRequestsInFlight.Dec()
}
