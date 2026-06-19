package repository

import (
	"strconv"
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

	if got := userRoomsKey("user-1"); got != "user:user-1:rooms" {
		t.Fatalf("expected user rooms key, got %q", got)
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

func TestMarkTypingPresenceUpdatesExistingUser(t *testing.T) {
	presences := []presenceservice.Presence{
		{
			UserID: "user-1",
			RoomID: "room-1",
			Status: presenceservice.PresenceStatusOnline,
		},
	}

	got := markTypingPresence(presences, "user-1")

	if len(got) != 1 {
		t.Fatalf("expected 1 presence, got %d", len(got))
	}

	if got[0].Status != presenceservice.PresenceStatusTyping {
		t.Fatalf("expected typing status, got %q", got[0].Status)
	}
}

func TestMarkTypingPresenceIgnoresMissingUser(t *testing.T) {
	got := markTypingPresence(nil, "user-1")

	if len(got) != 0 {
		t.Fatalf("expected no presence for missing online user, got %d", len(got))
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

func TestExpiredRoomOnlineScoreUsesMilliseconds(t *testing.T) {
	now := time.Unix(100, 500_000_000).UTC()
	got := expiredRoomOnlineScore(now)

	want := strconv.FormatInt(now.Add(-PresenceTTL).UnixMilli(), 10)
	if got != want {
		t.Fatalf("expected expired score %q, got %q", want, got)
	}
}

func TestTypingUserIDFromKeyReturnsOriginalWhenPrefixDoesNotMatch(t *testing.T) {
	got := typingUserIDFromKey("room-1", "room:room-2:typing:user-1")
	if got != "room:room-2:typing:user-1" {
		t.Fatalf("expected original key when prefix does not match, got %q", got)
	}
}

func TestRedisValueStringFormatsFallbackTypes(t *testing.T) {
	got := redisValueString(123)
	if got != "123" {
		t.Fatalf("expected formatted integer value, got %q", got)
	}
}

func TestParseLastSeenValueRejectsInvalidTime(t *testing.T) {
	_, err := parseLastSeenValue("not-a-time")
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

func TestMarkTypingPresenceOnlyUpdatesMatchingUser(t *testing.T) {
	presences := []presenceservice.Presence{
		{
			UserID: "user-1",
			Status: presenceservice.PresenceStatusOnline,
		},
		{
			UserID: "user-2",
			Status: presenceservice.PresenceStatusOnline,
		},
	}

	got := markTypingPresence(presences, "user-2")

	if got[0].Status != presenceservice.PresenceStatusOnline {
		t.Fatalf("expected user-1 to remain online, got %q", got[0].Status)
	}

	if got[1].Status != presenceservice.PresenceStatusTyping {
		t.Fatalf("expected user-2 typing, got %q", got[1].Status)
	}
}
