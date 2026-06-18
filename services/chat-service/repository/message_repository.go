package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	chatservice "github.com/kodokbakar/pylon/services/chat-service/service"
)

type MessageRepository struct {
	db *pgxpool.Pool
}

const createMessageQuery = `
	WITH inserted AS (
		INSERT INTO messages (room_id, sender_id, content, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, sender_id, content, type, created_at
	)
	SELECT
		i.id,
		i.room_id,
		i.sender_id,
		i.content,
		i.type,
		i.created_at,
		u.username,
		COALESCE(u.display_name, ''),
		COALESCE(u.avatar_url, '')
	FROM inserted i
	JOIN users u ON u.id = i.sender_id
`

const listMessagesByRoomQuery = `
	SELECT
		m.id,
		m.room_id,
		m.sender_id,
		m.content,
		m.type,
		m.created_at,
		u.username,
		COALESCE(u.display_name, ''),
		COALESCE(u.avatar_url, '')
	FROM messages m
	JOIN users u ON u.id = m.sender_id
	WHERE m.room_id = $1
	ORDER BY m.created_at DESC, m.id DESC
	LIMIT $2
`

const listMessagesByRoomBeforeIDQuery = `
	WITH cursor_message AS (
		SELECT id, created_at
		FROM messages
		WHERE room_id = $1
		  AND id = $2
	)
	SELECT
		m.id,
		m.room_id,
		m.sender_id,
		m.content,
		m.type,
		m.created_at,
		u.username,
		COALESCE(u.display_name, ''),
		COALESCE(u.avatar_url, '')
	FROM messages m
	JOIN users u ON u.id = m.sender_id
	JOIN cursor_message c ON true
	WHERE m.room_id = $1
	  AND (
		m.created_at < c.created_at
		OR (m.created_at = c.created_at AND m.id < c.id)
	  )
	ORDER BY m.created_at DESC, m.id DESC
	LIMIT $3
`

func NewMessageRepository(db *pgxpool.Pool) (*MessageRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres pool is required")
	}

	return &MessageRepository{db: db}, nil
}

func (r *MessageRepository) Create(ctx context.Context, input chatservice.CreateMessageInput) (*chatservice.Message, error) {
	msg, err := scanMessage(r.db.QueryRow(
		ctx,
		createMessageQuery,
		input.RoomID,
		input.SenderID,
		input.Content,
		string(input.Type),
	))
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	return msg, nil
}

func (r *MessageRepository) ListByRoom(ctx context.Context, input chatservice.GetMessagesInput) (*chatservice.GetMessagesResult, error) {
	limit := input.Limit + 1

	var rows pgxRows
	var err error

	if input.BeforeID == "" {
		rows, err = r.db.Query(ctx, listMessagesByRoomQuery, input.RoomID, limit)
	} else {
		rows, err = r.db.Query(ctx, listMessagesByRoomBeforeIDQuery, input.RoomID, input.BeforeID, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer rows.Close()

	messages := make([]chatservice.Message, 0, input.Limit+1)

	for rows.Next() {
		msg, err := scanMessage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		messages = append(messages, *msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	hasMore := len(messages) > input.Limit
	if hasMore {
		messages = messages[:input.Limit]
	}

	return &chatservice.GetMessagesResult{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}

type pgxScanner interface {
	Scan(dest ...any) error
}

func scanMessage(row pgxScanner) (*chatservice.Message, error) {
	var msg chatservice.Message
	var messageType string

	if err := row.Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.SenderID,
		&msg.Content,
		&messageType,
		&msg.CreatedAt,
		&msg.SenderUsername,
		&msg.SenderDisplayName,
		&msg.SenderAvatarURL,
	); err != nil {
		return nil, err
	}

	msg.Type = chatservice.MessageType(messageType)

	return &msg, nil
}

type pgxRows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}
