package repository

import (
	"strings"
	"testing"
)

func TestNewRoomRepositoryRequiresPostgresPool(t *testing.T) {
	_, err := NewRoomRepository(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRoomQueryReturnsRoomFields(t *testing.T) {
	expectedParts := []string{
		"INSERT INTO rooms",
		"RETURNING id, name, type, created_by, created_at",
	}

	for _, part := range expectedParts {
		if !strings.Contains(createRoomQuery, part) {
			t.Fatalf("expected create room query to contain %q, got query: %s", part, createRoomQuery)
		}
	}
}

func TestListRoomsByUserIDQueryUsesMembershipJoin(t *testing.T) {
	expectedParts := []string{
		"JOIN room_members rm ON rm.room_id = r.id",
		"WHERE rm.user_id = $1",
		"ORDER BY r.created_at DESC, r.id DESC",
	}

	for _, part := range expectedParts {
		if !strings.Contains(listRoomsByUserIDQuery, part) {
			t.Fatalf("expected list rooms query to contain %q, got query: %s", part, listRoomsByUserIDQuery)
		}
	}
}

func TestFindDirectRoomQueryUsesBothMembers(t *testing.T) {
	expectedParts := []string{
		"WHERE r.type = 'direct'",
		"rm1.user_id = $1",
		"rm2.user_id = $2",
	}

	for _, part := range expectedParts {
		if !strings.Contains(findDirectRoomQuery, part) {
			t.Fatalf("expected find direct room query to contain %q, got query: %s", part, findDirectRoomQuery)
		}
	}
}
