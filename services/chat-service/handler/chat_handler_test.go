package handler

import (
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	chatservice "github.com/kodokbakar/pylon/services/chat-service/service"
)

func TestDomainMessageToProtoMapsSenderFields(t *testing.T) {
	createdAt := time.Now()

	msg := domainMessageToProto(&chatservice.Message{
		ID:                "message-1",
		RoomID:            "room-1",
		SenderID:          "user-1",
		SenderUsername:    "alice",
		SenderDisplayName: "Alice Doe",
		SenderAvatarURL:   "https://example.com/avatar.png",
		Content:           "hello",
		Type:              chatservice.MessageTypeFile,
		CreatedAt:         createdAt,
	})

	if msg.GetId() != "message-1" {
		t.Fatalf("expected message-1, got %q", msg.GetId())
	}

	if msg.GetSenderUsername() != "alice" {
		t.Fatalf("expected sender username alice, got %q", msg.GetSenderUsername())
	}

	if msg.GetSenderDisplayName() != "Alice Doe" {
		t.Fatalf("expected sender display name Alice Doe, got %q", msg.GetSenderDisplayName())
	}

	if msg.GetSenderAvatarUrl() != "https://example.com/avatar.png" {
		t.Fatalf("expected sender avatar url, got %q", msg.GetSenderAvatarUrl())
	}

	if msg.GetType() != domainMessageTypeToProto(chatservice.MessageTypeFile) {
		t.Fatalf("expected file message type, got %v", msg.GetType())
	}

	if !msg.GetCreatedAt().AsTime().Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, msg.GetCreatedAt().AsTime())
	}
}

func TestConnectErrorMapsForbiddenToPermissionDenied(t *testing.T) {
	err := connectError(chatservice.ErrForbidden)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodePermissionDenied {
		t.Fatalf("expected permission denied, got %v", connectErr.Code())
	}
}

func TestConnectErrorMapsInvalidInputToInvalidArgument(t *testing.T) {
	err := connectError(chatservice.ErrInvalidInput)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid argument, got %v", connectErr.Code())
	}
}
