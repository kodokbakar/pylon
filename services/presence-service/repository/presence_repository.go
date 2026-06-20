package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	presenceservice "github.com/kodokbakar/pylon/services/presence-service/service"
)

const (
	PresenceTTL = 30 * time.Second
	TypingTTL   = 3 * time.Second
	LastSeenTTL = 24 * time.Hour
)

type PresenceRepository struct {
	client *redis.Client
}

func NewPresenceRepository(client *redis.Client) (*PresenceRepository, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	return &PresenceRepository{client: client}, nil
}

// SetOnline also acts as heartbeat renewal.
// If roomID is provided, it also refreshes the user's online score in that room.
func (r *PresenceRepository) SetOnline(ctx context.Context, userID, roomID string) error {
	now := time.Now().UTC()

	pipe := r.client.Pipeline()
	pipe.Set(ctx, userStatusKey(userID), string(presenceservice.PresenceStatusOnline), PresenceTTL)
	pipe.Set(ctx, userLastSeenKey(userID), now.Format(time.RFC3339Nano), LastSeenTTL)

	roomID = strings.TrimSpace(roomID)
	if roomID != "" {
		pipe.ZAdd(ctx, roomOnlineKey(roomID), redis.Z{
			Score:  float64(now.UnixMilli()),
			Member: userID,
		})
		pipe.SAdd(ctx, userRoomsKey(userID), roomID)
		pipe.Expire(ctx, userRoomsKey(userID), LastSeenTTL)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set user online presence: %w", err)
	}

	return nil
}

// SetOffline removes a user from a single room when roomID is provided.
// When roomID is empty, the user is marked globally offline and removed from
// tracked rooms without scanning all Redis keys.
func (r *PresenceRepository) SetOffline(ctx context.Context, userID, roomID string) error {
	now := time.Now().UTC()
	roomID = strings.TrimSpace(roomID)

	rooms := []string{roomID}
	if roomID == "" {
		trackedRooms, err := r.client.SMembers(ctx, userRoomsKey(userID)).Result()
		if err != nil {
			return fmt.Errorf("get user tracked rooms: %w", err)
		}

		rooms = trackedRooms
	}

	pipe := r.client.Pipeline()
	pipe.Set(ctx, userLastSeenKey(userID), now.Format(time.RFC3339Nano), LastSeenTTL)

	if roomID == "" {
		pipe.Del(ctx, userStatusKey(userID), userRoomsKey(userID))
	} else {
		pipe.SRem(ctx, userRoomsKey(userID), roomID)
	}

	for _, trackedRoomID := range rooms {
		trackedRoomID = strings.TrimSpace(trackedRoomID)
		if trackedRoomID == "" {
			continue
		}

		pipe.ZRem(ctx, roomOnlineKey(trackedRoomID), userID)
		pipe.Del(ctx, roomTypingKey(trackedRoomID, userID))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set user offline presence: %w", err)
	}

	return nil
}

func (r *PresenceRepository) SetTyping(ctx context.Context, roomID, userID string) error {
	now := time.Now().UTC()

	pipe := r.client.Pipeline()
	pipe.Set(ctx, roomTypingKey(roomID, userID), "1", TypingTTL)
	pipe.ZAdd(ctx, roomOnlineKey(roomID), redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: userID,
	})
	pipe.SAdd(ctx, userRoomsKey(userID), roomID)
	pipe.Expire(ctx, userRoomsKey(userID), LastSeenTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set user typing presence: %w", err)
	}

	return nil
}

func (r *PresenceRepository) GetPresence(ctx context.Context, userID string) (*presenceservice.Presence, error) {
	status, statusErr := r.client.Get(ctx, userStatusKey(userID)).Result()
	if statusErr != nil && statusErr != redis.Nil {
		return nil, fmt.Errorf("get user status: %w", statusErr)
	}

	lastSeen, err := r.getLastSeen(ctx, userID)
	if err != nil {
		return nil, err
	}

	if statusErr == redis.Nil || status == "" {
		status = string(presenceservice.PresenceStatusOffline)
	}

	return &presenceservice.Presence{
		UserID:   userID,
		Status:   presenceservice.PresenceStatus(status),
		LastSeen: lastSeen,
	}, nil
}

