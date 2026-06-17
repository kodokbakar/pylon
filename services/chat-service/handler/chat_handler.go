package handler

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/kodokbakar/pylon/gen/pylon/chat/v1"
	chatv1connect "github.com/kodokbakar/pylon/gen/pylon/chat/v1/chatv1connect"
	chatservice "github.com/kodokbakar/pylon/services/chat-service/service"
)

var _ chatv1connect.ChatServiceHandler = (*ChatHandler)(nil)

type ChatHandler struct {
	service *chatservice.ChatService
}

func NewChatHandler(service *chatservice.ChatService) (*ChatHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("chat service is required")
	}

	return &ChatHandler{service: service}, nil
}

func (h *ChatHandler) SendMessage(
	ctx context.Context,
	req *connect.Request[chatv1.SendMessageRequest],
) (*connect.Response[chatv1.SendMessageResponse], error) {
	msg, err := h.service.SendMessage(ctx, chatservice.SendMessageInput{
		RoomID:   req.Msg.GetRoomId(),
		SenderID: req.Msg.GetSenderId(),
		Content:  req.Msg.GetContent(),
		Type:     protoMessageTypeToDomain(req.Msg.GetType()),
	})
	if err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&chatv1.SendMessageResponse{
		Message: domainMessageToProto(msg),
	}), nil
}

func (h *ChatHandler) StreamMessages(
	ctx context.Context,
	req *connect.Request[chatv1.StreamMessagesRequest],
	stream *connect.ServerStream[chatv1.Message],
) error {
	messages, err := h.service.StreamMessages(ctx, chatservice.StreamMessagesInput{
		RoomID: req.Msg.GetRoomId(),
		UserID: req.Msg.GetUserId(),
	})
	if err != nil {
		return connectError(err)
	}

	for msg := range messages {
		if err := stream.Send(domainMessageToProto(msg)); err != nil {
			return fmt.Errorf("send streamed message: %w", err)
		}
	}

	return nil
}

func (h *ChatHandler) GetMessages(
	ctx context.Context,
	req *connect.Request[chatv1.GetMessagesRequest],
) (*connect.Response[chatv1.GetMessagesResponse], error) {
	result, err := h.service.GetMessages(ctx, chatservice.GetMessagesInput{
		RoomID:   req.Msg.GetRoomId(),
		Limit:    int(req.Msg.GetLimit()),
		BeforeID: req.Msg.GetBeforeId(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	messages := make([]*chatv1.Message, 0, len(result.Messages))
	for i := range result.Messages {
		messages = append(messages, domainMessageToProto(&result.Messages[i]))
	}

	return connect.NewResponse(&chatv1.GetMessagesResponse{
		Messages: messages,
		HasMore:  result.HasMore,
	}), nil
}

func connectError(err error) error {
	if errors.Is(err, chatservice.ErrInvalidInput) {
		return connect.NewError(connect.CodeInvalidArgument, err)
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

func protoMessageTypeToDomain(messageType chatv1.MessageType) chatservice.MessageType {
	switch messageType {
	case chatv1.MessageType_MESSAGE_TYPE_IMAGE:
		return chatservice.MessageTypeImage
	case chatv1.MessageType_MESSAGE_TYPE_SYSTEM:
		return chatservice.MessageTypeSystem
	default:
		return chatservice.MessageTypeText
	}
}

func domainMessageTypeToProto(messageType chatservice.MessageType) chatv1.MessageType {
	switch messageType {
	case chatservice.MessageTypeImage:
		return chatv1.MessageType_MESSAGE_TYPE_IMAGE
	case chatservice.MessageTypeSystem:
		return chatv1.MessageType_MESSAGE_TYPE_SYSTEM
	default:
		return chatv1.MessageType_MESSAGE_TYPE_TEXT
	}
}

func domainMessageToProto(msg *chatservice.Message) *chatv1.Message {
	if msg == nil {
		return nil
	}

	return &chatv1.Message{
		Id:        msg.ID,
		RoomId:    msg.RoomID,
		SenderId:  msg.SenderID,
		Content:   msg.Content,
		Type:      domainMessageTypeToProto(msg.Type),
		CreatedAt: timestamppb.New(msg.CreatedAt),
	}
}
