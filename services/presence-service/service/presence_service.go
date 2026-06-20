package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrInvalidInput = errors.New("invalid input")

var ErrUserOffline = errors.New("user is offline")

const DefaultPresenceEventBuffer = 64

type PresenceStatus string

const (
	PresenceStatusOnline  PresenceStatus = "online"
	PresenceStatusOffline PresenceStatus = "offline"
	PresenceStatusTyping  PresenceStatus = "typing"
)

type Presence struct {
	UserID   string
	RoomID   string
	Status   PresenceStatus
	LastSeen time.Time
}

type PresenceEvent struct {
	UserID    string
	RoomID    string
	Status    PresenceStatus
	Timestamp time.Time
}

type SetOnlineInput struct {
	UserID string
	RoomID string
}

type SetOfflineInput struct {
	UserID string
	RoomID string
}

type SetTypingInput struct {
	UserID string
	RoomID string
}

type GetPresenceInput struct {
	UserID string
}

type GetRoomPresenceInput struct {
	RoomID string
}

type StreamPresenceInput struct {
	RoomID string
}

type PresenceRepository interface {
	SetOnline(ctx context.Context, userID, roomID string) error
	SetOffline(ctx context.Context, userID, roomID string) error
	SetTyping(ctx context.Context, roomID, userID string) error
	GetPresence(ctx context.Context, userID string) (*Presence, error)
	GetRoomPresence(ctx context.Context, roomID string) ([]Presence, error)
}

// PresenceEventBroker is intentionally in-memory for the current single-instance
// deployment. TODO: replace this with Redis Pub/Sub using channel
// presence:room:{room_id} before running multiple presence-service instances.
type PresenceEventBroker struct {
	mu          sync.RWMutex
	buffer      int
	subscribers map[string]map[chan PresenceEvent]struct{}
}

type PresenceService struct {
	repo   PresenceRepository
	broker *PresenceEventBroker
}

func NewPresenceService(repo PresenceRepository) (*PresenceService, error) {
	if repo == nil {
		return nil, fmt.Errorf("presence repository is required")
	}

	return &PresenceService{
		repo:   repo,
		broker: NewPresenceEventBroker(DefaultPresenceEventBuffer),
	}, nil
}

func NewPresenceEventBroker(buffer int) *PresenceEventBroker {
	if buffer <= 0 {
		buffer = DefaultPresenceEventBuffer
	}

	return &PresenceEventBroker{
		buffer:      buffer,
		subscribers: make(map[string]map[chan PresenceEvent]struct{}),
	}
}

func (b *PresenceEventBroker) Subscribe(ctx context.Context, roomID string) (<-chan PresenceEvent, error) {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	ch := make(chan PresenceEvent, b.buffer)

	b.mu.Lock()
	if b.subscribers[roomID] == nil {
		b.subscribers[roomID] = make(map[chan PresenceEvent]struct{})
	}
	b.subscribers[roomID][ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		delete(b.subscribers[roomID], ch)
		if len(b.subscribers[roomID]) == 0 {
			delete(b.subscribers, roomID)
		}
		close(ch)
		b.mu.Unlock()
	}()

	return ch, nil
}

func (b *PresenceEventBroker) Publish(roomID string, event PresenceEvent) int {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return 0
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	dropped := 0
	for ch := range b.subscribers[roomID] {
		select {
		case ch <- event:
		default:
			dropped++
		}
	}

	return dropped
}

func (s *PresenceService) SetOnline(ctx context.Context, input SetOnlineInput) error {
	userID := strings.TrimSpace(input.UserID)
	roomID := strings.TrimSpace(input.RoomID)

	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	if err := s.repo.SetOnline(ctx, userID, roomID); err != nil {
		return fmt.Errorf("set user online: %w", err)
	}

	if roomID != "" {
		s.broker.Publish(roomID, PresenceEvent{
			UserID:    userID,
			RoomID:    roomID,
			Status:    PresenceStatusOnline,
			Timestamp: time.Now().UTC(),
		})
	}

	return nil
}

func (s *PresenceService) SetOffline(ctx context.Context, input SetOfflineInput) error {
	userID := strings.TrimSpace(input.UserID)
	roomID := strings.TrimSpace(input.RoomID)

	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	if err := s.repo.SetOffline(ctx, userID, roomID); err != nil {
		return fmt.Errorf("set user offline: %w", err)
	}

	if roomID != "" {
		s.broker.Publish(roomID, PresenceEvent{
			UserID:    userID,
			RoomID:    roomID,
			Status:    PresenceStatusOffline,
			Timestamp: time.Now().UTC(),
		})
	}

	return nil
}

func (s *PresenceService) SetTyping(ctx context.Context, input SetTypingInput) error {
	roomID := strings.TrimSpace(input.RoomID)
	userID := strings.TrimSpace(input.UserID)

	if roomID == "" {
		return fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	presence, err := s.repo.GetPresence(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user presence before typing: %w", err)
	}

	if presence == nil || presence.Status != PresenceStatusOnline {
		return fmt.Errorf("%w: user must be online before typing", ErrUserOffline)
	}

	if err := s.repo.SetTyping(ctx, roomID, userID); err != nil {
		return fmt.Errorf("set user typing: %w", err)
	}

	s.broker.Publish(roomID, PresenceEvent{
		UserID:    userID,
		RoomID:    roomID,
		Status:    PresenceStatusTyping,
		Timestamp: time.Now().UTC(),
	})

	return nil
}

func (s *PresenceService) GetPresence(ctx context.Context, input GetPresenceInput) (*Presence, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	presence, err := s.repo.GetPresence(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get presence: %w", err)
	}

	return presence, nil
}

func (s *PresenceService) GetRoomPresence(ctx context.Context, input GetRoomPresenceInput) ([]Presence, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if roomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	presences, err := s.repo.GetRoomPresence(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room presence: %w", err)
	}

	return presences, nil
}

func (s *PresenceService) StreamPresence(ctx context.Context, input StreamPresenceInput) (<-chan PresenceEvent, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if roomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	events, err := s.broker.Subscribe(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("subscribe presence events: %w", err)
	}

	return events, nil
}