func (r *PresenceRepository) GetRoomPresence(ctx context.Context, roomID string) ([]presenceservice.Presence, error) {
	if err := r.removeExpiredRoomOnlineMembers(ctx, roomID); err != nil {
		return nil, err
	}

	members, err := r.client.ZRangeWithScores(ctx, roomOnlineKey(roomID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("get room online users: %w", err)
	}

	presences := make([]presenceservice.Presence, 0, len(members))
	onlineUsers := make(map[string]struct{}, len(members))

	for _, member := range members {
		userID := redisValueString(member.Member)
		if userID == "" {
			continue
		}

		presence, err := r.GetPresence(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("get room user presence for %s: %w", userID, err)
		}

		if presence.Status != presenceservice.PresenceStatusOnline {
			continue
		}

		presence.RoomID = roomID
		if presence.LastSeen.IsZero() {
			presence.LastSeen = time.UnixMilli(int64(member.Score)).UTC()
		}

		onlineUsers[userID] = struct{}{}
		presences = append(presences, *presence)
	}

	typingUsers, err := r.getTypingUsers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	for _, userID := range typingUsers {
		if _, ok := onlineUsers[userID]; !ok {
			continue
		}

		presences = markTypingPresence(presences, userID)
	}

	return presences, nil
}

func (r *PresenceRepository) getLastSeen(ctx context.Context, userID string) (time.Time, error) {
	value, err := r.client.Get(ctx, userLastSeenKey(userID)).Result()
	if err == redis.Nil {
		return time.Time{}, nil
	}

	if err != nil {
		return time.Time{}, fmt.Errorf("get user last seen: %w", err)
	}

	lastSeen, err := parseLastSeenValue(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse user last seen: %w", err)
	}

	return lastSeen, nil
}

func redisValueString(value any) string {
	switch typedValue := value.(type) {
	case nil:
		return ""
	case string:
		return typedValue
	case []byte:
		return string(typedValue)
	default:
		return fmt.Sprint(typedValue)
	}
}

func parseLastSeenValue(value any) (time.Time, error) {
	lastSeenText := redisValueString(value)
	if lastSeenText == "" {
		return time.Time{}, nil
	}

	lastSeen, err := time.Parse(time.RFC3339Nano, lastSeenText)
	if err != nil {
		return time.Time{}, err
	}

	return lastSeen, nil
}

func (r *PresenceRepository) removeExpiredRoomOnlineMembers(ctx context.Context, roomID string) error {
	expiredBefore := expiredRoomOnlineScore(time.Now().UTC())

	if err := r.client.ZRemRangeByScore(ctx, roomOnlineKey(roomID), "-inf", expiredBefore).Err(); err != nil {
		return fmt.Errorf("remove expired room online members: %w", err)
	}

	return nil
}

func expiredRoomOnlineScore(now time.Time) string {
	return strconv.FormatInt(now.Add(-PresenceTTL).UnixMilli(), 10)
}

func (r *PresenceRepository) getTypingUsers(ctx context.Context, roomID string) ([]string, error) {
	pattern := roomTypingKey(roomID, "*")
	users := make([]string, 0)
	var cursor uint64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("scan room typing keys: %w", err)
		}

		for _, key := range keys {
			userID := typingUserIDFromKey(roomID, key)
			if userID != "" {
				users = append(users, userID)
			}
		}

		if nextCursor == 0 {
			return users, nil
		}

		cursor = nextCursor
	}
}

func markTypingPresence(presences []presenceservice.Presence, userID string) []presenceservice.Presence {
	for i := range presences {
		if presences[i].UserID == userID {
			presences[i].Status = presenceservice.PresenceStatusTyping
			return presences
		}
	}

	return presences
}

func userStatusKey(userID string) string {
	return fmt.Sprintf("user:%s:status", userID)
}

func userLastSeenKey(userID string) string {
	return fmt.Sprintf("user:%s:last_seen", userID)
}

func userRoomsKey(userID string) string {
	return fmt.Sprintf("user:%s:rooms", userID)
}

func roomTypingKey(roomID, userID string) string {
	return fmt.Sprintf("room:%s:typing:%s", roomID, userID)
}

func roomOnlineKey(roomID string) string {
	return fmt.Sprintf("room:%s:online", roomID)
}

func typingUserIDFromKey(roomID, key string) string {
	prefix := roomTypingKey(roomID, "")
	return strings.TrimPrefix(key, prefix)
}
