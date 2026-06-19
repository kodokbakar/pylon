package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"

	chatservice "github.com/kodokbakar/pylon/services/chat-service/service"
)

type fakeKafkaWriter struct {
	messages []kafka.Message
	writeErr error
	closeErr error
	closed   bool
}

func (w *fakeKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if w.writeErr != nil {
		return w.writeErr
	}

	w.messages = append(w.messages, msgs...)
	return nil
}

func (w *fakeKafkaWriter) Close() error {
	w.closed = true
	return w.closeErr
}

func TestNewProducerRequiresBrokers(t *testing.T) {
	_, err := NewProducer(nil, MessageEventsTopic, "test-client")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewProducerRequiresTopic(t *testing.T) {
	_, err := NewProducer([]string{"localhost:9092"}, "", "test-client")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewProducerConfigMatchesContract(t *testing.T) {
	producer, err := NewProducer([]string{"localhost:9092"}, MessageEventsTopic, "test-client")
	if err != nil {
		t.Fatalf("create producer: %v", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			t.Fatalf("close producer: %v", err)
		}
	}()

	writer, ok := producer.writer.(*kafka.Writer)
	if !ok {
		t.Fatalf("expected kafka writer, got %T", producer.writer)
	}

	if writer.Topic != MessageEventsTopic {
		t.Fatalf("expected topic %q, got %q", MessageEventsTopic, writer.Topic)
	}

	if writer.RequiredAcks != kafka.RequireAll {
		t.Fatalf("expected required acks all, got %v", writer.RequiredAcks)
	}

	if writer.BatchTimeout != defaultBatchTimeout {
		t.Fatalf("expected batch timeout %s, got %s", defaultBatchTimeout, writer.BatchTimeout)
	}

	if writer.BatchSize != defaultProducerBatchSize {
		t.Fatalf("expected batch size %d, got %d", defaultProducerBatchSize, writer.BatchSize)
	}

	if writer.Async {
		t.Fatal("expected synchronous producer")
	}

	if writer.Balancer == nil {
		t.Fatal("expected balancer to be configured")
	}
}

func TestNewMessageCreatedEventMatchesContract(t *testing.T) {
	createdAt := time.Now().UTC()

	event, err := NewMessageCreatedEvent(&chatservice.Message{
		ID:             "message-1",
		RoomID:         "room-1",
		SenderID:       "user-1",
		SenderUsername: "alice",
		Content:        "hello",
		Type:           chatservice.MessageTypeText,
		CreatedAt:      createdAt,
	})
	if err != nil {
		t.Fatalf("create message event: %v", err)
	}

	if event.EventID == "" {
		t.Fatal("expected event id to be generated")
	}

	if event.EventType != MessageCreatedEventType {
		t.Fatalf("expected message_created event type, got %q", event.EventType)
	}

	if event.Timestamp.IsZero() {
		t.Fatal("expected event timestamp")
	}

	if event.Data.MessageID != "message-1" {
		t.Fatalf("expected message id message-1, got %q", event.Data.MessageID)
	}

	if event.Data.RoomID != "room-1" {
		t.Fatalf("expected room id room-1, got %q", event.Data.RoomID)
	}

	if event.Data.SenderID != "user-1" {
		t.Fatalf("expected sender id user-1, got %q", event.Data.SenderID)
	}

	if event.Data.SenderUsername != "alice" {
		t.Fatalf("expected sender username alice, got %q", event.Data.SenderUsername)
	}

	if event.Data.Type != "text" {
		t.Fatalf("expected text message type, got %q", event.Data.Type)
	}

	if !event.Data.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, event.Data.CreatedAt)
	}
}

func TestNewMessageCreatedEventRequiresMessage(t *testing.T) {
	_, err := NewMessageCreatedEvent(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPublishMessageCreatedWritesKafkaMessage(t *testing.T) {
	writer := &fakeKafkaWriter{}
	producer := &Producer{writer: writer}

	createdAt := time.Now().UTC()
	err := producer.PublishMessageCreated(context.Background(), &chatservice.Message{
		ID:             "message-1",
		RoomID:         "room-1",
		SenderID:       "user-1",
		SenderUsername: "alice",
		Content:        "hello",
		Type:           chatservice.MessageTypeText,
		CreatedAt:      createdAt,
	})
	if err != nil {
		t.Fatalf("publish message created: %v", err)
	}

	if len(writer.messages) != 1 {
		t.Fatalf("expected 1 kafka message, got %d", len(writer.messages))
	}

	msg := writer.messages[0]
	if string(msg.Key) != "room-1" {
		t.Fatalf("expected key room-1, got %q", string(msg.Key))
	}

	if !hasHeader(msg.Headers, "event_type", MessageCreatedEventType) {
		t.Fatalf("expected event_type header %q, got %+v", MessageCreatedEventType, msg.Headers)
	}

	var event MessageCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		t.Fatalf("unmarshal event payload: %v", err)
	}

	if event.EventID == "" {
		t.Fatal("expected event id")
	}

	if event.EventType != MessageCreatedEventType {
		t.Fatalf("expected event type %q, got %q", MessageCreatedEventType, event.EventType)
	}

	if event.Data.MessageID != "message-1" {
		t.Fatalf("expected message-1, got %q", event.Data.MessageID)
	}

	if event.Data.RoomID != "room-1" {
		t.Fatalf("expected room-1, got %q", event.Data.RoomID)
	}

	if !event.Data.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, event.Data.CreatedAt)
	}
}

func TestPublishMessageCreatedReturnsWriteError(t *testing.T) {
	writeErr := errors.New("write failed")
	producer := &Producer{
		writer: &fakeKafkaWriter{writeErr: writeErr},
	}

	err := producer.PublishMessageCreated(context.Background(), &chatservice.Message{
		ID:             "message-1",
		RoomID:         "room-1",
		SenderID:       "user-1",
		SenderUsername: "alice",
		Content:        "hello",
		Type:           chatservice.MessageTypeText,
		CreatedAt:      time.Now(),
	})
	if !errors.Is(err, writeErr) {
		t.Fatalf("expected write error, got %v", err)
	}
}

func TestPublishMessageCreatedRequiresWriter(t *testing.T) {
	err := (*Producer)(nil).PublishMessageCreated(context.Background(), &chatservice.Message{
		ID:        "message-1",
		RoomID:    "room-1",
		SenderID:  "user-1",
		Content:   "hello",
		Type:      chatservice.MessageTypeText,
		CreatedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCloseClosesWriter(t *testing.T) {
	writer := &fakeKafkaWriter{}
	producer := &Producer{writer: writer}

	if err := producer.Close(); err != nil {
		t.Fatalf("close producer: %v", err)
	}

	if !writer.closed {
		t.Fatal("expected writer to be closed")
	}
}

func TestCloseReturnsWriterError(t *testing.T) {
	closeErr := errors.New("close failed")
	producer := &Producer{
		writer: &fakeKafkaWriter{closeErr: closeErr},
	}

	err := producer.Close()
	if !errors.Is(err, closeErr) {
		t.Fatalf("expected close error, got %v", err)
	}
}

func hasHeader(headers []kafka.Header, key, value string) bool {
	for _, header := range headers {
		if header.Key == key && string(header.Value) == value {
			return true
		}
	}

	return false
}
