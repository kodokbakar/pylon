package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsHandlerReturnsPrometheusMetrics(t *testing.T) {
	InitMetrics()

	httpRequestsTotal.WithLabelValues("GET", "/health", "200").Inc()
	httpRequestDuration.WithLabelValues("GET", "/health", "200").Observe(0.01)
	grpcRequestsTotal.WithLabelValues("chat", "SendMessage", "ok").Inc()
	grpcRequestDuration.WithLabelValues("chat", "SendMessage", "ok").Observe(0.01)
	kafkaMessagesTotal.WithLabelValues("message-events", "message.created", "ok").Inc()
	wsConnectionsActive.Set(1)

	handler := MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	expectedMetrics := []string{
		"pylon_http_requests_total",
		"pylon_http_request_duration_seconds",
		"pylon_grpc_requests_total",
		"pylon_grpc_request_duration_seconds",
		"pylon_kafka_messages_total",
		"pylon_ws_connections_active",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Fatalf("expected %s metric, got body %q", metric, body)
		}
	}
}
