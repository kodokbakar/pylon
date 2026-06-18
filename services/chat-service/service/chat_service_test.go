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

type fakeRoomMembershipChecker struct {
	isMemberFunc func(ctx context.Context, roomID, userID string) (bool, error)
}

func (c *fakeRoomMembershipChecker) IsMember(ctx context.Context, roomID, userID string) (bool, error) {
	if c.isMemberFunc == nil {
		return false, errors.New("is member func is not configured")
	}

	return c.isMemberFunc(ctx, roomID, userID)
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
					ID:             "message-1",
					RoomID:         input.RoomID,
					SenderID:       input.SenderID,
					SenderUsername: "alice",
					Content:        input.Content,
					Type:           input.Type,
					CreatedAt:      time.Now(),
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
		WithRoomMembershipChecker(&fakeRoomMembershipChecker{
			isMemberFunc: func(ctx context.Context, roomID, userID string) (bool, error) {
				if roomID != "room-1" {
					t.Fatalf("expected room-1, got %q", roomID)
				}

				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}

				return true, nil
			},
		}),
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
				if input.RoomID != "room-1" {
					t.Fatalf("expected room-1, got %q", input.RoomID)
				}

				if input.UserID != "user-1" {
					t.Fatalf("expected user-1, got %q", input.UserID)
				}

				if input.Limit != 100 {
					t.Fatalf("expected limit capped to 100, got %d", input.Limit)
				}

				return &GetMessagesResult{}, nil
			},
		},
		&fakeEventPublisher{},
		WithRoomMembershipChecker(&fakeRoomMembershipChecker{
			isMemberFunc: func(ctx context.Context, roomID, userID string) (bool, error) {
				if roomID != "room-1" {
					t.Fatalf("expected room-1, got %q", roomID)
				}

				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}

				return true, nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	_, err = svc.GetMessages(context.Background(), GetMessagesInput{
		RoomID: " room-1 ",
		UserID: " user-1 ",
		Limit:  500,
	})
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
}

func TestGetMessagesRejectsNonMember(t *testing.T) {
	svc, err := NewChatService(
		&fakeMessageRepository{},
		&fakeEventPublisher{},
		WithRoomMembershipChecker(&fakeRoomMembershipChecker{
			isMemberFunc: func(ctx context.Context, roomID, userID string) (bool, error) {
				return false, nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	_, err = svc.GetMessages(context.Background(), GetMessagesInput{
		RoomID: "room-1",
		UserID: "user-1",
		Limit:  50,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestInMemoryMessageBrokerReportsDroppedMessages(t *testing.T) {
	broker := NewInMemoryMessageBroker(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages, err := broker.Subscribe(ctx, "room-1")
	if err != nil {
		t.Fatalf("subscribe room messages: %v", err)
	}

	first := &Message{ID: "message-1", RoomID: "room-1"}
	second := &Message{ID: "message-2", RoomID: "room-1"}

	if dropped := broker.Publish("room-1", first); dropped != 0 {
		t.Fatalf("expected no dropped messages, got %d", dropped)
	}

	if dropped := broker.Publish("room-1", second); dropped != 1 {
		t.Fatalf("expected 1 dropped message, got %d", dropped)
	}

	got := <-messages
	if got.ID != first.ID {
		t.Fatalf("expected first message %q, got %q", first.ID, got.ID)
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

func TestSendMessageReturnsMessageWhenPublishFails(t *testing.T) {
	publishErr := errors.New("publish failed")

	svc, err := NewChatService(
		&fakeMessageRepository{
			createFunc: func(ctx context.Context, input CreateMessageInput) (*Message, error) {
				return &Message{
					ID:             "message-1",
					RoomID:         input.RoomID,
					SenderID:       input.SenderID,
					SenderUsername: "alice",
					Content:        input.Content,
					Type:           input.Type,
					CreatedAt:      time.Now(),
				}, nil
			},
		},
		&fakeEventPublisher{
			publishFunc: func(ctx context.Context, msg *Message) error {
				return publishErr
			},
		},
		WithRoomMembershipChecker(&fakeRoomMembershipChecker{
			isMemberFunc: func(ctx context.Context, roomID, userID string) (bool, error) {
				return true, nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	msg, err := svc.SendMessage(context.Background(), SendMessageInput{
		RoomID:   "room-1",
		SenderID: "user-1",
		Content:  "hello",
	})
	if err != nil {
		t.Fatalf("expected message to succeed when publish fails, got %v", err)
	}

	if msg.ID != "message-1" {
		t.Fatalf("expected message-1, got %q", msg.ID)
	}
}

func TestStreamMessagesReceivesPublishedMessage(t *testing.T) {
	broker := NewInMemoryMessageBroker(1)

	svc, err := NewChatService(
		&fakeMessageRepository{},
		&fakeEventPublisher{},
		WithMessageBroker(broker),
		WithRoomMembershipChecker(&fakeRoomMembershipChecker{
			isMemberFunc: func(ctx context.Context, roomID, userID string) (bool, error) {
				if roomID != "room-1" {
					t.Fatalf("expected room-1, got %q", roomID)
				}

				if userID != "user-1" {
					t.Fatalf("expected user-1, got %q", userID)
				}

				return true, nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages, err := svc.StreamMessages(ctx, StreamMessagesInput{
		RoomID: " room-1 ",
		UserID: " user-1 ",
	})
	if err != nil {
		t.Fatalf("stream messages: %v", err)
	}

	expected := &Message{
		ID:       "message-1",
		RoomID:   "room-1",
		SenderID: "user-1",
		Content:  "hello",
		Type:     MessageTypeText,
	}

	broker.Publish("room-1", expected)

	select {
	case got := <-messages:
		if got.ID != expected.ID {
			t.Fatalf("expected message id %q, got %q", expected.ID, got.ID)
		}
	case <-time.After(time.Second):
		t.Fatal("expected streamed message")
	}
}

func TestSendMessageRejectsNonMember(t *testing.T) {
	svc, err := NewChatService(
		&fakeMessageRepository{},
		&fakeEventPublisher{},
		WithRoomMembershipChecker(&fakeRoomMembershipChecker{
			isMemberFunc: func(ctx context.Context, roomID, userID string) (bool, error) {
				return false, nil
			},
		}),
	)
	if err != nil {
		t.Fatalf("create chat service: %v", err)
	}

	_, err = svc.SendMessage(context.Background(), SendMessageInput{
		RoomID:   "room-1",
		SenderID: "user-1",
		Content:  "hello",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}
