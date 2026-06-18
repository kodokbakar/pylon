package kafka

import (
	"testing"
	"time"

	chatservice "github.com/kodokbakar/pylon/services/chat-service/service"
)

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

func TestNewMessageCreatedEventMatchesContract(t *testing.T) {
	createdAt := time.Now()

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

	if event.EventType != "message_created" {
		t.Fatalf("expected message_created event type, got %q", event.EventType)
	}

	if event.MessageID != "message-1" {
		t.Fatalf("expected message id message-1, got %q", event.MessageID)
	}

	if event.RoomID != "room-1" {
		t.Fatalf("expected room id room-1, got %q", event.RoomID)
	}

	if event.SenderID != "user-1" {
		t.Fatalf("expected sender id user-1, got %q", event.SenderID)
	}

	if event.SenderUsername != "alice" {
		t.Fatalf("expected sender username alice, got %q", event.SenderUsername)
	}

	if event.Type != "text" {
		t.Fatalf("expected text message type, got %q", event.Type)
	}

	if !event.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, event.CreatedAt)
	}
}

func TestNewMessageCreatedEventRequiresMessage(t *testing.T) {
	_, err := NewMessageCreatedEvent(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
