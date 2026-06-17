package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeMessageRepository struct {
	createFunc     func(ctx context.Context, input CreateMessageInput) (*Message, error)
	listByRoomFunc func(ctx context.Context, input GetMessagesInput) (*GetMessagesResult, error)
}

func (r *fakeMessageRepository) Create(ctx context.Context, input CreateMessageInput) (*Message, error) {
	if r.createFunc == nil {
		return nil, errors.New("create func is not configured")
	}

	return r.createFunc(ctx, input)
}

func (r *fakeMessageRepository) ListByRoom(ctx context.Context, input GetMessagesInput) (*GetMessagesResult, error) {
	if r.listByRoomFunc == nil {
		return nil, errors.New("list by room func is not configured")
	}

	return r.listByRoomFunc(ctx, input)
}

type fakeEventPublisher struct {
	publishFunc func(ctx context.Context, msg *Message) error
}

func (p *fakeEventPublisher) PublishMessageCreated(ctx context.Context, msg *Message) error {
	if p.publishFunc == nil {
		return errors.New("publish func is not configured")
	}

	return p.publishFunc(ctx, msg)
}

func TestNewChatServiceRequiresRepository(t *testing.T) {
	_, err := NewChatService(nil, &fakeEventPublisher{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewChatServiceRequiresPublisher(t *testing.T) {
	_, err := NewChatService(&fakeMessageRepository{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSendMessageCreatesMessageAndPublishesEvent(t *testing.T) {
	published := false

	svc, err := NewChatService(
		&fakeMessageRepository{
			createFunc: func(ctx context.Context, input CreateMessageInput) (*Message, error) {
				if input.RoomID != "room-1" {
					t.Fatalf("expected room-1, got %q", input.RoomID)
				}

				if input.Type != MessageTypeText {
					t.Fatalf("expected text message type, got %q", input.Type)
				}

				return &Message{
					ID:        "message-1",
					RoomID:    input.RoomID,
					SenderID:  input.SenderID,
					Content:   input.Content,
					Type:      input.Type,
					CreatedAt: time.Now(),
				}, nil
			},
		},
		&fakeEventPublisher{
			publishFunc: func(ctx context.Context, msg *Message) error {
				published = true

				if msg.ID != "message-1" {
					t.Fatalf("expected message-1, got %q", msg.ID)
				}

				return nil
			},
		},
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	msg, err := svc.SendMessage(context.Background(), SendMessageInput{
		RoomID:   " room-1 ",
		SenderID: "user-1",
		Content:  "hello",
	})
	if err != nil {
		t.Fatalf("send message: %v", err)
	}

	if msg.ID != "message-1" {
		t.Fatalf("expected message-1, got %q", msg.ID)
	}

	if !published {
		t.Fatal("expected message created event to be published")
	}
}

func TestSendMessageRejectsEmptyContent(t *testing.T) {
	svc, err := NewChatService(&fakeMessageRepository{}, &fakeEventPublisher{})
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	_, err = svc.SendMessage(context.Background(), SendMessageInput{
		RoomID:   "room-1",
		SenderID: "user-1",
		Content:  " ",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSendMessageRejectsContentOverLimit(t *testing.T) {
	svc, err := NewChatService(&fakeMessageRepository{}, &fakeEventPublisher{})
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	tooLongContent := strings.Repeat("a", MaxMessageContentLength+1)

	_, err = svc.SendMessage(context.Background(), SendMessageInput{
		RoomID:   "room-1",
		SenderID: "user-1",
		Content:  tooLongContent,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestGetMessagesNormalizesLimit(t *testing.T) {
	svc, err := NewChatService(
		&fakeMessageRepository{
			listByRoomFunc: func(ctx context.Context, input GetMessagesInput) (*GetMessagesResult, error) {
				if input.Limit != 100 {
					t.Fatalf("expected limit capped to 100, got %d", input.Limit)
				}

				return &GetMessagesResult{}, nil
			},
		},
		&fakeEventPublisher{},
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	_, err = svc.GetMessages(context.Background(), GetMessagesInput{
		RoomID: "room-1",
		Limit:  500,
	})
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
}

func TestStreamMessagesRequiresRoomID(t *testing.T) {
	svc, err := NewChatService(&fakeMessageRepository{}, &fakeEventPublisher{})
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	_, err = svc.StreamMessages(context.Background(), StreamMessagesInput{
		UserID: "user-1",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

func TestSendMessageReturnsErrorWhenPublishFails(t *testing.T) {
	publishErr := errors.New("publish failed")

	svc, err := NewChatService(
		&fakeMessageRepository{
			createFunc: func(ctx context.Context, input CreateMessageInput) (*Message, error) {
				return &Message{
					ID:        "message-1",
					RoomID:    input.RoomID,
					SenderID:  input.SenderID,
					Content:   input.Content,
					Type:      input.Type,
					CreatedAt: time.Now(),
				}, nil
			},
		},
		&fakeEventPublisher{
			publishFunc: func(ctx context.Context, msg *Message) error {
				return publishErr
			},
		},
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	_, err = svc.SendMessage(context.Background(), SendMessageInput{
		RoomID:   "room-1",
		SenderID: "user-1",
		Content:  "hello",
	})
	if !errors.Is(err, publishErr) {
		t.Fatalf("expected publish error, got %v", err)
	}
}
