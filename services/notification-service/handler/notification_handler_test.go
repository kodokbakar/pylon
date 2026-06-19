package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"

	notificationv1 "github.com/kodokbakar/pylon/gen/pylon/notification/v1"
	notificationservice "github.com/kodokbakar/pylon/services/notification-service/service"
)

type fakeNotificationRepository struct {
	createFunc                      func(ctx context.Context, input notificationservice.CreateNotificationInput) (*notificationservice.Notification, error)
	listByUserIDFunc                func(ctx context.Context, input notificationservice.ListNotificationsInput) ([]notificationservice.Notification, error)
	countUnreadFunc                 func(ctx context.Context, userID string) (int, error)
	listByUserIDWithUnreadCountFunc func(ctx context.Context, input notificationservice.ListNotificationsInput) ([]notificationservice.Notification, int, error)
	markAsReadFunc                  func(ctx context.Context, notificationID, userID string) error
}

func (r *fakeNotificationRepository) Create(ctx context.Context, input notificationservice.CreateNotificationInput) (*notificationservice.Notification, error) {
	if r.createFunc == nil {
		return nil, errors.New("create func is not configured")
	}

	return r.createFunc(ctx, input)
}

func (r *fakeNotificationRepository) ListByUserID(ctx context.Context, input notificationservice.ListNotificationsInput) ([]notificationservice.Notification, error) {
	if r.listByUserIDFunc == nil {
		return nil, errors.New("list by user id func is not configured")
	}

	return r.listByUserIDFunc(ctx, input)
}

func (r *fakeNotificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	if r.countUnreadFunc == nil {
		return 0, errors.New("count unread func is not configured")
	}

	return r.countUnreadFunc(ctx, userID)
}

func (r *fakeNotificationRepository) ListByUserIDWithUnreadCount(
	ctx context.Context,
	input notificationservice.ListNotificationsInput,
) ([]notificationservice.Notification, int, error) {
	if r.listByUserIDWithUnreadCountFunc != nil {
		return r.listByUserIDWithUnreadCountFunc(ctx, input)
	}

	notifications, err := r.ListByUserID(ctx, input)
	if err != nil {
		return nil, 0, err
	}

	unreadCount, err := r.CountUnread(ctx, input.UserID)
	if err != nil {
		return nil, 0, err
	}

	return notifications, unreadCount, nil
}

func (r *fakeNotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	if r.markAsReadFunc == nil {
		return errors.New("mark as read func is not configured")
	}

	return r.markAsReadFunc(ctx, notificationID, userID)
}

func TestSendNotificationReturnsCreatedNotification(t *testing.T) {
	now := time.Now()

	svc, err := notificationservice.NewNotificationService(&fakeNotificationRepository{
		createFunc: func(ctx context.Context, input notificationservice.CreateNotificationInput) (*notificationservice.Notification, error) {
			if input.UserID != "user-1" {
				t.Fatalf("expected user-1, got %q", input.UserID)
			}

			if input.MessageID != "message-1" {
				t.Fatalf("expected message-1, got %q", input.MessageID)
			}

			return &notificationservice.Notification{
				ID:        "notification-1",
				UserID:    input.UserID,
				Type:      input.Type,
				Title:     input.Title,
				Body:      input.Body,
				RoomID:    input.RoomID,
				MessageID: input.MessageID,
				CreatedAt: now,
			}, nil
		},
	}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	handler, err := NewNotificationHandler(svc)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	res, err := handler.SendNotification(context.Background(), connect.NewRequest(&notificationv1.SendNotificationRequest{
		UserId:    "user-1",
		Type:      notificationv1.NotificationType_NOTIFICATION_TYPE_MESSAGE,
		Title:     "New message",
		Body:      "hello",
		RoomId:    "room-1",
		MessageId: "message-1",
	}))
	if err != nil {
		t.Fatalf("send notification: %v", err)
	}

	if res.Msg.GetId() != "notification-1" {
		t.Fatalf("expected notification-1, got %q", res.Msg.GetId())
	}

	if res.Msg.GetMessageId() != "message-1" {
		t.Fatalf("expected message-1, got %q", res.Msg.GetMessageId())
	}
}

func TestGetNotificationsPassesPaginationFields(t *testing.T) {
	svc, err := notificationservice.NewNotificationService(&fakeNotificationRepository{
		listByUserIDWithUnreadCountFunc: func(ctx context.Context, input notificationservice.ListNotificationsInput) ([]notificationservice.Notification, int, error) {
			if input.UserID != "user-1" {
				t.Fatalf("expected user-1, got %q", input.UserID)
			}

			if input.Limit != 20 {
				t.Fatalf("expected limit 20, got %d", input.Limit)
			}

			if input.Offset != 40 {
				t.Fatalf("expected offset 40, got %d", input.Offset)
			}

			if !input.UnreadOnly {
				t.Fatal("expected unread only")
			}

			return []notificationservice.Notification{
				{
					ID:        "notification-1",
					UserID:    "user-1",
					Type:      notificationservice.NotificationTypeMessage,
					Title:     "New message",
					Body:      "hello",
					MessageID: "message-1",
				},
			}, 1, nil
		},
	}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	handler, err := NewNotificationHandler(svc)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	res, err := handler.GetNotifications(context.Background(), connect.NewRequest(&notificationv1.GetNotificationsRequest{
		UserId:     "user-1",
		Limit:      20,
		Offset:     40,
		UnreadOnly: true,
	}))
	if err != nil {
		t.Fatalf("get notifications: %v", err)
	}

	if res.Msg.GetUnreadCount() != 1 {
		t.Fatalf("expected unread count 1, got %d", res.Msg.GetUnreadCount())
	}

	if len(res.Msg.GetNotifications()) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(res.Msg.GetNotifications()))
	}

	if res.Msg.GetNotifications()[0].GetMessageId() != "message-1" {
		t.Fatalf("expected message-1, got %q", res.Msg.GetNotifications()[0].GetMessageId())
	}
}

func TestMarkAsReadPassesRequestFields(t *testing.T) {
	called := false

	svc, err := notificationservice.NewNotificationService(&fakeNotificationRepository{
		markAsReadFunc: func(ctx context.Context, notificationID, userID string) error {
			called = true

			if notificationID != "notification-1" {
				t.Fatalf("expected notification-1, got %q", notificationID)
			}

			if userID != "user-1" {
				t.Fatalf("expected user-1, got %q", userID)
			}

			return nil
		},
	}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	handler, err := NewNotificationHandler(svc)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	_, err = handler.MarkAsRead(context.Background(), connect.NewRequest(&notificationv1.MarkAsReadRequest{
		NotificationId: "notification-1",
		UserId:         "user-1",
	}))
	if err != nil {
		t.Fatalf("mark as read: %v", err)
	}

	if !called {
		t.Fatal("expected mark as read to be called")
	}
}

func TestMarkAsReadMapsNotFoundToConnectNotFound(t *testing.T) {
	svc, err := notificationservice.NewNotificationService(&fakeNotificationRepository{
		markAsReadFunc: func(ctx context.Context, notificationID, userID string) error {
			return notificationservice.ErrNotFound
		},
	}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	handler, err := NewNotificationHandler(svc)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	_, err = handler.MarkAsRead(context.Background(), connect.NewRequest(&notificationv1.MarkAsReadRequest{
		NotificationId: "missing-notification",
		UserId:         "user-1",
	}))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		t.Fatalf("expected connect error, got %T", err)
	}

	if connectErr.Code() != connect.CodeNotFound {
		t.Fatalf("expected not found code, got %v", connectErr.Code())
	}
}
