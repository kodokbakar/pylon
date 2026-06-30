package handler

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	presencev1 "github.com/kodokbakar/pylon/gen/pylon/presence/v1"
	presencev1connect "github.com/kodokbakar/pylon/gen/pylon/presence/v1/presencev1connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ presencev1connect.PresenceServiceHandler = (*PresenceConnectHandler)(nil)

type PresenceConnectClient interface {
	SetOnline(context.Context, *connect.Request[presencev1.SetOnlineRequest]) (*connect.Response[emptypb.Empty], error)
	SetOffline(context.Context, *connect.Request[presencev1.SetOfflineRequest]) (*connect.Response[emptypb.Empty], error)
	SetTyping(context.Context, *connect.Request[presencev1.SetTypingRequest]) (*connect.Response[emptypb.Empty], error)
	GetPresence(context.Context, *connect.Request[presencev1.GetPresenceRequest]) (*connect.Response[presencev1.GetPresenceResponse], error)
	GetRoomPresence(context.Context, *connect.Request[presencev1.GetRoomPresenceRequest]) (*connect.Response[presencev1.GetRoomPresenceResponse], error)
	StreamPresence(context.Context, *connect.Request[presencev1.StreamPresenceRequest]) (*connect.ServerStreamForClient[presencev1.PresenceEvent], error)
}

type PresenceConnectHandler struct {
	client PresenceConnectClient
}

func NewPresenceConnectHandler(client PresenceConnectClient) (*PresenceConnectHandler, error) {
	if client == nil {
		return nil, fmt.Errorf("presence connect client is required")
	}

	return &PresenceConnectHandler{client: client}, nil
}

func (h *PresenceConnectHandler) SetOnline(
	ctx context.Context,
	req *connect.Request[presencev1.SetOnlineRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.SetOnline(ctx, connect.NewRequest(&presencev1.SetOnlineRequest{
		UserId: userID,
		RoomId: cleanPresenceRoomID(req.Msg.GetRoomId()),
	}))
}

func (h *PresenceConnectHandler) SetOffline(
	ctx context.Context,
	req *connect.Request[presencev1.SetOfflineRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.SetOffline(ctx, connect.NewRequest(&presencev1.SetOfflineRequest{
		UserId: userID,
		RoomId: cleanPresenceRoomID(req.Msg.GetRoomId()),
	}))
}

func (h *PresenceConnectHandler) SetTyping(
	ctx context.Context,
	req *connect.Request[presencev1.SetTypingRequest],
) (*connect.Response[emptypb.Empty], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.SetTyping(ctx, connect.NewRequest(&presencev1.SetTypingRequest{
		UserId: userID,
		RoomId: cleanPresenceRoomID(req.Msg.GetRoomId()),
	}))
}

func (h *PresenceConnectHandler) GetPresence(
	ctx context.Context,
	req *connect.Request[presencev1.GetPresenceRequest],
) (*connect.Response[presencev1.GetPresenceResponse], error) {
	userID, err := authenticatedUserID(ctx)
	if err != nil {
		return nil, err
	}

	return h.client.GetPresence(ctx, connect.NewRequest(&presencev1.GetPresenceRequest{
		UserId: userID,
	}))
}

func (h *PresenceConnectHandler) GetRoomPresence(
	ctx context.Context,
	req *connect.Request[presencev1.GetRoomPresenceRequest],
) (*connect.Response[presencev1.GetRoomPresenceResponse], error) {
	if _, err := authenticatedUserID(ctx); err != nil {
		return nil, err
	}

	return h.client.GetRoomPresence(ctx, connect.NewRequest(&presencev1.GetRoomPresenceRequest{
		RoomId: cleanPresenceRoomID(req.Msg.GetRoomId()),
	}))
}

func (h *PresenceConnectHandler) StreamPresence(
	ctx context.Context,
	req *connect.Request[presencev1.StreamPresenceRequest],
	stream *connect.ServerStream[presencev1.PresenceEvent],
) error {
	if _, err := authenticatedUserID(ctx); err != nil {
		return err
	}

	upstream, err := h.client.StreamPresence(ctx, connect.NewRequest(&presencev1.StreamPresenceRequest{
		RoomId: cleanPresenceRoomID(req.Msg.GetRoomId()),
	}))
	if err != nil {
		return err
	}

	for upstream.Receive() {
		if err := stream.Send(upstream.Msg()); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("send streamed presence event: %w", err))
		}
	}

	if err := upstream.Err(); err != nil {
		return err
	}

	return nil
}

func cleanPresenceRoomID(roomID string) string {
	return strings.TrimSpace(roomID)
}
