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

func TestNewMessageCreatedEventIncludesVersion(t *testing.T) {
	createdAt := time.Now()

	event, err := NewMessageCreatedEvent(&chatservice.Message{
		ID:        "message-1",
		RoomID:    "room-1",
		SenderID:  "user-1",
		Content:   "hello",
		Type:      chatservice.MessageTypeText,
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create message event: %v", err)
	}

	if event.Version != MessageCreatedEventVersion {
		t.Fatalf("expected version %q, got %q", MessageCreatedEventVersion, event.Version)
	}

	if event.Type != "message.created" {
		t.Fatalf("expected message.created type, got %q", event.Type)
	}

	if event.MessageID != "message-1" {
		t.Fatalf("expected message id message-1, got %q", event.MessageID)
	}

	if !event.Timestamp.Equal(createdAt) {
		t.Fatalf("expected timestamp %s, got %s", createdAt, event.Timestamp)
	}
}

func TestNewMessageCreatedEventRequiresMessage(t *testing.T) {
	_, err := NewMessageCreatedEvent(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
