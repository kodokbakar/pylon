package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New("invalid input")

var ErrNotFound = errors.New("not found")

var ErrForbidden = errors.New("forbidden")

const MaxRoomNameLength = 255

type RoomType string

const (
	RoomTypeDirect  RoomType = "direct"
	RoomTypeGroup   RoomType = "group"
	RoomTypeChannel RoomType = "channel"
)

const (
	RoomRoleOwner  = "owner"
	RoomRoleAdmin  = "admin"
	RoomRoleMember = "member"
)

type Room struct {
	ID        string
	Name      string
	Type      RoomType
	CreatedBy string
	CreatedAt time.Time
}

type RoomMember struct {
	UserID      string
	RoomID      string
	Role        string
	JoinedAt    time.Time
	Username    string
	DisplayName string
	AvatarURL   string
}

type CreateRoomInput struct {
	Name      string
	Type      RoomType
	CreatorID string
	MemberIDs []string
}

type CreateRoomRecordInput struct {
	Name      string
	Type      RoomType
	CreatedBy string
}

type GetRoomInput struct {
	RoomID string
}

type ListRoomsInput struct {
	UserID string
}

type JoinRoomInput struct {
	RoomID string
	UserID string
}

type LeaveRoomInput struct {
	RoomID string
	UserID string
}

type GetRoomMembersInput struct {
	RoomID string
}

type AddRoomMemberInput struct {
	RoomID string
	UserID string
	Role   string
}

type RemoveRoomMemberInput struct {
	RoomID string
	UserID string
}

type RoomRepository interface {
	Create(ctx context.Context, input CreateRoomRecordInput) (*Room, error)
	GetByID(ctx context.Context, roomID string) (*Room, error)
	ListByUserID(ctx context.Context, userID string) ([]Room, error)
	FindDirectRoom(ctx context.Context, userID, otherUserID string) (*Room, error)
}

type MemberRepository interface {
	Add(ctx context.Context, input AddRoomMemberInput) error
	Remove(ctx context.Context, input RemoveRoomMemberInput) error
	ListByRoomID(ctx context.Context, roomID string) ([]RoomMember, error)
	GetRole(ctx context.Context, roomID, userID string) (string, error)
}

type RoomService struct {
	rooms   RoomRepository
	members MemberRepository
}

func NewRoomService(rooms RoomRepository, members MemberRepository) (*RoomService, error) {
	if rooms == nil {
		return nil, fmt.Errorf("room repository is required")
	}

	if members == nil {
		return nil, fmt.Errorf("member repository is required")
	}

	return &RoomService{
		rooms:   rooms,
		members: members,
	}, nil
}

func (s *RoomService) CreateRoom(ctx context.Context, input CreateRoomInput) (*Room, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.CreatorID = strings.TrimSpace(input.CreatorID)
	input.Type = normalizeRoomType(input.Type)

	if err := validateCreateRoomInput(input); err != nil {
		return nil, err
	}

	memberIDs := uniqueMemberIDs(input.MemberIDs, input.CreatorID)

	if input.Type == RoomTypeDirect {
		existingRoom, err := s.rooms.FindDirectRoom(ctx, input.CreatorID, memberIDs[0])
		if err == nil {
			return existingRoom, nil
		}

		if !errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("find existing direct room: %w", err)
		}
	}

	room, err := s.rooms.Create(ctx, CreateRoomRecordInput{
		Name:      input.Name,
		Type:      input.Type,
		CreatedBy: input.CreatorID,
	})
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	if err := s.members.Add(ctx, AddRoomMemberInput{
		RoomID: room.ID,
		UserID: input.CreatorID,
		Role:   RoomRoleOwner,
	}); err != nil {
		return nil, fmt.Errorf("add room owner: %w", err)
	}

	for _, memberID := range memberIDs {
		if err := s.members.Add(ctx, AddRoomMemberInput{
			RoomID: room.ID,
			UserID: memberID,
			Role:   RoomRoleMember,
		}); err != nil {
			return nil, fmt.Errorf("add room member %s: %w", memberID, err)
		}
	}

	return room, nil
}

