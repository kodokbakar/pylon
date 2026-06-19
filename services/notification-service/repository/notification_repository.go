package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	notificationservice "github.com/kodokbakar/pylon/services/notification-service/service"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

const createNotificationQuery = `
	INSERT INTO notifications (user_id, type, title, body, room_id, message_id)
	VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid, NULLIF($6, '')::uuid)
	ON CONFLICT (user_id, message_id) WHERE message_id IS NOT NULL
	DO UPDATE SET title = notifications.title
	RETURNING id::text, user_id::text, type, title, body, COALESCE(room_id::text, ''), COALESCE(message_id::text, ''), read, created_at
`

const listNotificationsByUserIDQuery = `
	SELECT id::text, user_id::text, type, title, body, COALESCE(room_id::text, ''), COALESCE(message_id::text, ''), read, created_at	FROM notifications
	WHERE user_id = $1
	  AND ($2::boolean = false OR read = false)
	ORDER BY created_at DESC, id DESC
	LIMIT $3
`

const countUnreadNotificationsQuery = `
	SELECT COUNT(*)
	FROM notifications
	WHERE user_id = $1
	  AND read = false
`

const markNotificationAsReadQuery = `
	UPDATE notifications
	SET read = true
	WHERE id = $1
	  AND user_id = $2
`

func NewNotificationRepository(db *pgxpool.Pool) (*NotificationRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres pool is required")
	}

	return &NotificationRepository{db: db}, nil
}

func (r *NotificationRepository) Create(ctx context.Context, input notificationservice.CreateNotificationInput) (*notificationservice.Notification, error) {
	notification, err := scanNotification(r.db.QueryRow(
		ctx,
		createNotificationQuery,
		input.UserID,
		string(input.Type),
		input.Title,
		input.Body,
		input.RoomID,
		input.MessageID,
	))
	if err != nil {
		return nil, fmt.Errorf("insert notification: %w", err)
	}

	return notification, nil
}

func (r *NotificationRepository) ListByUserID(ctx context.Context, input notificationservice.ListNotificationsInput) ([]notificationservice.Notification, error) {
	rows, err := r.db.Query(ctx, listNotificationsByUserIDQuery, input.UserID, input.UnreadOnly, input.Limit)
	if err != nil {
		return nil, fmt.Errorf("query notifications by user id: %w", err)
	}
	defer rows.Close()

	notifications := make([]notificationservice.Notification, 0)
	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}

		notifications = append(notifications, *notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int

	if err := r.db.QueryRow(ctx, countUnreadNotificationsQuery, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}

	return count, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	tag, err := r.db.Exec(ctx, markNotificationAsReadQuery, notificationID, userID)
	if err != nil {
		return fmt.Errorf("update notification read status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: notification %s for user %s", notificationservice.ErrNotFound, notificationID, userID)
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(row scanner) (*notificationservice.Notification, error) {
	var notification notificationservice.Notification
	var notificationType string

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notificationType,
		&notification.Title,
		&notification.Body,
		&notification.RoomID,
		&notification.MessageID,
		&notification.Read,
		&notification.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: notification", notificationservice.ErrNotFound)
	}
	if err != nil {
		return nil, err
	}

	notification.Type = notificationservice.NotificationType(notificationType)

	return &notification, nil
}
