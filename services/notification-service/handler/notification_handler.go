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

	notificationv1 "github.com/kodokbakar/pylon/gen/pylon/notification/v1"
	notificationv1connect "github.com/kodokbakar/pylon/gen/pylon/notification/v1/notificationv1connect"
	notificationservice "github.com/kodokbakar/pylon/services/notification-service/service"
)

var _ notificationv1connect.NotificationServiceHandler = (*NotificationHandler)(nil)

// NotificationHandler is an internal Connect RPC handler.
// Authentication and caller identity enforcement are handled by API Gateway;
// do not expose this service directly to public clients without adding auth.
type NotificationHandler struct {
	service *notificationservice.NotificationService
}

func NewNotificationHandler(service *notificationservice.NotificationService) (*NotificationHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("notification service is required")
	}

	return &NotificationHandler{service: service}, nil
}

func (h *NotificationHandler) SendNotification(
	ctx context.Context,
	req *connect.Request[notificationv1.SendNotificationRequest],
) (*connect.Response[notificationv1.Notification], error) {
	notification, err := h.service.SendNotification(ctx, notificationservice.SendNotificationInput{
		UserID:    req.Msg.GetUserId(),
		Type:      protoNotificationTypeToDomain(req.Msg.GetType()),
		Title:     req.Msg.GetTitle(),
		Body:      req.Msg.GetBody(),
		RoomID:    req.Msg.GetRoomId(),
		MessageID: req.Msg.GetMessageId(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(domainNotificationToProto(notification)), nil
}

func (h *NotificationHandler) GetNotifications(
	ctx context.Context,
	req *connect.Request[notificationv1.GetNotificationsRequest],
) (*connect.Response[notificationv1.GetNotificationsResponse], error) {
	result, err := h.service.GetNotifications(ctx, notificationservice.GetNotificationsInput{
		UserID:     req.Msg.GetUserId(),
		Limit:      int(req.Msg.GetLimit()),
		Offset:     int(req.Msg.GetOffset()),
		UnreadOnly: req.Msg.GetUnreadOnly(),
	})
	if err != nil {
		return nil, connectError(err)
	}

	protoNotifications := make([]*notificationv1.Notification, 0, len(result.Notifications))
	for i := range result.Notifications {
		protoNotifications = append(protoNotifications, domainNotificationToProto(&result.Notifications[i]))
	}

	return connect.NewResponse(&notificationv1.GetNotificationsResponse{
		Notifications: protoNotifications,
		UnreadCount:   int32(result.UnreadCount),
	}), nil
}

func (h *NotificationHandler) MarkAsRead(
	ctx context.Context,
	req *connect.Request[notificationv1.MarkAsReadRequest],
) (*connect.Response[emptypb.Empty], error) {
	if err := h.service.MarkAsRead(ctx, notificationservice.MarkAsReadInput{
		NotificationID: req.Msg.GetNotificationId(),
		UserID:         req.Msg.GetUserId(),
	}); err != nil {
		return nil, connectError(err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func connectError(err error) error {
	if errors.Is(err, notificationservice.ErrInvalidInput) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	if errors.Is(err, notificationservice.ErrNotFound) {
		return connect.NewError(connect.CodeNotFound, err)
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

func protoNotificationTypeToDomain(notificationType notificationv1.NotificationType) notificationservice.NotificationType {
	switch notificationType {
	case notificationv1.NotificationType_NOTIFICATION_TYPE_MESSAGE:
		return notificationservice.NotificationTypeMessage
	case notificationv1.NotificationType_NOTIFICATION_TYPE_INVITE:
		return notificationservice.NotificationTypeInvite
	case notificationv1.NotificationType_NOTIFICATION_TYPE_MENTION:
		return notificationservice.NotificationTypeMention
	default:
		return ""
	}
}

func domainNotificationTypeToProto(notificationType notificationservice.NotificationType) notificationv1.NotificationType {
	switch notificationType {
	case notificationservice.NotificationTypeMessage:
		return notificationv1.NotificationType_NOTIFICATION_TYPE_MESSAGE
	case notificationservice.NotificationTypeInvite:
		return notificationv1.NotificationType_NOTIFICATION_TYPE_INVITE
	case notificationservice.NotificationTypeMention:
		return notificationv1.NotificationType_NOTIFICATION_TYPE_MENTION
	default:
		return notificationv1.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED
	}
}

func domainNotificationToProto(notification *notificationservice.Notification) *notificationv1.Notification {
	if notification == nil {
		return nil
	}

	return &notificationv1.Notification{
		Id:        notification.ID,
		UserId:    notification.UserID,
		Type:      domainNotificationTypeToProto(notification.Type),
		Title:     notification.Title,
		Body:      notification.Body,
		RoomId:    notification.RoomID,
		Read:      notification.Read,
		CreatedAt: timestampOrNil(notification.CreatedAt),
		MessageId: notification.MessageID,
	}
}

func timestampOrNil(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}

	return timestamppb.New(t)
}
