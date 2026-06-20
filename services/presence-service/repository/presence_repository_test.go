package repository

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

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

func TestPresenceRepositorySetOnlineStoresPresenceAndRoomMembership(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.SetOnline(ctx, "user-1", "room-1"); err != nil {
		t.Fatalf("set online: %v", err)
	}

	status, err := repo.client.Get(ctx, userStatusKey("user-1")).Result()
	if err != nil {
		t.Fatalf("get status: %v", err)
	}

	if status != string(presenceservice.PresenceStatusOnline) {
		t.Fatalf("expected online status, got %q", status)
	}

	if exists := repo.client.Exists(ctx, userLastSeenKey("user-1")).Val(); exists != 1 {
		t.Fatalf("expected last seen key to exist, got %d", exists)
	}

	rooms, err := repo.client.SMembers(ctx, userRoomsKey("user-1")).Result()
	if err != nil {
		t.Fatalf("get user rooms: %v", err)
	}

	if len(rooms) != 1 || rooms[0] != "room-1" {
		t.Fatalf("expected room-1 membership, got %#v", rooms)
	}

	if _, err := repo.client.ZScore(ctx, roomOnlineKey("room-1"), "user-1").Result(); err != nil {
		t.Fatalf("expected room online score: %v", err)
	}
}

func TestPresenceRepositorySetOfflineForSingleRoomRemovesRoomState(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.SetOnline(ctx, "user-1", "room-1"); err != nil {
		t.Fatalf("set online: %v", err)
	}

	if err := repo.SetTyping(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("set typing: %v", err)
	}

	if err := repo.SetOffline(ctx, "user-1", "room-1"); err != nil {
		t.Fatalf("set offline: %v", err)
	}

	if _, err := repo.client.ZScore(ctx, roomOnlineKey("room-1"), "user-1").Result(); err != redis.Nil {
		t.Fatalf("expected user removed from room online set, got %v", err)
	}

	if exists := repo.client.Exists(ctx, roomTypingKey("room-1", "user-1")).Val(); exists != 0 {
		t.Fatalf("expected typing key deleted, got exists=%d", exists)
	}
}

func TestPresenceRepositorySetOfflineWithoutRoomRemovesTrackedRooms(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.SetOnline(ctx, "user-1", "room-1"); err != nil {
		t.Fatalf("set online room-1: %v", err)
	}

	if err := repo.SetOnline(ctx, "user-1", "room-2"); err != nil {
		t.Fatalf("set online room-2: %v", err)
	}

	if err := repo.SetTyping(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("set typing: %v", err)
	}

	if err := repo.SetOffline(ctx, "user-1", ""); err != nil {
		t.Fatalf("set offline globally: %v", err)
	}

	if exists := repo.client.Exists(ctx, userStatusKey("user-1")).Val(); exists != 0 {
		t.Fatalf("expected status key deleted, got exists=%d", exists)
	}

	if exists := repo.client.Exists(ctx, userRoomsKey("user-1")).Val(); exists != 0 {
		t.Fatalf("expected rooms key deleted, got exists=%d", exists)
	}

	for _, roomID := range []string{"room-1", "room-2"} {
		if _, err := repo.client.ZScore(ctx, roomOnlineKey(roomID), "user-1").Result(); err != redis.Nil {
			t.Fatalf("expected user removed from %s online set, got %v", roomID, err)
		}
	}
}

func TestPresenceRepositorySetTypingStoresTypingState(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.SetTyping(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("set typing: %v", err)
	}

	if exists := repo.client.Exists(ctx, roomTypingKey("room-1", "user-1")).Val(); exists != 1 {
		t.Fatalf("expected typing key, got exists=%d", exists)
	}

	if _, err := repo.client.ZScore(ctx, roomOnlineKey("room-1"), "user-1").Result(); err != nil {
		t.Fatalf("expected online score: %v", err)
	}

	rooms, err := repo.client.SMembers(ctx, userRoomsKey("user-1")).Result()
	if err != nil {
		t.Fatalf("get user rooms: %v", err)
	}

	if len(rooms) != 1 || rooms[0] != "room-1" {
		t.Fatalf("expected room-1 membership, got %#v", rooms)
	}
}

func TestPresenceRepositoryGetPresenceReturnsOnlineStatus(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.SetOnline(ctx, "user-1", "room-1"); err != nil {
		t.Fatalf("set online: %v", err)
	}

	presence, err := repo.GetPresence(ctx, "user-1")
	if err != nil {
		t.Fatalf("get presence: %v", err)
	}

	if presence.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", presence.UserID)
	}

	if presence.Status != presenceservice.PresenceStatusOnline {
		t.Fatalf("expected online, got %q", presence.Status)
	}

	if presence.LastSeen.IsZero() {
		t.Fatal("expected last seen")
	}
}

