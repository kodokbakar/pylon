package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
	roomservice "github.com/kodokbakar/pylon/services/room-service/service"
)

var _ roomv1connect.RoomServiceHandler = (*RoomHandler)(nil)

// RoomHandler is an internal Connect RPC handler.
// Authentication and caller identity enforcement are handled by API Gateway;
// do not expose this service directly to public clients without adding auth.
type RoomHandler struct {
	service *roomservice.RoomService
}

func NewRoomHandler(service *roomservice.RoomService) (*RoomHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("room service is required")
	}

	return &RoomHandler{service: service}, nil
}

func (h *RoomHandler) CreateRoom(
	ctx context.Context,
	req *connect.Request[roomv1.CreateRoomRequest],
) (*connect.Response[roomv1.CreateRoomResponse], error) {
	room, err := h.service.CreateRoom(ctx, roomservice.CreateRoomInput{
		Name:      req.Msg.GetName(),
		Type:      protoRoomTypeToDomain(req.Msg.GetType()),
		CreatorID: req.Msg.GetCreatorId(),
		MemberIDs: req.Msg.GetMemberIds(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&roomv1.CreateRoomResponse{
		Room: domainRoomToProto(room),
	}), nil
}

func (h *RoomHandler) GetRoom(
	ctx context.Context,
	req *connect.Request[roomv1.GetRoomRequest],
) (*connect.Response[roomv1.GetRoomResponse], error) {
	room, err := h.service.GetRoom(ctx, roomservice.GetRoomInput{
		RoomID: req.Msg.GetRoomId(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&roomv1.GetRoomResponse{
		Room: domainRoomToProto(room),
	}), nil
}

func (h *RoomHandler) ListRooms(
	ctx context.Context,
	req *connect.Request[roomv1.ListRoomsRequest],
) (*connect.Response[roomv1.ListRoomsResponse], error) {
	rooms, err := h.service.ListRooms(ctx, roomservice.ListRoomsInput{
		UserID: req.Msg.GetUserId(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	protoRooms := make([]*roomv1.Room, 0, len(rooms))
	for i := range rooms {
		protoRooms = append(protoRooms, domainRoomToProto(&rooms[i]))
	}

	return connect.NewResponse(&roomv1.ListRoomsResponse{
		Rooms: protoRooms,
	}), nil
}

func (h *RoomHandler) JoinRoom(
	ctx context.Context,
	req *connect.Request[roomv1.JoinRoomRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.JoinRoom(ctx, roomservice.JoinRoomInput{
		RoomID: req.Msg.GetRoomId(),
		UserID: req.Msg.GetUserId(),
	}); err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *RoomHandler) LeaveRoom(
	ctx context.Context,
	req *connect.Request[roomv1.LeaveRoomRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.LeaveRoom(ctx, roomservice.LeaveRoomInput{
		RoomID: req.Msg.GetRoomId(),
		UserID: req.Msg.GetUserId(),
	}); err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *RoomHandler) GetRoomMembers(
	ctx context.Context,
	req *connect.Request[roomv1.GetRoomMembersRequest],
) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
	members, err := h.service.GetRoomMembers(ctx, roomservice.GetRoomMembersInput{
		RoomID: req.Msg.GetRoomId(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	protoMembers := make([]*roomv1.RoomMember, 0, len(members))
	for i := range members {
		protoMembers = append(protoMembers, domainMemberToProto(&members[i]))
	}

	return connect.NewResponse(&roomv1.GetRoomMembersResponse{
		Members: protoMembers,
	}), nil
}

func connectError(err error) error {
	if errors.Is(err, roomservice.ErrInvalidInput) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	if errors.Is(err, roomservice.ErrNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
	}

	if errors.Is(err, roomservice.ErrForbidden) {
		return connect.NewError(connect.CodePermissionDenied, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			return connect.NewError(connect.CodeFailedPrecondition, err)
		case "23505":
			return connect.NewError(connect.CodeAlreadyExists, err)
		case "23502", "23514":
			return connect.NewError(connect.CodeInvalidArgument, err)
		}
	}

	return connect.NewError(connect.CodeInternal, err)
}

func protoRoomTypeToDomain(roomType roomv1.RoomType) roomservice.RoomType {
	switch roomType {
	case roomv1.RoomType_ROOM_TYPE_DIRECT:
		return roomservice.RoomTypeDirect
	case roomv1.RoomType_ROOM_TYPE_CHANNEL:
		return roomservice.RoomTypeChannel
	case roomv1.RoomType_ROOM_TYPE_GROUP:
		return roomservice.RoomTypeGroup
	default:
		return ""
	}
}

func domainRoomTypeToProto(roomType roomservice.RoomType) roomv1.RoomType {
	switch roomType {
	case roomservice.RoomTypeDirect:
		return roomv1.RoomType_ROOM_TYPE_DIRECT
	case roomservice.RoomTypeChannel:
		return roomv1.RoomType_ROOM_TYPE_CHANNEL
	default:
		return roomv1.RoomType_ROOM_TYPE_GROUP
	}
}

func domainRoomToProto(room *roomservice.Room) *roomv1.Room {
	if room == nil {
		return nil
	}

	return &roomv1.Room{
		Id:        room.ID,
		Name:      room.Name,
		Type:      domainRoomTypeToProto(room.Type),
		CreatedBy: room.CreatedBy,
		CreatedAt: timestampOrNil(room.CreatedAt),
	}
}

func domainMemberToProto(member *roomservice.RoomMember) *roomv1.RoomMember {
	if member == nil {
		return nil
	}

	return &roomv1.RoomMember{
		UserId:      member.UserID,
		RoomId:      member.RoomID,
		Role:        member.Role,
		JoinedAt:    timestampOrNil(member.JoinedAt),
		Username:    member.Username,
		DisplayName: member.DisplayName,
		AvatarUrl:   member.AvatarURL,
	}
}

func timestampOrNil(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}

	return timestamppb.New(t)
}
