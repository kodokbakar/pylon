package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type fakeRoomRepository struct {
	createFunc         func(ctx context.Context, input CreateRoomRecordInput) (*Room, error)
	getByIDFunc        func(ctx context.Context, roomID string) (*Room, error)
	listByUserIDFunc   func(ctx context.Context, userID string) ([]Room, error)
	findDirectRoomFunc func(ctx context.Context, userID, otherUserID string) (*Room, error)
}

func (r *fakeRoomRepository) Create(ctx context.Context, input CreateRoomRecordInput) (*Room, error) {
	if r.createFunc == nil {
		return nil, errors.New("create func is not configured")
	}

	return r.createFunc(ctx, input)
}

func (r *fakeRoomRepository) GetByID(ctx context.Context, roomID string) (*Room, error) {
	if r.getByIDFunc == nil {
		return nil, errors.New("get by id func is not configured")
	}

	return r.getByIDFunc(ctx, roomID)
}

func (r *fakeRoomRepository) ListByUserID(ctx context.Context, userID string) ([]Room, error) {
	if r.listByUserIDFunc == nil {
		return nil, errors.New("list by user id func is not configured")
	}

	return r.listByUserIDFunc(ctx, userID)
}

func (r *fakeRoomRepository) FindDirectRoom(ctx context.Context, userID, otherUserID string) (*Room, error) {
	if r.findDirectRoomFunc == nil {
		return nil, fmt.Errorf("%w: direct room not found", ErrNotFound)
	}

	return r.findDirectRoomFunc(ctx, userID, otherUserID)
}

type fakeMemberRepository struct {
	addFunc          func(ctx context.Context, input AddRoomMemberInput) error
	removeFunc       func(ctx context.Context, input RemoveRoomMemberInput) error
	listByRoomIDFunc func(ctx context.Context, roomID string) ([]RoomMember, error)
	getRoleFunc      func(ctx context.Context, roomID, userID string) (string, error)
}

func (r *fakeMemberRepository) Add(ctx context.Context, input AddRoomMemberInput) error {
	if r.addFunc == nil {
		return errors.New("add func is not configured")
	}

	return r.addFunc(ctx, input)
}

func (r *fakeMemberRepository) Remove(ctx context.Context, input RemoveRoomMemberInput) error {
	if r.removeFunc == nil {
		return errors.New("remove func is not configured")
	}

	return r.removeFunc(ctx, input)
}

func (r *fakeMemberRepository) ListByRoomID(ctx context.Context, roomID string) ([]RoomMember, error) {
	if r.listByRoomIDFunc == nil {
		return nil, errors.New("list by room id func is not configured")
	}

	return r.listByRoomIDFunc(ctx, roomID)
}

func (r *fakeMemberRepository) GetRole(ctx context.Context, roomID, userID string) (string, error) {
	if r.getRoleFunc == nil {
		return "", errors.New("get role func is not configured")
	}

	return r.getRoleFunc(ctx, roomID, userID)
}

