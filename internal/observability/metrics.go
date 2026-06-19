package observability

import (
	"net/http"

	"github.com/kodokbakar/pylon/internal/metrics"
)

func InitMetrics() {
	// Metrics are registered during internal/metrics package initialization.
}

func MetricsHandler() http.Handler {
	return metrics.Handler()
}
