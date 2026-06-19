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
		"(user_id, type, title, body, room_id, message_id)",
		"NULLIF($6, '')::uuid",
		"ON CONFLICT (user_id, message_id) WHERE message_id IS NOT NULL",
		"DO UPDATE SET title = notifications.title",
		"RETURNING id::text, user_id::text, type, title, body, COALESCE(room_id::text, ''), COALESCE(message_id::text, ''), read, created_at",
	}

	for _, part := range expectedParts {
		if !strings.Contains(createNotificationQuery, part) {
			t.Fatalf("expected create query to contain %q, got query: %s", part, createNotificationQuery)
		}
	}
}

func TestListNotificationsByUserIDQueryUsesUnreadFilterAndLimit(t *testing.T) {
	expectedParts := []string{
		"FROM notifications",
		"WHERE user_id = $1",
		"AND ($2::boolean = false OR read = false)",
		"ORDER BY created_at DESC, id DESC",
		"LIMIT $3 OFFSET $4",
		"COALESCE(message_id::text, '')",
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

func TestCreateNotificationQueryIsIdempotentForMessageEvents(t *testing.T) {
	expectedParts := []string{
		"message_id",
		"ON CONFLICT (user_id, message_id) WHERE message_id IS NOT NULL",
		"DO UPDATE SET title = notifications.title",
	}

	for _, part := range expectedParts {
		if !strings.Contains(createNotificationQuery, part) {
			t.Fatalf("expected create query to contain %q, got query: %s", part, createNotificationQuery)
		}
	}
}
