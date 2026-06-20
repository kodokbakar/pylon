package repository

import (
	"strings"
	"testing"
)

func TestNewMemberRepositoryRequiresPostgresPool(t *testing.T) {
	_, err := NewMemberRepository(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAddRoomMemberQueryUsesDoNothingOnConflict(t *testing.T) {
	expectedParts := []string{
		"INSERT INTO room_members",
		"ON CONFLICT (room_id, user_id) DO NOTHING",
	}

	for _, part := range expectedParts {
		if !strings.Contains(addRoomMemberQuery, part) {
			t.Fatalf("expected add room member query to contain %q, got query: %s", part, addRoomMemberQuery)
		}
	}
}

func TestListRoomMembersQueryOrdersMembers(t *testing.T) {
	expectedParts := []string{
		"SELECT",
		"rm.room_id",
		"rm.user_id",
		"rm.role",
		"rm.joined_at",
		"FROM room_members rm",
		"WHERE rm.room_id = $1",
		"ORDER BY rm.joined_at ASC, rm.user_id ASC",
	}

	for _, part := range expectedParts {
		if !strings.Contains(listRoomMembersQuery, part) {
			t.Fatalf("expected list room members query to contain %q, got query: %s", part, listRoomMembersQuery)
		}
	}
}

func TestGetRoomMemberRoleQuerySelectsRole(t *testing.T) {
	expectedParts := []string{
		"SELECT role",
		"WHERE room_id = $1",
		"AND user_id = $2",
	}

	for _, part := range expectedParts {
		if !strings.Contains(getRoomMemberRoleQuery, part) {
			t.Fatalf("expected get role query to contain %q, got query: %s", part, getRoomMemberRoleQuery)
		}
	}
}

func TestListRoomMembersQueryJoinsUsers(t *testing.T) {
	expectedParts := []string{
		"FROM room_members rm",
		"JOIN users u ON u.id = rm.user_id",
		"u.username",
		"COALESCE(u.display_name, '')",
		"COALESCE(u.avatar_url, '')",
		"ORDER BY rm.joined_at ASC, rm.user_id ASC",
	}

	for _, part := range expectedParts {
		if !strings.Contains(listRoomMembersQuery, part) {
			t.Fatalf("expected list room members query to contain %q, got query: %s", part, listRoomMembersQuery)
		}
	}
}
