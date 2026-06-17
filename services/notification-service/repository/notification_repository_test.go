package repository

import (
	"strings"
	"testing"
)

func TestNewNotificationRepositoryRequiresPostgresPool(t *testing.T) {
	_, err := NewNotificationRepository(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateNotificationQueryReturnsNotificationFields(t *testing.T) {
	expectedParts := []string{
		"INSERT INTO notifications",
		"RETURNING id::text, user_id::text, type, title, body, COALESCE(room_id::text, ''), read, created_at",
	}

	for _, part := range expectedParts {
		if !strings.Contains(createNotificationQuery, part) {
			t.Fatalf("expected create query to contain %q, got query: %s", part, createNotificationQuery)
		}
	}
}

func TestListNotificationsByUserIDQueryUsesUnreadFilterAndLimit(t *testing.T) {
	expectedParts := []string{
		"WHERE user_id = $1",
		"AND ($2::boolean = false OR read = false)",
		"ORDER BY created_at DESC, id DESC",
		"LIMIT $3",
	}

	for _, part := range expectedParts {
		if !strings.Contains(listNotificationsByUserIDQuery, part) {
			t.Fatalf("expected list query to contain %q, got query: %s", part, listNotificationsByUserIDQuery)
		}
	}
}

func TestMarkNotificationAsReadQueryUpdatesReadStatusWithOwnershipCheck(t *testing.T) {
	expectedParts := []string{
		"UPDATE notifications",
		"SET read = true",
		"WHERE id = $1",
		"AND user_id = $2",
	}

	for _, part := range expectedParts {
		if !strings.Contains(markNotificationAsReadQuery, part) {
			t.Fatalf("expected mark as read query to contain %q, got query: %s", part, markNotificationAsReadQuery)
		}
	}
}
