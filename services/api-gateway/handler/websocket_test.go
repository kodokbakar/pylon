package handler

import (
	"testing"

	chatv1 "github.com/kodokbakar/pylon/gen/pylon/chat/v1"
	gatewaymanager "github.com/kodokbakar/pylon/services/api-gateway/manager"
)

func TestNormalizeOriginPatternsTrimsAndAddsHostPatterns(t *testing.T) {
	got := normalizeOriginPatterns([]string{
		" http://localhost:3000 ",
		"",
		" http://localhost:5173 ",
	})

	expected := []string{
		"http://localhost:3000",
		"localhost:3000",
		"http://localhost:5173",
		"localhost:5173",
	}

	if len(got) != len(expected) {
		t.Fatalf("expected %d origin patterns, got %#v", len(expected), got)
	}

	for i, pattern := range expected {
		if got[i] != pattern {
			t.Fatalf("expected origin pattern %d to be %q, got %q", i, pattern, got[i])
		}
	}
}

func TestClientMessageTypeToProtoDefaultsToText(t *testing.T) {
	got, err := clientMessageTypeToProto("")
	if err != nil {
		t.Fatalf("convert message type: %v", err)
	}

	if got != chatv1.MessageType_MESSAGE_TYPE_TEXT {
		t.Fatalf("expected text type, got %v", got)
	}
}

func TestClientMessageTypeToProtoSupportsFile(t *testing.T) {
	got, err := clientMessageTypeToProto("file")
	if err != nil {
		t.Fatalf("convert message type: %v", err)
	}

	if got != chatv1.MessageType_MESSAGE_TYPE_FILE {
		t.Fatalf("expected file type, got %v", got)
	}
}

func TestClientMessageTypeToProtoRejectsUnsupportedType(t *testing.T) {
	_, err := clientMessageTypeToProto("video")
	if err == nil {
		t.Fatal("expected unsupported message type error")
	}
}

func TestTypingEnvelopeUsesConnectionUsername(t *testing.T) {
	got := typingEnvelope(&gatewaymanager.Connection{
		UserID:   "user-1",
		Username: "alice",
	}, "room-1")

	if got.Type != "typing" {
		t.Fatalf("expected typing type, got %q", got.Type)
	}

	if got.RoomID != "room-1" {
		t.Fatalf("expected room-1, got %q", got.RoomID)
	}

	if got.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", got.UserID)
	}

	if got.Username != "alice" {
		t.Fatalf("expected username alice, got %q", got.Username)
	}
}

func TestTypingEnvelopeDoesNotUseUserIDAsUsernameWhenUsernameMissing(t *testing.T) {
	got := typingEnvelope(&gatewaymanager.Connection{
		UserID: "user-1",
	}, "room-1")

	if got.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", got.UserID)
	}

	if got.Username != "" {
		t.Fatalf("expected empty username, got %q", got.Username)
	}
}
