package handler

import (
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
