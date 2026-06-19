package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeNotificationRepository struct {
	createFunc       func(ctx context.Context, input CreateNotificationInput) (*Notification, error)
	listByUserIDFunc func(ctx context.Context, input ListNotificationsInput) ([]Notification, error)
	countUnreadFunc  func(ctx context.Context, userID string) (int, error)
	markAsReadFunc   func(ctx context.Context, notificationID, userID string) error
}

func (r *fakeNotificationRepository) Create(ctx context.Context, input CreateNotificationInput) (*Notification, error) {
	if r.createFunc == nil {
		return nil, errors.New("create func is not configured")
	}

	return r.createFunc(ctx, input)
}

func (r *fakeNotificationRepository) ListByUserID(ctx context.Context, input ListNotificationsInput) ([]Notification, error) {
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

func (r *fakeNotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	if r.markAsReadFunc == nil {
		return errors.New("mark as read func is not configured")
	}

	return r.markAsReadFunc(ctx, notificationID, userID)
}

type fakeNotificationPusher struct {
	pushFunc func(ctx context.Context, notification *Notification) error
}

func (p *fakeNotificationPusher) PushNotification(ctx context.Context, notification *Notification) error {
	if p.pushFunc == nil {
		return nil
	}

	return p.pushFunc(ctx, notification)
}

func TestNewNotificationServiceRequiresRepository(t *testing.T) {
	_, err := NewNotificationService(nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSendNotificationValidatesUserID(t *testing.T) {
	svc, err := NewNotificationService(&fakeNotificationRepository{}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	_, err = svc.SendNotification(context.Background(), SendNotificationInput{
		Type:  NotificationTypeMessage,
		Title: "New message",
		Body:  "hello",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSendNotificationRejectsLongTitle(t *testing.T) {
	svc, err := NewNotificationService(&fakeNotificationRepository{}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	_, err = svc.SendNotification(context.Background(), SendNotificationInput{
		UserID: "user-1",
		Type:   NotificationTypeMessage,
		Title:  strings.Repeat("a", MaxNotificationTitleLen+1),
		Body:   "hello",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSendNotificationRejectsLongBody(t *testing.T) {
	svc, err := NewNotificationService(&fakeNotificationRepository{}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	_, err = svc.SendNotification(context.Background(), SendNotificationInput{
		UserID: "user-1",
		Type:   NotificationTypeMessage,
		Title:  "New message",
		Body:   strings.Repeat("a", MaxNotificationBodyLen+1),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSendNotificationCreatesAndPushesNotification(t *testing.T) {
	now := time.Now()
	pushed := false

	svc, err := NewNotificationService(
		&fakeNotificationRepository{
			createFunc: func(ctx context.Context, input CreateNotificationInput) (*Notification, error) {
				if input.UserID != "user-1" {
					t.Fatalf("expected user-1, got %q", input.UserID)
				}

				if input.MessageID != "message-1" {
					t.Fatalf("expected message-1, got %q", input.MessageID)
				}

				if input.Type != NotificationTypeMessage {
					t.Fatalf("expected message type, got %q", input.Type)
				}

				return &Notification{
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
		},
		&fakeNotificationPusher{
			pushFunc: func(ctx context.Context, notification *Notification) error {
				pushed = true

				if notification.ID != "notification-1" {
					t.Fatalf("expected notification-1, got %q", notification.ID)
				}

				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	notification, err := svc.SendNotification(context.Background(), SendNotificationInput{
		UserID:    " user-1 ",
		Type:      NotificationTypeMessage,
		Title:     " New message ",
		Body:      " hello ",
		RoomID:    " room-1 ",
		MessageID: " message-1 ",
	})
	if err != nil {
		t.Fatalf("send notification: %v", err)
	}

	if notification.ID != "notification-1" {
		t.Fatalf("expected notification-1, got %q", notification.ID)
	}

	if notification.MessageID != "message-1" {
		t.Fatalf("expected message-1, got %q", notification.MessageID)
	}

	if !pushed {
		t.Fatal("expected notification to be pushed")
	}
}

func TestSendNotificationReturnsSuccessWhenPushFailsAfterPersist(t *testing.T) {
	svc, err := NewNotificationService(
		&fakeNotificationRepository{
			createFunc: func(ctx context.Context, input CreateNotificationInput) (*Notification, error) {
				return &Notification{
					ID:     "notification-1",
					UserID: input.UserID,
					Type:   input.Type,
					Title:  input.Title,
					Body:   input.Body,
					RoomID: input.RoomID,
				}, nil
			},
		},
		&fakeNotificationPusher{
			pushFunc: func(ctx context.Context, notification *Notification) error {
				return errors.New("websocket unavailable")
			},
		},
	)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	notification, err := svc.SendNotification(context.Background(), SendNotificationInput{
		UserID: "user-1",
		Type:   NotificationTypeMessage,
		Title:  "New message",
		Body:   "hello",
	})
	if err != nil {
		t.Fatalf("expected push failure to be best-effort, got %v", err)
	}

	if notification.ID != "notification-1" {
		t.Fatalf("expected notification-1, got %q", notification.ID)
	}
}

func TestGetNotificationsValidatesUserID(t *testing.T) {
	svc, err := NewNotificationService(&fakeNotificationRepository{}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	_, err = svc.GetNotifications(context.Background(), GetNotificationsInput{})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestGetNotificationsReturnsRepositoryValues(t *testing.T) {
	svc, err := NewNotificationService(
		&fakeNotificationRepository{
			listByUserIDFunc: func(ctx context.Context, input ListNotificationsInput) ([]Notification, error) {
				if input.UserID != "user-1" {
					t.Fatalf("expected user-1, got %q", input.UserID)
				}

				if input.Limit != DefaultNotificationLimit {
					t.Fatalf("expected default limit %d, got %d", DefaultNotificationLimit, input.Limit)
				}

				if input.Offset != 10 {
					t.Fatalf("expected offset 10, got %d", input.Offset)
				}

				return []Notification{
					{
						ID:     "notification-1",
						UserID: input.UserID,
						Type:   NotificationTypeMessage,
						Title:  "New message",
						Body:   "hello",
					},
				}, nil
			},
			countUnreadFunc: func(ctx context.Context, userID string) (int, error) {
				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}

				return 1, nil
			},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	result, err := svc.GetNotifications(context.Background(), GetNotificationsInput{
		UserID: " user-1 ",
		Offset: 10,
	})
	if err != nil {
		t.Fatalf("get notifications: %v", err)
	}

	if len(result.Notifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(result.Notifications))
	}

	if result.UnreadCount != 1 {
		t.Fatalf("expected unread count 1, got %d", result.UnreadCount)
	}
}

func TestMarkAsReadValidatesUserID(t *testing.T) {
	svc, err := NewNotificationService(&fakeNotificationRepository{}, nil)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	err = svc.MarkAsRead(context.Background(), MarkAsReadInput{
		NotificationID: "notification-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestMarkAsReadCallsRepository(t *testing.T) {
	called := false

	svc, err := NewNotificationService(
		&fakeNotificationRepository{
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
		},
		nil,
	)
	if err != nil {
		t.Fatalf("create notification service: %v", err)
	}

	err = svc.MarkAsRead(context.Background(), MarkAsReadInput{
		NotificationID: " notification-1 ",
		UserID:         " user-1 ",
	})
	if err != nil {
		t.Fatalf("mark as read: %v", err)
	}

	if !called {
		t.Fatal("expected repository to be called")
	}
}

func TestNormalizeLimitCapsLargeLimit(t *testing.T) {
	if got := normalizeLimit(MaxNotificationLimit + 1); got != MaxNotificationLimit {
		t.Fatalf("expected max limit %d, got %d", MaxNotificationLimit, got)
	}
}

func TestNormalizeOffsetRejectsNegativeValue(t *testing.T) {
	if got := normalizeOffset(-1); got != 0 {
		t.Fatalf("expected offset 0, got %d", got)
	}
}
