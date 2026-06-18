package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakePresenceRepository struct {
	setOnlineFunc       func(ctx context.Context, userID string) error
	setOfflineFunc      func(ctx context.Context, userID string) error
	setTypingFunc       func(ctx context.Context, roomID, userID string) error
	getPresenceFunc     func(ctx context.Context, userID string) (*Presence, error)
	getRoomPresenceFunc func(ctx context.Context, roomID string) ([]Presence, error)
}

func (r *fakePresenceRepository) SetOnline(ctx context.Context, userID string) error {
	if r.setOnlineFunc == nil {
		return errors.New("set online func is not configured")
	}

	return r.setOnlineFunc(ctx, userID)
}

func (r *fakePresenceRepository) SetOffline(ctx context.Context, userID string) error {
	if r.setOfflineFunc == nil {
		return errors.New("set offline func is not configured")
	}

	return r.setOfflineFunc(ctx, userID)
}

func (r *fakePresenceRepository) SetTyping(ctx context.Context, roomID, userID string) error {
	if r.setTypingFunc == nil {
		return errors.New("set typing func is not configured")
	}

	return r.setTypingFunc(ctx, roomID, userID)
}

func (r *fakePresenceRepository) GetPresence(ctx context.Context, userID string) (*Presence, error) {
	if r.getPresenceFunc == nil {
		return nil, errors.New("get presence func is not configured")
	}

	return r.getPresenceFunc(ctx, userID)
}

func (r *fakePresenceRepository) GetRoomPresence(ctx context.Context, roomID string) ([]Presence, error) {
	if r.getRoomPresenceFunc == nil {
		return nil, errors.New("get room presence func is not configured")
	}

	return r.getRoomPresenceFunc(ctx, roomID)
}

