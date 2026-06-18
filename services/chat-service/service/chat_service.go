package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidInput = errors.New("invalid input")

const MaxMessageContentLength = 10_000

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeSystem MessageType = "system"
)

type Message struct {
	ID        string
	RoomID    string
	SenderID  string
	Content   string
	Type      MessageType
	CreatedAt time.Time
}

type SendMessageInput struct {
	RoomID   string
	SenderID string
	Content  string
	Type     MessageType
}

type CreateMessageInput struct {
	RoomID   string
	SenderID string
	Content  string
	Type     MessageType
}

type GetMessagesInput struct {
	RoomID   string
	Limit    int
	BeforeID string
}

type GetMessagesResult struct {
	Messages []Message
	HasMore  bool
}

type StreamMessagesInput struct {
	RoomID string
	UserID string
}

type MessageRepository interface {
	Create(ctx context.Context, input CreateMessageInput) (*Message, error)
	ListByRoom(ctx context.Context, input GetMessagesInput) (*GetMessagesResult, error)
}

type EventPublisher interface {
	PublishMessageCreated(ctx context.Context, msg *Message) error
}

type ChatService struct {
	repo      MessageRepository
	publisher EventPublisher
}

func NewChatService(repo MessageRepository, publisher EventPublisher) (*ChatService, error) {
	if repo == nil {
		return nil, fmt.Errorf("message repository is required")
	}

	if publisher == nil {
		return nil, fmt.Errorf("event publisher is required")
	}

	return &ChatService{
		repo:      repo,
		publisher: publisher,
	}, nil
}

func (s *ChatService) SendMessage(ctx context.Context, input SendMessageInput) (*Message, error) {
	if err := validateSendMessageInput(input); err != nil {
		return nil, err
	}

	msg, err := s.repo.Create(ctx, CreateMessageInput{
		RoomID:   strings.TrimSpace(input.RoomID),
		SenderID: strings.TrimSpace(input.SenderID),
		Content:  strings.TrimSpace(input.Content),
		Type:     normalizeMessageType(input.Type),
	})
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	if err := s.publisher.PublishMessageCreated(ctx, msg); err != nil {
		return nil, fmt.Errorf("publish message created event: %w", err)
	}

	return msg, nil
}

func (s *ChatService) GetMessages(ctx context.Context, input GetMessagesInput) (*GetMessagesResult, error) {
	input.RoomID = strings.TrimSpace(input.RoomID)
	input.BeforeID = strings.TrimSpace(input.BeforeID)
	input.Limit = normalizeLimit(input.Limit)

	if input.RoomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	result, err := s.repo.ListByRoom(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("list messages by room: %w", err)
	}

	return result, nil
}

func (s *ChatService) StreamMessages(ctx context.Context, input StreamMessagesInput) (<-chan *Message, error) {
	input.RoomID = strings.TrimSpace(input.RoomID)
	input.UserID = strings.TrimSpace(input.UserID)

	if input.RoomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	if input.UserID == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	messages := make(chan *Message)

	go func() {
		defer close(messages)
		<-ctx.Done()
	}()

	return messages, nil
}

func validateSendMessageInput(input SendMessageInput) error {
	if strings.TrimSpace(input.RoomID) == "" {
		return fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	if strings.TrimSpace(input.SenderID) == "" {
		return fmt.Errorf("%w: sender id is required", ErrInvalidInput)
	}

	content := strings.TrimSpace(input.Content)
	if content == "" {
		return fmt.Errorf("%w: content is required", ErrInvalidInput)
	}

	if len([]rune(content)) > MaxMessageContentLength {
		return fmt.Errorf("%w: content exceeds maximum length of %d characters", ErrInvalidInput, MaxMessageContentLength)
	}

	if !isValidMessageType(input.Type) {
		return fmt.Errorf("%w: unsupported message type %q", ErrInvalidInput, input.Type)
	}

	return nil
}

func normalizeMessageType(messageType MessageType) MessageType {
	if messageType == "" {
		return MessageTypeText
	}

	return messageType
}

func isValidMessageType(messageType MessageType) bool {
	switch normalizeMessageType(messageType) {
	case MessageTypeText, MessageTypeImage, MessageTypeSystem:
		return true
	default:
		return false
	}
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}

	if limit > 100 {
		return 100
	}

	return limit
}