func TestPresenceRepositoryGetPresenceDefaultsToOffline(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	presence, err := repo.GetPresence(ctx, "user-1")
	if err != nil {
		t.Fatalf("get presence: %v", err)
	}

	if presence.Status != presenceservice.PresenceStatusOffline {
		t.Fatalf("expected offline, got %q", presence.Status)
	}
}

func TestPresenceRepositoryGetPresenceWrapsInvalidLastSeen(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.client.Set(ctx, userLastSeenKey("user-1"), "not-a-time", LastSeenTTL).Err(); err != nil {
		t.Fatalf("set invalid last seen: %v", err)
	}

	_, err := repo.GetPresence(ctx, "user-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "parse user last seen") {
		t.Fatalf("expected parse user last seen error, got %v", err)
	}
}

func TestPresenceRepositoryGetRoomPresenceReturnsOnlineAndTypingUsers(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.SetOnline(ctx, "user-1", "room-1"); err != nil {
		t.Fatalf("set online: %v", err)
	}

	if err := repo.SetTyping(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("set typing: %v", err)
	}

	presences, err := repo.GetRoomPresence(ctx, "room-1")
	if err != nil {
		t.Fatalf("get room presence: %v", err)
	}

	if len(presences) != 1 {
		t.Fatalf("expected 1 presence, got %#v", presences)
	}

	if presences[0].UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", presences[0].UserID)
	}

	if presences[0].RoomID != "room-1" {
		t.Fatalf("expected room-1, got %q", presences[0].RoomID)
	}

	if presences[0].Status != presenceservice.PresenceStatusTyping {
		t.Fatalf("expected typing, got %q", presences[0].Status)
	}
}

func TestPresenceRepositoryGetRoomPresenceRemovesExpiredOnlineMembers(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	oldScore := float64(time.Now().UTC().Add(-PresenceTTL - time.Second).UnixMilli())
	if err := repo.client.ZAdd(ctx, roomOnlineKey("room-1"), redis.Z{
		Score:  oldScore,
		Member: "user-1",
	}).Err(); err != nil {
		t.Fatalf("add expired online member: %v", err)
	}

	if err := repo.client.Set(ctx, userStatusKey("user-1"), string(presenceservice.PresenceStatusOnline), PresenceTTL).Err(); err != nil {
		t.Fatalf("set status: %v", err)
	}

	presences, err := repo.GetRoomPresence(ctx, "room-1")
	if err != nil {
		t.Fatalf("get room presence: %v", err)
	}

	if len(presences) != 0 {
		t.Fatalf("expected expired member removed, got %#v", presences)
	}

	if count := repo.client.ZCard(ctx, roomOnlineKey("room-1")).Val(); count != 0 {
		t.Fatalf("expected empty online set, got %d", count)
	}
}

func TestPresenceRepositoryGetTypingUsersScansKeys(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	if err := repo.SetTyping(ctx, "room-1", "user-1"); err != nil {
		t.Fatalf("set typing user-1: %v", err)
	}

	if err := repo.SetTyping(ctx, "room-1", "user-2"); err != nil {
		t.Fatalf("set typing user-2: %v", err)
	}

	users, err := repo.getTypingUsers(ctx, "room-1")
	if err != nil {
		t.Fatalf("get typing users: %v", err)
	}

	userSet := map[string]bool{}
	for _, userID := range users {
		userSet[userID] = true
	}

	if !userSet["user-1"] || !userSet["user-2"] {
		t.Fatalf("expected user-1 and user-2, got %#v", users)
	}
}

func TestPresenceRepositoryGetLastSeenReturnsZeroWhenMissing(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestPresenceRepository(t)

	got, err := repo.getLastSeen(ctx, "user-1")
	if err != nil {
		t.Fatalf("get missing last seen: %v", err)
	}

	if !got.IsZero() {
		t.Fatalf("expected zero last seen, got %s", got)
	}
}

func newTestPresenceRepository(t *testing.T) (*PresenceRepository, *miniredis.Miniredis) {
	t.Helper()

	redisServer := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close redis client: %v", err)
		}
	})

	repo, err := NewPresenceRepository(client)
	if err != nil {
		t.Fatalf("create presence repository: %v", err)
	}

	return repo, redisServer
}
