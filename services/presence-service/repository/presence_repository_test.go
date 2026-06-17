package repository

import (
	"testing"
	"time"

	presenceservice "github.com/kodokbakar/pylon/services/presence-service/service"
)

func TestNewPresenceRepositoryRequiresRedisClient(t *testing.T) {
	_, err := NewPresenceRepository(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPresenceKeys(t *testing.T) {
	if got := userStatusKey("user-1"); got != "user:user-1:status" {
		t.Fatalf("expected user status key, got %q", got)
	}

	if got := userLastSeenKey("user-1"); got != "user:user-1:last_seen" {
		t.Fatalf("expected user last seen key, got %q", got)
	}

	if got := roomTypingKey("room-1", "user-1"); got != "room:room-1:typing:user-1" {
		t.Fatalf("expected room typing key, got %q", got)
	}

	if got := roomOnlineKey("room-1"); got != "room:room-1:online" {
		t.Fatalf("expected room online key, got %q", got)
	}
}

func TestPresenceTTLs(t *testing.T) {
	if PresenceTTL != 30*time.Second {
		t.Fatalf("expected presence ttl 30s, got %s", PresenceTTL)
	}

	if TypingTTL != 3*time.Second {
		t.Fatalf("expected typing ttl 3s, got %s", TypingTTL)
	}

	if LastSeenTTL != 24*time.Hour {
		t.Fatalf("expected last seen ttl 24h, got %s", LastSeenTTL)
	}
}

func TestTypingUserIDFromKey(t *testing.T) {
	got := typingUserIDFromKey("room-1", "room:room-1:typing:user-1")
	if got != "user-1" {
		t.Fatalf("expected user-1, got %q", got)
	}
}

func TestUpsertTypingPresenceUpdatesExistingUser(t *testing.T) {
	presences := []presenceservice.Presence{
		{
			UserID: "user-1",
			RoomID: "room-1",
			Status: presenceservice.PresenceStatusOnline,
		},
	}

	got := upsertTypingPresence(presences, "room-1", "user-1")

	if len(got) != 1 {
		t.Fatalf("expected 1 presence, got %d", len(got))
	}

	if got[0].Status != presenceservice.PresenceStatusTyping {
		t.Fatalf("expected typing status, got %q", got[0].Status)
	}
}

func TestUpsertTypingPresenceAddsMissingUser(t *testing.T) {
	got := upsertTypingPresence(nil, "room-1", "user-1")

	if len(got) != 1 {
		t.Fatalf("expected 1 presence, got %d", len(got))
	}

	if got[0].UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", got[0].UserID)
	}

	if got[0].Status != presenceservice.PresenceStatusTyping {
		t.Fatalf("expected typing status, got %q", got[0].Status)
	}
}

func TestRedisValueString(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "nil value",
			value: nil,
			want:  "",
		},
		{
			name:  "string value",
			value: "online",
			want:  "online",
		},
		{
			name:  "bytes value",
			value: []byte("offline"),
			want:  "offline",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redisValueString(tt.value); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestParseLastSeenValue(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	got, err := parseLastSeenValue(now.Format(time.RFC3339Nano))
	if err != nil {
		t.Fatalf("parse last seen value: %v", err)
	}

	if !got.Equal(now) {
		t.Fatalf("expected %s, got %s", now, got)
	}
}

func TestParseLastSeenValueReturnsZeroForMissingValue(t *testing.T) {
	got, err := parseLastSeenValue(nil)
	if err != nil {
		t.Fatalf("parse missing last seen value: %v", err)
	}

	if !got.IsZero() {
		t.Fatalf("expected zero time, got %s", got)
	}
}
