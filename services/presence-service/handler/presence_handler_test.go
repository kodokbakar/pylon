package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	presencev1 "github.com/kodokbakar/pylon/gen/pylon/presence/v1"
	presenceservice "github.com/kodokbakar/pylon/services/presence-service/service"
)

func TestConnectErrorMapsInvalidInput(t *testing.T) {
	err := connectError(presenceservice.ErrInvalidInput)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid argument, got %v", connectErr.Code())
	}
}

func TestConnectErrorMapsUserOffline(t *testing.T) {
	err := connectError(presenceservice.ErrUserOffline)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeFailedPrecondition {
		t.Fatalf("expected failed precondition, got %v", connectErr.Code())
	}
}

func TestConnectErrorMapsUnknownError(t *testing.T) {
	err := connectError(errors.New("boom"))

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInternal {
		t.Fatalf("expected internal, got %v", connectErr.Code())
	}
}

func TestDomainStatusToProto(t *testing.T) {
	tests := []struct {
		name   string
		status presenceservice.PresenceStatus
		want   presencev1.PresenceStatus
	}{
		{
			name:   "online",
			status: presenceservice.PresenceStatusOnline,
			want:   presencev1.PresenceStatus_PRESENCE_STATUS_ONLINE,
		},
		{
			name:   "typing",
			status: presenceservice.PresenceStatusTyping,
			want:   presencev1.PresenceStatus_PRESENCE_STATUS_TYPING,
		},
		{
			name:   "offline",
			status: presenceservice.PresenceStatusOffline,
			want:   presencev1.PresenceStatus_PRESENCE_STATUS_OFFLINE,
		},
		{
			name:   "unknown defaults to offline",
			status: presenceservice.PresenceStatus("unknown"),
			want:   presencev1.PresenceStatus_PRESENCE_STATUS_OFFLINE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domainStatusToProto(tt.status); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestDomainEventToProto(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	got := domainEventToProto(presenceservice.PresenceEvent{
		UserID:    "user-1",
		RoomID:    "room-1",
		Status:    presenceservice.PresenceStatusTyping,
		Timestamp: now,
	})

	if got.GetUserId() != "user-1" {
		t.Fatalf("expected user-1, got %q", got.GetUserId())
	}

	if got.GetRoomId() != "room-1" {
		t.Fatalf("expected room-1, got %q", got.GetRoomId())
	}

	if got.GetStatus() != presencev1.PresenceStatus_PRESENCE_STATUS_TYPING {
		t.Fatalf("expected typing status, got %v", got.GetStatus())
	}

	if !got.GetTimestamp().AsTime().Equal(now) {
		t.Fatalf("expected timestamp %s, got %s", now, got.GetTimestamp().AsTime())
	}
}

func TestTimestampOrNilReturnsNilForZeroTime(t *testing.T) {
	if got := timestampOrNil(time.Time{}); got != nil {
		t.Fatalf("expected nil timestamp, got %v", got)
	}
}

func TestTimestampOrNilReturnsTimestamp(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	got := timestampOrNil(now)
	if got == nil {
		t.Fatal("expected timestamp, got nil")
	}

	if !got.AsTime().Equal(now) {
		t.Fatalf("expected timestamp %s, got %s", now, got.AsTime())
	}
}

func TestNewPresenceHandlerRequiresService(t *testing.T) {
	_, err := NewPresenceHandler(nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPresenceHandlerSetOnlineForwardsRequest(t *testing.T) {
	called := false

	handler := newTestPresenceHandler(t, &fakePresenceRepository{
		setOnlineFunc: func(ctx context.Context, userID, roomID string) error {
			called = true

			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			if roomID != "room-1" {
				t.Fatalf("expected room-1, got %q", roomID)
			}

			return nil
		},
	})

	_, err := handler.SetOnline(context.Background(), connect.NewRequest(&presencev1.SetOnlineRequest{
		UserId: " user-1 ",
		RoomId: " room-1 ",
	}))
	if err != nil {
		t.Fatalf("set online: %v", err)
	}

	if !called {
		t.Fatal("expected repository set online to be called")
	}
}

func TestPresenceHandlerSetOfflineForwardsRequest(t *testing.T) {
	called := false

	handler := newTestPresenceHandler(t, &fakePresenceRepository{
		setOfflineFunc: func(ctx context.Context, userID, roomID string) error {
			called = true

			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			if roomID != "room-1" {
				t.Fatalf("expected room-1, got %q", roomID)
			}

			return nil
		},
	})

	_, err := handler.SetOffline(context.Background(), connect.NewRequest(&presencev1.SetOfflineRequest{
		UserId: " user-1 ",
		RoomId: " room-1 ",
	}))
	if err != nil {
		t.Fatalf("set offline: %v", err)
	}

	if !called {
		t.Fatal("expected repository set offline to be called")
	}
}

func TestPresenceHandlerSetTypingMapsOfflineUser(t *testing.T) {
	handler := newTestPresenceHandler(t, &fakePresenceRepository{
		getPresenceFunc: func(ctx context.Context, userID string) (*presenceservice.Presence, error) {
			return &presenceservice.Presence{
				UserID: userID,
				Status: presenceservice.PresenceStatusOffline,
			}, nil
		},
	})

	_, err := handler.SetTyping(context.Background(), connect.NewRequest(&presencev1.SetTypingRequest{
		UserId: "user-1",
		RoomId: "room-1",
	}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("expected failed precondition, got %v", connect.CodeOf(err))
	}
}

func TestPresenceHandlerGetPresenceReturnsProtoResponse(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	handler := newTestPresenceHandler(t, &fakePresenceRepository{
		getPresenceFunc: func(ctx context.Context, userID string) (*presenceservice.Presence, error) {
			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			return &presenceservice.Presence{
				UserID:   userID,
				Status:   presenceservice.PresenceStatusOnline,
				LastSeen: now,
			}, nil
		},
	})

	res, err := handler.GetPresence(context.Background(), connect.NewRequest(&presencev1.GetPresenceRequest{
		UserId: " user-1 ",
	}))
	if err != nil {
		t.Fatalf("get presence: %v", err)
	}

	if res.Msg.GetStatus() != presencev1.PresenceStatus_PRESENCE_STATUS_ONLINE {
		t.Fatalf("expected online status, got %v", res.Msg.GetStatus())
	}

	if !res.Msg.GetLastSeen().AsTime().Equal(now) {
		t.Fatalf("expected last seen %s, got %s", now, res.Msg.GetLastSeen().AsTime())
	}
}

func TestPresenceHandlerGetRoomPresenceReturnsEvents(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	handler := newTestPresenceHandler(t, &fakePresenceRepository{
		getRoomPresenceFunc: func(ctx context.Context, roomID string) ([]presenceservice.Presence, error) {
			if roomID != "room-1" {
				t.Fatalf("expected room-1, got %q", roomID)
			}

			return []presenceservice.Presence{
				{
					UserID:   "user-1",
					RoomID:   "room-1",
					Status:   presenceservice.PresenceStatusTyping,
					LastSeen: now,
				},
			}, nil
		},
	})

	res, err := handler.GetRoomPresence(context.Background(), connect.NewRequest(&presencev1.GetRoomPresenceRequest{
		RoomId: " room-1 ",
	}))
	if err != nil {
		t.Fatalf("get room presence: %v", err)
	}

	if len(res.Msg.GetPresences()) != 1 {
		t.Fatalf("expected 1 presence, got %d", len(res.Msg.GetPresences()))
	}

	event := res.Msg.GetPresences()[0]
	if event.GetUserId() != "user-1" {
		t.Fatalf("expected user-1, got %q", event.GetUserId())
	}

	if event.GetStatus() != presencev1.PresenceStatus_PRESENCE_STATUS_TYPING {
		t.Fatalf("expected typing status, got %v", event.GetStatus())
	}
}

type fakePresenceRepository struct {
	setOnlineFunc       func(ctx context.Context, userID, roomID string) error
	setOfflineFunc      func(ctx context.Context, userID, roomID string) error
	setTypingFunc       func(ctx context.Context, roomID, userID string) error
	getPresenceFunc     func(ctx context.Context, userID string) (*presenceservice.Presence, error)
	getRoomPresenceFunc func(ctx context.Context, roomID string) ([]presenceservice.Presence, error)
}

func (r *fakePresenceRepository) SetOnline(ctx context.Context, userID, roomID string) error {
	if r.setOnlineFunc != nil {
		return r.setOnlineFunc(ctx, userID, roomID)
	}

	return nil
}

func (r *fakePresenceRepository) SetOffline(ctx context.Context, userID, roomID string) error {
	if r.setOfflineFunc != nil {
		return r.setOfflineFunc(ctx, userID, roomID)
	}

	return nil
}

func (r *fakePresenceRepository) SetTyping(ctx context.Context, roomID, userID string) error {
	if r.setTypingFunc != nil {
		return r.setTypingFunc(ctx, roomID, userID)
	}

	return nil
}

func (r *fakePresenceRepository) GetPresence(ctx context.Context, userID string) (*presenceservice.Presence, error) {
	if r.getPresenceFunc != nil {
		return r.getPresenceFunc(ctx, userID)
	}

	return &presenceservice.Presence{
		UserID: userID,
		Status: presenceservice.PresenceStatusOnline,
	}, nil
}

func (r *fakePresenceRepository) GetRoomPresence(ctx context.Context, roomID string) ([]presenceservice.Presence, error) {
	if r.getRoomPresenceFunc != nil {
		return r.getRoomPresenceFunc(ctx, roomID)
	}

	return nil, nil
}

func newTestPresenceHandler(t *testing.T, repo *fakePresenceRepository) *PresenceHandler {
	t.Helper()

	service, err := presenceservice.NewPresenceService(repo)
	if err != nil {
		t.Fatalf("create presence service: %v", err)
	}

	handler, err := NewPresenceHandler(service)
	if err != nil {
		t.Fatalf("create presence handler: %v", err)
	}

	return handler
}
