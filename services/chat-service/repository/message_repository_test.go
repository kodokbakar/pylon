package repository

import (
	"strings"
	"testing"
)

func TestNewMessageRepositoryRequiresPostgresPool(t *testing.T) {
	_, err := NewMessageRepository(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListMessagesByRoomUsesStableOrdering(t *testing.T) {
	if !strings.Contains(listMessagesByRoomQuery, "ORDER BY created_at DESC, id DESC") {
		t.Fatalf("expected stable ordering by created_at and id, got query: %s", listMessagesByRoomQuery)
	}
}

func TestListMessagesByRoomBeforeIDUsesCompositeCursor(t *testing.T) {
	expectedParts := []string{
		"WITH cursor_message AS",
		"m.created_at < c.created_at",
		"m.created_at = c.created_at AND m.id < c.id",
		"ORDER BY m.created_at DESC, m.id DESC",
	}

	for _, part := range expectedParts {
		if !strings.Contains(listMessagesByRoomBeforeIDQuery, part) {
			t.Fatalf("expected query to contain %q, got query: %s", part, listMessagesByRoomBeforeIDQuery)
		}
	}
}