func TestNewRoomServiceRequiresRoomRepository(t *testing.T) {
	_, err := NewRoomService(nil, &fakeMemberRepository{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewRoomServiceRequiresMemberRepository(t *testing.T) {
	_, err := NewRoomService(&fakeRoomRepository{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRoomValidatesName(t *testing.T) {
	svc, err := NewRoomService(&fakeRoomRepository{}, &fakeMemberRepository{})
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	_, err = svc.CreateRoom(context.Background(), CreateRoomInput{
		CreatorID: "user-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestCreateRoomCreatesRoomAndAddsMembers(t *testing.T) {
	addedMembers := make([]AddRoomMemberInput, 0)

	svc, err := NewRoomService(
		&fakeRoomRepository{
			createFunc: func(ctx context.Context, input CreateRoomRecordInput) (*Room, error) {
				if input.Name != "Team Room" {
					t.Fatalf("expected Team Room, got %q", input.Name)
				}

				if input.Type != RoomTypeGroup {
					t.Fatalf("expected group room type, got %q", input.Type)
				}

				if input.CreatedBy != "user-1" {
					t.Fatalf("expected user-1 creator, got %q", input.CreatedBy)
				}

				return &Room{
					ID:        "room-1",
					Name:      input.Name,
					Type:      input.Type,
					CreatedBy: input.CreatedBy,
					CreatedAt: time.Now(),
				}, nil
			},
		},
		&fakeMemberRepository{
			addFunc: func(ctx context.Context, input AddRoomMemberInput) error {
				addedMembers = append(addedMembers, input)
				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	room, err := svc.CreateRoom(context.Background(), CreateRoomInput{
		Name:      " Team Room ",
		CreatorID: " user-1 ",
		MemberIDs: []string{"user-2", "user-1", "user-2", " "},
	})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	if room.ID != "room-1" {
		t.Fatalf("expected room-1, got %q", room.ID)
	}

	if len(addedMembers) != 2 {
		t.Fatalf("expected 2 member inserts, got %d", len(addedMembers))
	}

	if addedMembers[0].UserID != "user-1" || addedMembers[0].Role != RoomRoleOwner {
		t.Fatalf("expected creator to be owner, got %+v", addedMembers[0])
	}

	if addedMembers[1].UserID != "user-2" || addedMembers[1].Role != RoomRoleMember {
		t.Fatalf("expected user-2 to be member, got %+v", addedMembers[1])
	}
}

func TestGetRoomValidatesRoomID(t *testing.T) {
	svc, err := NewRoomService(&fakeRoomRepository{}, &fakeMemberRepository{})
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	_, err = svc.GetRoom(context.Background(), GetRoomInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestListRoomsReturnsRepositoryValue(t *testing.T) {
	svc, err := NewRoomService(
		&fakeRoomRepository{
			listByUserIDFunc: func(ctx context.Context, userID string) ([]Room, error) {
				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}

				return []Room{
					{
						ID:   "room-1",
						Name: "General",
						Type: RoomTypeChannel,
					},
				}, nil
			},
		},
		&fakeMemberRepository{},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	rooms, err := svc.ListRooms(context.Background(), ListRoomsInput{
		UserID: " user-1 ",
	})
	if err != nil {
		t.Fatalf("list rooms: %v", err)
	}

	if len(rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(rooms))
	}

	if rooms[0].ID != "room-1" {
		t.Fatalf("expected room-1, got %q", rooms[0].ID)
	}
}

func TestJoinRoomValidatesUserID(t *testing.T) {
	svc, err := NewRoomService(&fakeRoomRepository{}, &fakeMemberRepository{})
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	err = svc.JoinRoom(context.Background(), JoinRoomInput{
		RoomID: "room-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestJoinRoomAddsMember(t *testing.T) {
	called := false

	svc, err := NewRoomService(
		&fakeRoomRepository{
			getByIDFunc: func(ctx context.Context, roomID string) (*Room, error) {
				if roomID != "room-1" {
					t.Fatalf("expected room-1, got %q", roomID)
				}

				return &Room{ID: roomID, Type: RoomTypeChannel}, nil
			},
		},
		&fakeMemberRepository{
			addFunc: func(ctx context.Context, input AddRoomMemberInput) error {
				called = true

				if input.RoomID != "room-1" {
					t.Fatalf("expected room-1, got %q", input.RoomID)
				}

				if input.UserID != "user-1" {
					t.Fatalf("expected user-1, got %q", input.UserID)
				}

				if input.Role != RoomRoleMember {
					t.Fatalf("expected member role, got %q", input.Role)
				}

				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	err = svc.JoinRoom(context.Background(), JoinRoomInput{
		RoomID: " room-1 ",
		UserID: " user-1 ",
	})
	if err != nil {
		t.Fatalf("join room: %v", err)
	}

	if !called {
		t.Fatal("expected member repository add to be called")
	}
}

func TestLeaveRoomCallsRepository(t *testing.T) {
	called := false

	svc, err := NewRoomService(
		&fakeRoomRepository{},
		&fakeMemberRepository{
			removeFunc: func(ctx context.Context, input RemoveRoomMemberInput) error {
				called = true

				if input.RoomID != "room-1" {
					t.Fatalf("expected room-1, got %q", input.RoomID)
				}

				if input.UserID != "user-1" {
					t.Fatalf("expected user-1, got %q", input.UserID)
				}

				return nil
			},
			getRoleFunc: func(ctx context.Context, roomID, userID string) (string, error) {
				if roomID != "room-1" {
					t.Fatalf("expected room-1, got %q", roomID)
				}

				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}

				return RoomRoleMember, nil
			},
		},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	err = svc.LeaveRoom(context.Background(), LeaveRoomInput{
		RoomID: " room-1 ",
		UserID: " user-1 ",
	})
	if err != nil {
		t.Fatalf("leave room: %v", err)
	}

	if !called {
		t.Fatal("expected member repository remove to be called")
	}
}

func TestGetRoomMembersReturnsRepositoryValue(t *testing.T) {
	svc, err := NewRoomService(
		&fakeRoomRepository{
			getByIDFunc: func(ctx context.Context, roomID string) (*Room, error) {
				if roomID != "room-1" {
					t.Fatalf("expected room-1, got %q", roomID)
				}

				return &Room{ID: roomID}, nil
			},
		},
		&fakeMemberRepository{
			listByRoomIDFunc: func(ctx context.Context, roomID string) ([]RoomMember, error) {
				if roomID != "room-1" {
					t.Fatalf("expected room-1, got %q", roomID)
				}

				return []RoomMember{
					{
						RoomID:      roomID,
						UserID:      "user-1",
						Role:        RoomRoleOwner,
						Username:    "alice",
						DisplayName: "Alice",
						AvatarURL:   "https://example.com/alice.png",
					},
				}, nil
			},
		},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	members, err := svc.GetRoomMembers(context.Background(), GetRoomMembersInput{
		RoomID: " room-1 ",
	})
	if err != nil {
		t.Fatalf("get room members: %v", err)
	}

	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}

	if members[0].Role != RoomRoleOwner {
		t.Fatalf("expected owner role, got %q", members[0].Role)
	}

	if members[0].Username != "alice" {
		t.Fatalf("expected username alice, got %q", members[0].Username)
	}

	if members[0].DisplayName != "Alice" {
		t.Fatalf("expected display name Alice, got %q", members[0].DisplayName)
	}

	if members[0].AvatarURL != "https://example.com/alice.png" {
		t.Fatalf("expected avatar url, got %q", members[0].AvatarURL)
	}
}

func TestCreateRoomRejectsNameOverMaxLength(t *testing.T) {
	svc, err := NewRoomService(&fakeRoomRepository{}, &fakeMemberRepository{})
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	_, err = svc.CreateRoom(context.Background(), CreateRoomInput{
		Name:      strings.Repeat("a", MaxRoomNameLength+1),
		CreatorID: "user-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestCreateDirectRoomRequiresExactlyOneOtherMember(t *testing.T) {
	svc, err := NewRoomService(&fakeRoomRepository{}, &fakeMemberRepository{})
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	_, err = svc.CreateRoom(context.Background(), CreateRoomInput{
		Name:      "DM",
		Type:      RoomTypeDirect,
		CreatorID: "user-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestCreateDirectRoomReturnsExistingRoom(t *testing.T) {
	existing := &Room{
		ID:        "room-existing",
		Name:      "DM",
		Type:      RoomTypeDirect,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}

	svc, err := NewRoomService(
		&fakeRoomRepository{
			findDirectRoomFunc: func(ctx context.Context, userID, otherUserID string) (*Room, error) {
				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}

				if otherUserID != "user-2" {
					t.Fatalf("expected user-2, got %q", otherUserID)
				}

				return existing, nil
			},
		},
		&fakeMemberRepository{},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	room, err := svc.CreateRoom(context.Background(), CreateRoomInput{
		Name:      "DM",
		Type:      RoomTypeDirect,
		CreatorID: "user-1",
		MemberIDs: []string{"user-2"},
	})
	if err != nil {
		t.Fatalf("create direct room: %v", err)
	}

	if room.ID != existing.ID {
		t.Fatalf("expected existing room %q, got %q", existing.ID, room.ID)
	}
}

func TestJoinRoomRejectsDirectRoom(t *testing.T) {
	svc, err := NewRoomService(
		&fakeRoomRepository{
			getByIDFunc: func(ctx context.Context, roomID string) (*Room, error) {
				return &Room{
					ID:   roomID,
					Type: RoomTypeDirect,
				}, nil
			},
		},
		&fakeMemberRepository{},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	err = svc.JoinRoom(context.Background(), JoinRoomInput{
		RoomID: "room-1",
		UserID: "user-1",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestJoinRoomRejectsGroupRoom(t *testing.T) {
	svc, err := NewRoomService(
		&fakeRoomRepository{
			getByIDFunc: func(ctx context.Context, roomID string) (*Room, error) {
				return &Room{
					ID:   roomID,
					Type: RoomTypeGroup,
				}, nil
			},
		},
		&fakeMemberRepository{},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	err = svc.JoinRoom(context.Background(), JoinRoomInput{
		RoomID: "room-1",
		UserID: "user-1",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestLeaveRoomRejectsOwner(t *testing.T) {
	svc, err := NewRoomService(
		&fakeRoomRepository{},
		&fakeMemberRepository{
			getRoleFunc: func(ctx context.Context, roomID, userID string) (string, error) {
				return RoomRoleOwner, nil
			},
		},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	err = svc.LeaveRoom(context.Background(), LeaveRoomInput{
		RoomID: "room-1",
		UserID: "user-1",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestGetRoomMembersReturnsNotFoundWhenRoomDoesNotExist(t *testing.T) {
	svc, err := NewRoomService(
		&fakeRoomRepository{
			getByIDFunc: func(ctx context.Context, roomID string) (*Room, error) {
				return nil, fmt.Errorf("%w: room %s", ErrNotFound, roomID)
			},
		},
		&fakeMemberRepository{},
	)
	if err != nil {
		t.Fatalf("create room service: %v", err)
	}

	_, err = svc.GetRoomMembers(context.Background(), GetRoomMembersInput{
		RoomID: "room-1",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
