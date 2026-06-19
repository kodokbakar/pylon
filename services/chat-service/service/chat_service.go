package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kodokbakar/pylon/internal/metrics"
)

var ErrInvalidInput = errors.New("invalid input")

var ErrForbidden = errors.New("forbidden")

const MaxMessageContentLength = 10_000

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeSystem MessageType = "system"
	MessageTypeFile   MessageType = "file"
)

type Message struct {
	ID                string
	RoomID            string
	SenderID          string
	SenderUsername    string
	SenderDisplayName string
	SenderAvatarURL   string
	Content           string
	Type              MessageType
	CreatedAt         time.Time
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
	UserID   string
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

type RoomMembershipChecker interface {
	IsMember(ctx context.Context, roomID, userID string) (bool, error)
}

type EventPublisher interface {
	PublishMessageCreated(ctx context.Context, msg *Message) error
}

type ChatServiceOption func(*ChatService) error

type ChatService struct {
	repo       MessageRepository
	membership RoomMembershipChecker
	publisher  EventPublisher
	broker     MessageBroker
}

func NewChatService(repo MessageRepository, publisher EventPublisher, options ...ChatServiceOption) (*ChatService, error) {
	if repo == nil {
		return nil, fmt.Errorf("message repository is required")
	}

	if publisher == nil {
		return nil, fmt.Errorf("event publisher is required")
	}

	svc := &ChatService{
		repo:      repo,
		publisher: publisher,
		broker:    NewInMemoryMessageBroker(DefaultMessageBrokerBuffer),
	}

	for _, option := range options {
		if option == nil {
			continue
		}

		if err := option(svc); err != nil {
			return nil, err
		}
	}

	return svc, nil
}

func WithRoomMembershipChecker(checker RoomMembershipChecker) ChatServiceOption {
	return func(s *ChatService) error {
		if checker == nil {
			return fmt.Errorf("room membership checker is required")
		}

		s.membership = checker
		return nil
	}
}

func WithMessageBroker(broker MessageBroker) ChatServiceOption {
	return func(s *ChatService) error {
		if broker == nil {
			return fmt.Errorf("message broker is required")
		}

		s.broker = broker
		return nil
	}
}

func (s *ChatService) SendMessage(ctx context.Context, input SendMessageInput) (*Message, error) {
	if err := validateSendMessageInput(input); err != nil {
		return nil, err
	}

	input.RoomID = strings.TrimSpace(input.RoomID)
	input.SenderID = strings.TrimSpace(input.SenderID)
	input.Content = strings.TrimSpace(input.Content)
	input.Type = normalizeMessageType(input.Type)

	if err := s.ensureRoomMember(ctx, input.RoomID, input.SenderID); err != nil {
		return nil, err
	}

	msg, err := s.repo.Create(ctx, CreateMessageInput(input))
	if err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	if err := s.publisher.PublishMessageCreated(ctx, msg); err != nil {
		log.Printf("publish message created event failed: %v", err)
	}

	metrics.RecordMessageSent("unknown", string(msg.Type))

	if dropped := s.broker.Publish(msg.RoomID, msg); dropped > 0 {
		log.Printf("dropped %d chat stream messages for room %s", dropped, msg.RoomID)
	}

	return msg, nil
}

func (s *ChatService) GetMessages(ctx context.Context, input GetMessagesInput) (*GetMessagesResult, error) {
	input.RoomID = strings.TrimSpace(input.RoomID)
	input.UserID = strings.TrimSpace(input.UserID)
	input.BeforeID = strings.TrimSpace(input.BeforeID)
	input.Limit = normalizeLimit(input.Limit)

	if input.RoomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	if input.UserID == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}

	if err := s.ensureRoomMember(ctx, input.RoomID, input.UserID); err != nil {
		return nil, err
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

	if err := s.ensureRoomMember(ctx, input.RoomID, input.UserID); err != nil {
		return nil, err
	}

	messages, err := s.broker.Subscribe(ctx, input.RoomID)
	if err != nil {
		return nil, fmt.Errorf("subscribe room messages: %w", err)
	}

	return messages, nil
}

func (s *ChatService) ensureRoomMember(ctx context.Context, roomID, userID string) error {
	if s.membership == nil {
		return fmt.Errorf("%w: room membership checker is not configured", ErrForbidden)
	}

	isMember, err := s.membership.IsMember(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("check room membership: %w", err)
	}

	if !isMember {
		return fmt.Errorf("%w: user is not a room member", ErrForbidden)
	}

	return nil
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
	case MessageTypeText, MessageTypeImage, MessageTypeSystem, MessageTypeFile:
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
