package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New("invalid input")

var ErrUserOffline = errors.New("user is offline")

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
}

type SetOfflineInput struct {
	UserID string
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
	SetOnline(ctx context.Context, userID string) error
	SetOffline(ctx context.Context, userID string) error
	SetTyping(ctx context.Context, roomID, userID string) error
	GetPresence(ctx context.Context, userID string) (*Presence, error)
	GetRoomPresence(ctx context.Context, roomID string) ([]Presence, error)
}

type PresenceService struct {
	repo PresenceRepository
}

func NewPresenceService(repo PresenceRepository) (*PresenceService, error) {
	if repo == nil {
		return nil, fmt.Errorf("presence repository is required")
	}

	return &PresenceService{repo: repo}, nil
}

func (s *PresenceService) SetOnline(ctx context.Context, input SetOnlineInput) error {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	if err := s.repo.SetOnline(ctx, userID); err != nil {
		return fmt.Errorf("set user online: %w", err)
	}

	return nil
}

func (s *PresenceService) SetOffline(ctx context.Context, input SetOfflineInput) error {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	if err := s.repo.SetOffline(ctx, userID); err != nil {
		return fmt.Errorf("set user offline: %w", err)
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

	events := make(chan PresenceEvent)

	go func() {
		defer close(events)
		<-ctx.Done()
	}()

	return events, nil
}
