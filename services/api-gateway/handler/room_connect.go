package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ roomv1connect.RoomServiceHandler = (*RoomConnectHandler)(nil)

type RoomConnectClient interface {
	CreateRoom(context.Context, *connect.Request[roomv1.CreateRoomRequest]) (*connect.Response[roomv1.CreateRoomResponse], error)
	GetRoom(context.Context, *connect.Request[roomv1.GetRoomRequest]) (*connect.Response[roomv1.GetRoomResponse], error)
	ListRooms(context.Context, *connect.Request[roomv1.ListRoomsRequest]) (*connect.Response[roomv1.ListRoomsResponse], error)
	JoinRoom(context.Context, *connect.Request[roomv1.JoinRoomRequest]) (*connect.Response[emptypb.Empty], error)
	LeaveRoom(context.Context, *connect.Request[roomv1.LeaveRoomRequest]) (*connect.Response[emptypb.Empty], error)
	GetRoomMembers(context.Context, *connect.Request[roomv1.GetRoomMembersRequest]) (*connect.Response[roomv1.GetRoomMembersResponse], error)
}

type RoomConnectHandler struct {
	client RoomConnectClient
}

func NewRoomConnectHandler(client RoomConnectClient) (*RoomConnectHandler, error) {
	if client == nil {
		return nil, fmt.Errorf("room connect client is required")
	}

	return &RoomConnectHandler{client: client}, nil
}

func (h *RoomConnectHandler) CreateRoom(
	ctx context.Context,
	req *connect.Request[roomv1.CreateRoomRequest],
) (*connect.Response[roomv1.CreateRoomResponse], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.CreateRoom(ctx, connect.NewRequest(&roomv1.CreateRoomRequest{
		Name:      req.Msg.GetName(),
		Type:      req.Msg.GetType(),
		CreatorId: userID,
		MemberIds: req.Msg.GetMemberIds(),
	}))
}

func (h *RoomConnectHandler) GetRoom(
	ctx context.Context,
	req *connect.Request[roomv1.GetRoomRequest],
) (*connect.Response[roomv1.GetRoomResponse], error) {
	if _, err := authenticatedUserID(ctx); err != nil {
		return nil, err
	}

	return h.client.GetRoom(ctx, connect.NewRequest(&roomv1.GetRoomRequest{
		RoomId: req.Msg.GetRoomId(),
	}))
}

func (h *RoomConnectHandler) ListRooms(
	ctx context.Context,
	req *connect.Request[roomv1.ListRoomsRequest],
) (*connect.Response[roomv1.ListRoomsResponse], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.ListRooms(ctx, connect.NewRequest(&roomv1.ListRoomsRequest{
		UserId: userID,
	}))
}

func (h *RoomConnectHandler) JoinRoom(
	ctx context.Context,
	req *connect.Request[roomv1.JoinRoomRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.JoinRoom(ctx, connect.NewRequest(&roomv1.JoinRoomRequest{
		RoomId: req.Msg.GetRoomId(),
		UserId: userID,
	}))
}

func (h *RoomConnectHandler) LeaveRoom(
	ctx context.Context,
	req *connect.Request[roomv1.LeaveRoomRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.LeaveRoom(ctx, connect.NewRequest(&roomv1.LeaveRoomRequest{
		RoomId: req.Msg.GetRoomId(),
		UserId: userID,
	}))
}

func (h *RoomConnectHandler) GetRoomMembers(
	ctx context.Context,
	req *connect.Request[roomv1.GetRoomMembersRequest],
) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
	if _, err := authenticatedUserID(ctx); err != nil {
		return nil, err
	}

	return h.client.GetRoomMembers(ctx, connect.NewRequest(&roomv1.GetRoomMembersRequest{
		RoomId: req.Msg.GetRoomId(),
	}))
}

func authenticatedUserID(ctx context.Context) (string, error) {
	userID, ok := gatewaymiddleware.UserIDFromContext(ctx)
	if !ok || strings.TrimSpace(userID) == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("user id is required"))
	}

	return strings.TrimSpace(userID), nil
}
