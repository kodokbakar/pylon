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

type NotificationType string

const (
	NotificationTypeMessage NotificationType = "message"
	NotificationTypeInvite  NotificationType = "invite"
	NotificationTypeMention NotificationType = "mention"
)

const (
	DefaultNotificationLimit = 20
	MaxNotificationLimit     = 100
	MaxNotificationTitleLen  = 255
	MaxNotificationBodyLen   = 10000
)

type Notification struct {
	ID        string
	UserID    string
	Type      NotificationType
	Title     string
	Body      string
	RoomID    string
	MessageID string
	Read      bool
	CreatedAt time.Time
}

type SendNotificationInput struct {
	UserID    string
	Type      NotificationType
	Title     string
	Body      string
	RoomID    string
	MessageID string
}

type GetNotificationsInput struct {
	UserID     string
	Limit      int
	Offset     int
	UnreadOnly bool
}

type GetNotificationsResult struct {
	Notifications []Notification
	UnreadCount   int
}

type MarkAsReadInput struct {
	NotificationID string
	UserID         string
}

type CreateNotificationInput struct {
	UserID    string
	Type      NotificationType
	Title     string
	Body      string
	RoomID    string
	MessageID string
}

type ListNotificationsInput struct {
	UserID     string
	Limit      int
	Offset     int
	UnreadOnly bool
}

type NotificationRepository interface {
	Create(ctx context.Context, input CreateNotificationInput) (*Notification, error)
	ListByUserID(ctx context.Context, input ListNotificationsInput) ([]Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	MarkAsRead(ctx context.Context, notificationID, userID string) error
}

type NotificationPusher interface {
	PushNotification(ctx context.Context, notification *Notification) error
}

type NoopPusher struct{}

func (NoopPusher) PushNotification(ctx context.Context, notification *Notification) error {
	return nil
}

type NotificationService struct {
	repo   NotificationRepository
	pusher NotificationPusher
}

func NewNotificationService(repo NotificationRepository, pusher NotificationPusher) (*NotificationService, error) {
	if repo == nil {
		return nil, fmt.Errorf("notification repository is required")
	}

	if pusher == nil {
		pusher = NoopPusher{}
	}

	return &NotificationService{
		repo:   repo,
		pusher: pusher,
	}, nil
}

func (s *NotificationService) SendNotification(ctx context.Context, input SendNotificationInput) (*Notification, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	input.RoomID = strings.TrimSpace(input.RoomID)
	input.MessageID = strings.TrimSpace(input.MessageID)

	if err := validateSendNotificationInput(input); err != nil {
		return nil, err
	}

	notification, err := s.repo.Create(ctx, CreateNotificationInput(input))
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	if err := s.pusher.PushNotification(ctx, notification); err != nil {
		// Notification is already persisted. Push is best-effort so callers do not
		// retry and create duplicate notifications.
		return notification, nil
	}

	return notification, nil
}

func (s *NotificationService) GetNotifications(ctx context.Context, input GetNotificationsInput) (*GetNotificationsResult, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	limit := normalizeLimit(input.Limit)
	offset := normalizeOffset(input.Offset)

	notifications, err := s.repo.ListByUserID(ctx, ListNotificationsInput{
		UserID:     userID,
		Limit:      limit,
		Offset:     offset,
		UnreadOnly: input.UnreadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	unreadCount, err := s.repo.CountUnread(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count unread notifications: %w", err)
	}

	return &GetNotificationsResult{
		Notifications: notifications,
		UnreadCount:   unreadCount,
	}, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, input MarkAsReadInput) error {
	notificationID := strings.TrimSpace(input.NotificationID)
	if notificationID == "" {
		return fmt.Errorf("%w: notification id is required", ErrInvalidInput)
	}

	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	if err := s.repo.MarkAsRead(ctx, notificationID, userID); err != nil {
		return fmt.Errorf("mark notification as read: %w", err)
	}

	return nil
}

func validateSendNotificationInput(input SendNotificationInput) error {
	if input.UserID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	if !isValidNotificationType(input.Type) {
		return fmt.Errorf("%w: unsupported notification type %q", ErrInvalidInput, input.Type)
	}

	if input.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if len([]rune(input.Title)) > MaxNotificationTitleLen {
		return fmt.Errorf("%w: title exceeds maximum length of %d characters", ErrInvalidInput, MaxNotificationTitleLen)
	}

	if input.Body == "" {
		return fmt.Errorf("%w: body is required", ErrInvalidInput)
	}

	if len([]rune(input.Body)) > MaxNotificationBodyLen {
		return fmt.Errorf("%w: body exceeds maximum length of %d characters", ErrInvalidInput, MaxNotificationBodyLen)
	}

	return nil
}

func isValidNotificationType(notificationType NotificationType) bool {
	switch notificationType {
	case NotificationTypeMessage, NotificationTypeInvite, NotificationTypeMention:
		return true
	default:
		return false
	}
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultNotificationLimit
	}

	if limit > MaxNotificationLimit {
		return MaxNotificationLimit
	}

	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}

	return offset
}
