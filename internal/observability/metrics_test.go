package observability

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	internalmetrics "github.com/kodokbakar/pylon/internal/metrics"
)

func TestMetricsHandlerReturnsPrometheusMetrics(t *testing.T) {
	internalmetrics.ObserveHTTPRequest("GET", "/health", http.StatusOK, 10*time.Millisecond)
	internalmetrics.IncHTTPRequestsInFlight()
	internalmetrics.DecHTTPRequestsInFlight()

	internalmetrics.ObserveGRPCRequest("chat-service", "SendMessage", http.StatusOK, 10*time.Millisecond)

	internalmetrics.RecordKafkaMessagePublished("message-events")
	internalmetrics.RecordKafkaMessageConsumed("message-events", "notification-consumer-group")
	internalmetrics.ObserveKafkaPublish("message-events", 10*time.Millisecond)

	internalmetrics.IncActiveConnections(internalmetrics.RoomTypeAll)
	internalmetrics.DecActiveConnections(internalmetrics.RoomTypeAll)
	internalmetrics.RecordMessageSent(internalmetrics.RoomTypeAll, "text")
	internalmetrics.RecordRoomCreated()
	internalmetrics.SetUsersOnline(1)

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
		"pylon_http_requests_in_flight",
		"pylon_grpc_requests_total",
		"pylon_grpc_request_duration_seconds",
		"pylon_kafka_messages_published_total",
		"pylon_kafka_messages_consumed_total",
		"pylon_kafka_publish_duration_seconds",
		"pylon_websocket_connections_active",
		"pylon_messages_sent_total",
		"pylon_rooms_created_total",
		"pylon_users_online",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(body, metric) {
			t.Fatalf("expected %s metric, got body %q", metric, body)
		}
	}
}