func TestNewPresenceServiceRequiresRepository(t *testing.T) {
	_, err := NewPresenceService(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSetOnlineValidatesUserID(t *testing.T) {
	svc, err := NewPresenceService(&fakePresenceRepository{})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	err = svc.SetOnline(context.Background(), SetOnlineInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSetOnlineCallsRepository(t *testing.T) {
	called := false

	svc, err := NewPresenceService(&fakePresenceRepository{
		setOnlineFunc: func(ctx context.Context, userID string) error {
			called = true

			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			return nil
		},
	})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	err = svc.SetOnline(context.Background(), SetOnlineInput{UserID: " user-1 "})
	if err != nil {
		t.Fatalf("set online: %v", err)
	}

	if !called {
		t.Fatal("expected repository to be called")
	}
}

func TestSetOfflineValidatesUserID(t *testing.T) {
	svc, err := NewPresenceService(&fakePresenceRepository{})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	err = svc.SetOffline(context.Background(), SetOfflineInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSetOfflineCallsRepository(t *testing.T) {
	called := false

	svc, err := NewPresenceService(&fakePresenceRepository{
		setOfflineFunc: func(ctx context.Context, userID string) error {
			called = true

			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			return nil
		},
	})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	err = svc.SetOffline(context.Background(), SetOfflineInput{UserID: " user-1 "})
	if err != nil {
		t.Fatalf("set offline: %v", err)
	}

	if !called {
		t.Fatal("expected repository to be called")
	}
}

func TestSetTypingValidatesRoomID(t *testing.T) {
	svc, err := NewPresenceService(&fakePresenceRepository{})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	err = svc.SetTyping(context.Background(), SetTypingInput{
		UserID: "user-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSetTypingRequiresOnlineUser(t *testing.T) {
	svc, err := NewPresenceService(&fakePresenceRepository{
		getPresenceFunc: func(ctx context.Context, userID string) (*Presence, error) {
			return &Presence{
				UserID: userID,
				Status: PresenceStatusOffline,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	err = svc.SetTyping(context.Background(), SetTypingInput{
		RoomID: "room-1",
		UserID: "user-1",
	})
	if !errors.Is(err, ErrUserOffline) {
		t.Fatalf("expected user offline error, got %v", err)
	}
}

func TestSetTypingCallsRepositoryWhenUserIsOnline(t *testing.T) {
	called := false

	svc, err := NewPresenceService(&fakePresenceRepository{
		getPresenceFunc: func(ctx context.Context, userID string) (*Presence, error) {
			return &Presence{
				UserID: userID,
				Status: PresenceStatusOnline,
			}, nil
		},
		setTypingFunc: func(ctx context.Context, roomID, userID string) error {
			called = true

			if roomID != "room-1" {
				t.Fatalf("expected room-1, got %q", roomID)
			}

			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			return nil
		},
	})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	err = svc.SetTyping(context.Background(), SetTypingInput{
		RoomID: " room-1 ",
		UserID: " user-1 ",
	})
	if err != nil {
		t.Fatalf("set typing: %v", err)
	}

	if !called {
		t.Fatal("expected repository set typing to be called")
	}
}

func TestGetPresenceReturnsRepositoryValue(t *testing.T) {
	now := time.Now()

	svc, err := NewPresenceService(&fakePresenceRepository{
		getPresenceFunc: func(ctx context.Context, userID string) (*Presence, error) {
			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			return &Presence{
				UserID:   userID,
				Status:   PresenceStatusOnline,
				LastSeen: now,
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	presence, err := svc.GetPresence(context.Background(), GetPresenceInput{UserID: " user-1 "})
	if err != nil {
		t.Fatalf("get presence: %v", err)
	}

	if presence.Status != PresenceStatusOnline {
		t.Fatalf("expected online status, got %q", presence.Status)
	}

	if !presence.LastSeen.Equal(now) {
		t.Fatalf("expected last seen %s, got %s", now, presence.LastSeen)
	}
}

func TestGetRoomPresenceValidatesRoomID(t *testing.T) {
	svc, err := NewPresenceService(&fakePresenceRepository{})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	_, err = svc.GetRoomPresence(context.Background(), GetRoomPresenceInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestGetRoomPresenceReturnsRepositoryValue(t *testing.T) {
	now := time.Now()

	svc, err := NewPresenceService(&fakePresenceRepository{
		getRoomPresenceFunc: func(ctx context.Context, roomID string) ([]Presence, error) {
			if roomID != "room-1" {
				t.Fatalf("expected room-1, got %q", roomID)
			}

			return []Presence{
				{
					UserID:   "user-1",
					RoomID:   roomID,
					Status:   PresenceStatusOnline,
					LastSeen: now,
				},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	presences, err := svc.GetRoomPresence(context.Background(), GetRoomPresenceInput{
		RoomID: " room-1 ",
	})
	if err != nil {
		t.Fatalf("get room presence: %v", err)
	}

	if len(presences) != 1 {
		t.Fatalf("expected 1 presence, got %d", len(presences))
	}

	if presences[0].UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", presences[0].UserID)
	}

	if presences[0].Status != PresenceStatusOnline {
		t.Fatalf("expected online status, got %q", presences[0].Status)
	}
}

func TestStreamPresenceValidatesRoomID(t *testing.T) {
	svc, err := NewPresenceService(&fakePresenceRepository{})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	_, err = svc.StreamPresence(context.Background(), StreamPresenceInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestStreamPresenceReturnsChannelAndClosesWhenContextIsCanceled(t *testing.T) {
	svc, err := NewPresenceService(&fakePresenceRepository{})
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	events, err := svc.StreamPresence(ctx, StreamPresenceInput{
		RoomID: "room-1",
	})
	if err != nil {
		t.Fatalf("stream presence: %v", err)
	}

	if events == nil {
		t.Fatal("expected events channel, got nil")
	}

	cancel()

	select {
	case _, ok := <-events:
		if ok {
			t.Fatal("expected events channel to be closed")
		}
	case <-time.After(time.Second):
		t.Fatal("expected events channel to close after context cancellation")
	}
}
