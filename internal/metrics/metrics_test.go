package metrics

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestHTTPMetricsRecordCounterGaugeAndHistogram(t *testing.T) {
	path := "/unit-test/" + safeMetricLabel(t.Name())
	method := "PATCH"
	statusCode := "299"

	beforeCounter := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, statusCode))
	beforeHistogramCount := testutil.CollectAndCount(httpRequestDuration)
	beforeInFlight := testutil.ToFloat64(httpRequestsInFlight)

	ObserveHTTPRequest(method, path, http.StatusMultipleChoices-1, 25*time.Millisecond)

	afterCounter := testutil.ToFloat64(httpRequestsTotal.WithLabelValues(method, path, statusCode))
	if afterCounter != beforeCounter+1 {
		t.Fatalf("expected http counter to increment by 1, before=%f after=%f", beforeCounter, afterCounter)
	}

	afterHistogramCount := testutil.CollectAndCount(httpRequestDuration)
	if afterHistogramCount != beforeHistogramCount+1 {
		t.Fatalf("expected http histogram series count to increase by 1, before=%d after=%d", beforeHistogramCount, afterHistogramCount)
	}

	IncHTTPRequestsInFlight()
	afterInc := testutil.ToFloat64(httpRequestsInFlight)
	if afterInc != beforeInFlight+1 {
		t.Fatalf("expected in-flight gauge to increment by 1, before=%f after=%f", beforeInFlight, afterInc)
	}

	DecHTTPRequestsInFlight()
	afterDec := testutil.ToFloat64(httpRequestsInFlight)
	if afterDec != beforeInFlight {
		t.Fatalf("expected in-flight gauge to return to %f, got %f", beforeInFlight, afterDec)
	}
}

func TestGRPCMetricsRecordCounterAndHistogram(t *testing.T) {
	service := "unit-service-" + safeMetricLabel(t.Name())
	method := "UnitMethod"
	statusCode := "207"

	beforeCounter := testutil.ToFloat64(grpcRequestsTotal.WithLabelValues(service, method, statusCode))
	beforeHistogramCount := testutil.CollectAndCount(grpcRequestDuration)

	ObserveGRPCRequest(service, method, http.StatusMultiStatus, 30*time.Millisecond)

	afterCounter := testutil.ToFloat64(grpcRequestsTotal.WithLabelValues(service, method, statusCode))
	if afterCounter != beforeCounter+1 {
		t.Fatalf("expected grpc counter to increment by 1, before=%f after=%f", beforeCounter, afterCounter)
	}

	afterHistogramCount := testutil.CollectAndCount(grpcRequestDuration)
	if afterHistogramCount != beforeHistogramCount+1 {
		t.Fatalf("expected grpc histogram series count to increase by 1, before=%d after=%d", beforeHistogramCount, afterHistogramCount)
	}
}

func TestKafkaMetricsRecordCountersAndHistogram(t *testing.T) {
	topic := "unit-topic-" + safeMetricLabel(t.Name())
	consumerGroup := "unit-group-" + safeMetricLabel(t.Name())

	beforePublished := testutil.ToFloat64(kafkaMessagesPublished.WithLabelValues(topic))
	beforeConsumed := testutil.ToFloat64(kafkaMessagesConsumed.WithLabelValues(topic, consumerGroup))
	beforeHistogramCount := testutil.CollectAndCount(kafkaPublishDuration)

	RecordKafkaMessagePublished(topic)
	RecordKafkaMessageConsumed(topic, consumerGroup)
	ObserveKafkaPublish(topic, 40*time.Millisecond)

	afterPublished := testutil.ToFloat64(kafkaMessagesPublished.WithLabelValues(topic))
	if afterPublished != beforePublished+1 {
		t.Fatalf("expected published counter to increment by 1, before=%f after=%f", beforePublished, afterPublished)
	}

	afterConsumed := testutil.ToFloat64(kafkaMessagesConsumed.WithLabelValues(topic, consumerGroup))
	if afterConsumed != beforeConsumed+1 {
		t.Fatalf("expected consumed counter to increment by 1, before=%f after=%f", beforeConsumed, afterConsumed)
	}

	afterHistogramCount := testutil.CollectAndCount(kafkaPublishDuration)
	if afterHistogramCount != beforeHistogramCount+1 {
		t.Fatalf("expected kafka publish histogram series count to increase by 1, before=%d after=%d", beforeHistogramCount, afterHistogramCount)
	}
}

func TestBusinessMetricsRecordValues(t *testing.T) {
	roomType := "direct-" + safeMetricLabel(t.Name())
	messageType := "text"

	beforeConnections := testutil.ToFloat64(activeConnections.WithLabelValues(roomType))
	beforeMessages := testutil.ToFloat64(messagesSent.WithLabelValues(roomType, messageType))
	beforeRooms := testutil.ToFloat64(roomsCreated)

	IncActiveConnections(roomType)
	RecordMessageSent(roomType, messageType)
	RecordRoomCreated()

	afterConnections := testutil.ToFloat64(activeConnections.WithLabelValues(roomType))
	if afterConnections != beforeConnections+1 {
		t.Fatalf("expected active connections to increment by 1, before=%f after=%f", beforeConnections, afterConnections)
	}

	afterMessages := testutil.ToFloat64(messagesSent.WithLabelValues(roomType, messageType))
	if afterMessages != beforeMessages+1 {
		t.Fatalf("expected messages sent to increment by 1, before=%f after=%f", beforeMessages, afterMessages)
	}

	afterRooms := testutil.ToFloat64(roomsCreated)
	if afterRooms != beforeRooms+1 {
		t.Fatalf("expected rooms created to increment by 1, before=%f after=%f", beforeRooms, afterRooms)
	}

	DecActiveConnections(roomType)
	afterDisconnect := testutil.ToFloat64(activeConnections.WithLabelValues(roomType))
	if afterDisconnect != beforeConnections {
		t.Fatalf("expected active connections to return to %f, got %f", beforeConnections, afterDisconnect)
	}

	SetUsersOnline(42)
	if got := testutil.ToFloat64(usersOnline); got != 42 {
		t.Fatalf("expected users online gauge 42, got %f", got)
	}
}

func TestBusinessMetricsFallbackRoomTypeUsesAll(t *testing.T) {
	messageType := "system-" + safeMetricLabel(t.Name())
	beforeConnections := testutil.ToFloat64(activeConnections.WithLabelValues(RoomTypeAll))
	beforeMessages := testutil.ToFloat64(messagesSent.WithLabelValues(RoomTypeAll, messageType))

	IncActiveConnections("")
	RecordMessageSent("", messageType)

	afterConnections := testutil.ToFloat64(activeConnections.WithLabelValues(RoomTypeAll))
	if afterConnections != beforeConnections+1 {
		t.Fatalf("expected empty active connection room type to use %q, before=%f after=%f", RoomTypeAll, beforeConnections, afterConnections)
	}

	afterMessages := testutil.ToFloat64(messagesSent.WithLabelValues(RoomTypeAll, messageType))
	if afterMessages != beforeMessages+1 {
		t.Fatalf("expected empty message room type to use %q, before=%f after=%f", RoomTypeAll, beforeMessages, afterMessages)
	}

	DecActiveConnections("")
}

func safeMetricLabel(value string) string {
	value = strings.NewReplacer(
		"/", "_",
		" ", "_",
		"-", "_",
	).Replace(value)

	return strings.ToLower(value)
}
