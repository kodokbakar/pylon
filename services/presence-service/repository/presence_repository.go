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

func (r *PresenceRepository) SetOnline(ctx context.Context, userID string) error {
	now := time.Now().UTC()

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, userStatusKey(userID), string(presenceservice.PresenceStatusOnline), PresenceTTL)
	pipe.Set(ctx, userLastSeenKey(userID), now.Format(time.RFC3339Nano), LastSeenTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set user online presence: %w", err)
	}

	return nil
}

func (r *PresenceRepository) SetOffline(ctx context.Context, userID string) error {
	now := time.Now().UTC()

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, userStatusKey(userID))
	pipe.Set(ctx, userLastSeenKey(userID), now.Format(time.RFC3339Nano), LastSeenTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("set user offline presence: %w", err)
	}

	if err := r.removeUserFromRoomOnlineSets(ctx, userID); err != nil {
		return fmt.Errorf("remove user from room online sets: %w", err)
	}

	return nil
}

func (r *PresenceRepository) SetTyping(ctx context.Context, roomID, userID string) error {
	now := time.Now().UTC()

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, roomTypingKey(roomID, userID), "1", TypingTTL)
	pipe.ZAdd(ctx, roomOnlineKey(roomID), redis.Z{
		Score:  float64(now.Unix()),
		Member: userID,
	})
	pipe.Expire(ctx, roomOnlineKey(roomID), PresenceTTL)

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

	userIDs, err := r.client.ZRange(ctx, roomOnlineKey(roomID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("get room online users: %w", err)
	}

	presences, err := r.getPresences(ctx, roomID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get room presences: %w", err)
	}

	typingUsers, err := r.getTypingUsers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	for _, userID := range typingUsers {
		presences = upsertTypingPresence(presences, roomID, userID)
	}

	return presences, nil
}

func (r *PresenceRepository) getPresences(ctx context.Context, roomID string, userIDs []string) ([]presenceservice.Presence, error) {
	if len(userIDs) == 0 {
		return []presenceservice.Presence{}, nil
	}

	statusKeys := make([]string, 0, len(userIDs))
	lastSeenKeys := make([]string, 0, len(userIDs))

	for _, userID := range userIDs {
		statusKeys = append(statusKeys, userStatusKey(userID))
		lastSeenKeys = append(lastSeenKeys, userLastSeenKey(userID))
	}

	pipe := r.client.Pipeline()
	statusCmd := pipe.MGet(ctx, statusKeys...)
	lastSeenCmd := pipe.MGet(ctx, lastSeenKeys...)

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("batch get room presence values: %w", err)
	}

	statusValues, err := statusCmd.Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("batch get user statuses: %w", err)
	}

	lastSeenValues, err := lastSeenCmd.Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("batch get user last seen values: %w", err)
	}

	presences := make([]presenceservice.Presence, 0, len(userIDs))
	for i, userID := range userIDs {
		status := redisValueString(statusValues[i])
		if status == "" {
			status = string(presenceservice.PresenceStatusOffline)
		}

		lastSeen, err := parseLastSeenValue(lastSeenValues[i])
		if err != nil {
			return nil, fmt.Errorf("parse last seen for user %s: %w", userID, err)
		}

		presences = append(presences, presenceservice.Presence{
			UserID:   userID,
			RoomID:   roomID,
			Status:   presenceservice.PresenceStatus(status),
			LastSeen: lastSeen,
		})
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
	expiredBefore := strconv.FormatInt(time.Now().UTC().Add(-PresenceTTL).Unix(), 10)

	if err := r.client.ZRemRangeByScore(ctx, roomOnlineKey(roomID), "-inf", expiredBefore).Err(); err != nil {
		return fmt.Errorf("remove expired room online members: %w", err)
	}

	return nil
}

func (r *PresenceRepository) removeUserFromRoomOnlineSets(ctx context.Context, userID string) error {
	var cursor uint64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, "room:*:online", 100).Result()
		if err != nil {
			return fmt.Errorf("scan room online sets: %w", err)
		}

		for _, key := range keys {
			if err := r.client.ZRem(ctx, key, userID).Err(); err != nil {
				return fmt.Errorf("remove user from %s: %w", key, err)
			}
		}

		if nextCursor == 0 {
			return nil
		}

		cursor = nextCursor
	}
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

func upsertTypingPresence(presences []presenceservice.Presence, roomID, userID string) []presenceservice.Presence {
	for i := range presences {
		if presences[i].UserID == userID {
			presences[i].Status = presenceservice.PresenceStatusTyping
			return presences
		}
	}

	return append(presences, presenceservice.Presence{
		UserID: userID,
		RoomID: roomID,
		Status: presenceservice.PresenceStatusTyping,
	})
}

func userStatusKey(userID string) string {
	return fmt.Sprintf("user:%s:status", userID)
}

func userLastSeenKey(userID string) string {
	return fmt.Sprintf("user:%s:last_seen", userID)
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
