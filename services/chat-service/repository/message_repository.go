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

const listMessagesByRoomQuery = `
	SELECT id, room_id, sender_id, content, type, created_at
	FROM messages
	WHERE room_id = $1
	ORDER BY created_at DESC, id DESC
	LIMIT $2
`

const listMessagesByRoomBeforeIDQuery = `
	WITH cursor_message AS (
		SELECT id, created_at
		FROM messages
		WHERE room_id = $1
		  AND id = $2
	)
	SELECT m.id, m.room_id, m.sender_id, m.content, m.type, m.created_at
	FROM messages m
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
	var msg chatservice.Message
	var messageType string

	err := r.db.QueryRow(ctx, `
		INSERT INTO messages (room_id, sender_id, content, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, sender_id, content, type, created_at
	`, input.RoomID, input.SenderID, input.Content, string(input.Type)).Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.SenderID,
		&msg.Content,
		&messageType,
		&msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert message: %w", err)
	}

	msg.Type = chatservice.MessageType(messageType)

	return &msg, nil
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
		var msg chatservice.Message
		var messageType string

		if err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.SenderID,
			&msg.Content,
			&messageType,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		msg.Type = chatservice.MessageType(messageType)
		messages = append(messages, msg)
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

type pgxRows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}