func (s *RoomService) GetRoom(ctx context.Context, input GetRoomInput) (*Room, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if roomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	room, err := s.rooms.GetByID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}

	return room, nil
}

func (s *RoomService) ListRooms(ctx context.Context, input ListRoomsInput) ([]Room, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	rooms, err := s.rooms.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}

	return rooms, nil
}

func (s *RoomService) JoinRoom(ctx context.Context, input JoinRoomInput) error {
	roomID := strings.TrimSpace(input.RoomID)
	userID := strings.TrimSpace(input.UserID)

	if roomID == "" {
		return fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	room, err := s.rooms.GetByID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("get room before join: %w", err)
	}

	if room.Type != RoomTypeChannel {
		return fmt.Errorf("%w: only channel rooms can be joined directly", ErrForbidden)
	}

	if err := s.members.Add(ctx, AddRoomMemberInput{
		RoomID: roomID,
		UserID: userID,
		Role:   RoomRoleMember,
	}); err != nil {
		return fmt.Errorf("join room: %w", err)
	}

	return nil
}

func (s *RoomService) LeaveRoom(ctx context.Context, input LeaveRoomInput) error {
	roomID := strings.TrimSpace(input.RoomID)
	userID := strings.TrimSpace(input.UserID)

	if roomID == "" {
		return fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	role, err := s.members.GetRole(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("get room member role before leave: %w", err)
	}

	if role == RoomRoleOwner {
		return fmt.Errorf("%w: owner cannot leave room without ownership transfer", ErrForbidden)
	}

	if err := s.members.Remove(ctx, RemoveRoomMemberInput{
		RoomID: roomID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("leave room: %w", err)
	}

	return nil
}

func (s *RoomService) GetRoomMembers(ctx context.Context, input GetRoomMembersInput) ([]RoomMember, error) {
	roomID := strings.TrimSpace(input.RoomID)
	if roomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	if _, err := s.rooms.GetByID(ctx, roomID); err != nil {
		return nil, fmt.Errorf("get room before listing members: %w", err)
	}

	members, err := s.members.ListByRoomID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room members: %w", err)
	}

	return members, nil
}

func validateCreateRoomInput(input CreateRoomInput) error {
	if input.Name == "" {
		return fmt.Errorf("%w: room name is required", ErrInvalidInput)
	}

	if len([]rune(input.Name)) > MaxRoomNameLength {
		return fmt.Errorf("%w: room name exceeds maximum length of %d characters", ErrInvalidInput, MaxRoomNameLength)
	}

	if input.CreatorID == "" {
		return fmt.Errorf("%w: creator id is required", ErrInvalidInput)
	}

	if !isValidRoomType(input.Type) {
		return fmt.Errorf("%w: unsupported room type %q", ErrInvalidInput, input.Type)
	}

	if input.Type == RoomTypeDirect && len(uniqueMemberIDs(input.MemberIDs, input.CreatorID)) != 1 {
		return fmt.Errorf("%w: direct room requires exactly one other member", ErrInvalidInput)
	}

	return nil
}

func normalizeRoomType(roomType RoomType) RoomType {
	if roomType == "" {
		return RoomTypeGroup
	}

	return roomType
}

func isValidRoomType(roomType RoomType) bool {
	switch roomType {
	case RoomTypeDirect, RoomTypeGroup, RoomTypeChannel:
		return true
	default:
		return false
	}
}

func uniqueMemberIDs(memberIDs []string, creatorID string) []string {
	seen := map[string]struct{}{
		creatorID: {},
	}

	result := make([]string, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			continue
		}

		if _, exists := seen[memberID]; exists {
			continue
		}

		seen[memberID] = struct{}{}
		result = append(result, memberID)
	}

	return result
}
