package handler

import (
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgconn"

	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	roomservice "github.com/kodokbakar/pylon/services/room-service/service"
)

func TestConnectErrorMapsInvalidInput(t *testing.T) {
	err := connectError(roomservice.ErrInvalidInput)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeInvalidArgument {
		t.Fatalf("expected invalid argument, got %v", connectErr.Code())
	}
}

func TestConnectErrorMapsNotFound(t *testing.T) {
	err := connectError(roomservice.ErrNotFound)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeNotFound {
		t.Fatalf("expected not found, got %v", connectErr.Code())
	}
}

func TestConnectErrorMapsForbidden(t *testing.T) {
	err := connectError(roomservice.ErrForbidden)

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodePermissionDenied {
		t.Fatalf("expected permission denied, got %v", connectErr.Code())
	}
}

func TestConnectErrorMapsUniqueViolation(t *testing.T) {
	err := connectError(&pgconn.PgError{Code: "23505"})

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeAlreadyExists {
		t.Fatalf("expected already exists, got %v", connectErr.Code())
	}
}

func TestRoomTypeMapping(t *testing.T) {
	if got := protoRoomTypeToDomain(roomv1.RoomType_ROOM_TYPE_DIRECT); got != roomservice.RoomTypeDirect {
		t.Fatalf("expected direct, got %q", got)
	}

	if got := protoRoomTypeToDomain(roomv1.RoomType_ROOM_TYPE_GROUP); got != roomservice.RoomTypeGroup {
		t.Fatalf("expected group, got %q", got)
	}

	if got := protoRoomTypeToDomain(roomv1.RoomType_ROOM_TYPE_CHANNEL); got != roomservice.RoomTypeChannel {
		t.Fatalf("expected channel, got %q", got)
	}

	if got := domainRoomTypeToProto(roomservice.RoomTypeDirect); got != roomv1.RoomType_ROOM_TYPE_DIRECT {
		t.Fatalf("expected direct proto, got %v", got)
	}

	if got := domainRoomTypeToProto(roomservice.RoomTypeChannel); got != roomv1.RoomType_ROOM_TYPE_CHANNEL {
		t.Fatalf("expected channel proto, got %v", got)
	}
}

func TestDomainRoomToProto(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	got := domainRoomToProto(&roomservice.Room{
		ID:        "room-1",
		Name:      "General",
		Type:      roomservice.RoomTypeChannel,
		CreatedBy: "user-1",
		CreatedAt: now,
	})

	if got.GetId() != "room-1" {
		t.Fatalf("expected room-1, got %q", got.GetId())
	}

	if got.GetName() != "General" {
		t.Fatalf("expected General, got %q", got.GetName())
	}

	if got.GetType() != roomv1.RoomType_ROOM_TYPE_CHANNEL {
		t.Fatalf("expected channel, got %v", got.GetType())
	}

	if got.GetCreatedBy() != "user-1" {
		t.Fatalf("expected user-1, got %q", got.GetCreatedBy())
	}

	if !got.GetCreatedAt().AsTime().Equal(now) {
		t.Fatalf("expected created_at %s, got %s", now, got.GetCreatedAt().AsTime())
	}
}

func TestDomainMemberToProtoIncludesUserMetadata(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Nanosecond)

	got := domainMemberToProto(&roomservice.RoomMember{
		UserID:      "user-1",
		RoomID:      "room-1",
		Role:        roomservice.RoomRoleOwner,
		JoinedAt:    now,
		Username:    "alice",
		DisplayName: "Alice",
		AvatarURL:   "https://example.com/alice.png",
	})

	if got.GetUserId() != "user-1" {
		t.Fatalf("expected user-1, got %q", got.GetUserId())
	}

	if got.GetRoomId() != "room-1" {
		t.Fatalf("expected room-1, got %q", got.GetRoomId())
	}

	if got.GetRole() != roomservice.RoomRoleOwner {
		t.Fatalf("expected owner, got %q", got.GetRole())
	}

	if got.GetUsername() != "alice" {
		t.Fatalf("expected alice, got %q", got.GetUsername())
	}

	if got.GetDisplayName() != "Alice" {
		t.Fatalf("expected Alice, got %q", got.GetDisplayName())
	}

	if got.GetAvatarUrl() != "https://example.com/alice.png" {
		t.Fatalf("expected avatar url, got %q", got.GetAvatarUrl())
	}

	if !got.GetJoinedAt().AsTime().Equal(now) {
		t.Fatalf("expected joined_at %s, got %s", now, got.GetJoinedAt().AsTime())
	}
}

func TestTimestampOrNilReturnsNilForZeroTime(t *testing.T) {
	if got := timestampOrNil(time.Time{}); got != nil {
		t.Fatalf("expected nil timestamp, got %v", got)
	}
}
