package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	presencev1 "github.com/kodokbakar/pylon/gen/pylon/presence/v1"
	presencev1connect "github.com/kodokbakar/pylon/gen/pylon/presence/v1/presencev1connect"
	presenceservice "github.com/kodokbakar/pylon/services/presence-service/service"
)

var _ presencev1connect.PresenceServiceHandler = (*PresenceHandler)(nil)

type PresenceHandler struct {
	service *presenceservice.PresenceService
}

func NewPresenceHandler(service *presenceservice.PresenceService) (*PresenceHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("presence service is required")
	}

	return &PresenceHandler{service: service}, nil
}

func (h *PresenceHandler) SetOnline(
	ctx context.Context,
	req *connect.Request[presencev1.SetOnlineRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.SetOnline(ctx, presenceservice.SetOnlineInput{
		UserID: req.Msg.GetUserId(),
	}); err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *PresenceHandler) SetOffline(
	ctx context.Context,
	req *connect.Request[presencev1.SetOfflineRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.SetOffline(ctx, presenceservice.SetOfflineInput{
		UserID: req.Msg.GetUserId(),
	}); err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *PresenceHandler) SetTyping(
	ctx context.Context,
	req *connect.Request[presencev1.SetTypingRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.SetTyping(ctx, presenceservice.SetTypingInput{
		UserID: req.Msg.GetUserId(),
		RoomID: req.Msg.GetRoomId(),
	}); err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *PresenceHandler) GetPresence(
	ctx context.Context,
	req *connect.Request[presencev1.GetPresenceRequest],
) (*connect.Response[presencev1.GetPresenceResponse], error) {
	presence, err := h.service.GetPresence(ctx, presenceservice.GetPresenceInput{
		UserID: req.Msg.GetUserId(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&presencev1.GetPresenceResponse{
		Status:   domainStatusToProto(presence.Status),
		LastSeen: timestampOrNil(presence.LastSeen),
	}), nil
}

func (h *PresenceHandler) GetRoomPresence(
	ctx context.Context,
	req *connect.Request[presencev1.GetRoomPresenceRequest],
) (*connect.Response[presencev1.GetRoomPresenceResponse], error) {
	presences, err := h.service.GetRoomPresence(ctx, presenceservice.GetRoomPresenceInput{
		RoomID: req.Msg.GetRoomId(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	events := make([]*presencev1.PresenceEvent, 0, len(presences))
	for _, presence := range presences {
		events = append(events, &presencev1.PresenceEvent{
			UserId:    presence.UserID,
			RoomId:    presence.RoomID,
			Status:    domainStatusToProto(presence.Status),
			Timestamp: timestampOrNil(presence.LastSeen),
		})
	}

	return connect.NewResponse(&presencev1.GetRoomPresenceResponse{
		Presences: events,
	}), nil
}

func (h *PresenceHandler) StreamPresence(
	ctx context.Context,
	req *connect.Request[presencev1.StreamPresenceRequest],
	stream *connect.ServerStream[presencev1.PresenceEvent],
) error {
	events, err := h.service.StreamPresence(ctx, presenceservice.StreamPresenceInput{
		RoomID: req.Msg.GetRoomId(),
	})
	if err != nil {
		return connectError(err)
	}

	for event := range events {
		if err := stream.Send(domainEventToProto(event)); err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("send streamed presence event: %w", err))
		}
	}

	return nil
}

func connectError(err error) error {
	if errors.Is(err, presenceservice.ErrInvalidInput) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	if errors.Is(err, presenceservice.ErrUserOffline) {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}

	return connect.NewError(connect.CodeInternal, err)
}

func domainStatusToProto(status presenceservice.PresenceStatus) presencev1.PresenceStatus {
	switch status {
	case presenceservice.PresenceStatusOnline:
		return presencev1.PresenceStatus_PRESENCE_STATUS_ONLINE
	case presenceservice.PresenceStatusTyping:
		return presencev1.PresenceStatus_PRESENCE_STATUS_TYPING
	default:
		return presencev1.PresenceStatus_PRESENCE_STATUS_OFFLINE
	}
}

func domainEventToProto(event presenceservice.PresenceEvent) *presencev1.PresenceEvent {
	return &presencev1.PresenceEvent{
		UserId:    event.UserID,
		RoomId:    event.RoomID,
		Status:    domainStatusToProto(event.Status),
		Timestamp: timestampOrNil(event.Timestamp),
	}
}

func timestampOrNil(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}

	return timestamppb.New(t)
}
