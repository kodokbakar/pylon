package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	kafkago "github.com/segmentio/kafka-go"

	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	notificationservice "github.com/kodokbakar/pylon/services/notification-service/service"
)

type fakeMessageReader struct {
	messages  []kafkago.Message
	fetchErr  error
	commitErr error
	closeErr  error
	commits   []kafkago.Message
	closed    bool
	index     int
}

func (r *fakeMessageReader) FetchMessage(ctx context.Context) (kafkago.Message, error) {
	if r.fetchErr != nil {
		return kafkago.Message{}, r.fetchErr
	}

	if r.index >= len(r.messages) {
		return kafkago.Message{}, context.Canceled
	}

	message := r.messages[r.index]
	r.index++

	return message, nil
}

func (r *fakeMessageReader) CommitMessages(ctx context.Context, messages ...kafkago.Message) error {
	if r.commitErr != nil {
		return r.commitErr
	}

	r.commits = append(r.commits, messages...)
	return nil
}

func (r *fakeMessageReader) Close() error {
	r.closed = true
	return r.closeErr
}

func TestNewConsumerRequiresBrokers(t *testing.T) {
	_, err := NewConsumer(nil, MessageEventsTopic, NotificationConsumerGroupID, fakeRoomClient{}, fakeNotificationSender{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewConsumerRequiresRoomClient(t *testing.T) {
	_, err := NewConsumer([]string{"localhost:9092"}, MessageEventsTopic, NotificationConsumerGroupID, nil, fakeNotificationSender{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewReaderConfigMatchesContract(t *testing.T) {
	cfg, err := newReaderConfig([]string{"localhost:9092"}, MessageEventsTopic, "")
	if err != nil {
		t.Fatalf("create reader config: %v", err)
	}

	if cfg.Topic != MessageEventsTopic {
		t.Fatalf("expected topic %q, got %q", MessageEventsTopic, cfg.Topic)
	}

	if cfg.GroupID != NotificationConsumerGroupID {
		t.Fatalf("expected group id %q, got %q", NotificationConsumerGroupID, cfg.GroupID)
	}

	if cfg.StartOffset != kafkago.FirstOffset {
		t.Fatalf("expected first offset, got %d", cfg.StartOffset)
	}

	if cfg.MinBytes != 1 {
		t.Fatalf("expected min bytes 1, got %d", cfg.MinBytes)
	}

	if cfg.MaxBytes != 10e6 {
		t.Fatalf("expected max bytes 10e6, got %d", cfg.MaxBytes)
	}
}

func TestDecodeMessageCreatedEvent(t *testing.T) {
	createdAt := time.Now().UTC()

	payload, err := json.Marshal(MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Timestamp: time.Now().UTC(),
		Data: MessageCreatedEventData{
			MessageID:      "message-1",
			RoomID:         "room-1",
			SenderID:       "user-1",
			SenderUsername: "alice",
			Content:        "hello",
			Type:           "text",
			CreatedAt:      createdAt,
		},
	})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	event, err := DecodeMessageCreatedEvent(payload)
	if err != nil {
		t.Fatalf("decode event: %v", err)
	}

	if event.EventType != MessageCreatedEventType {
		t.Fatalf("expected event type %q, got %q", MessageCreatedEventType, event.EventType)
	}

	if event.Data.RoomID != "room-1" {
		t.Fatalf("expected room-1, got %q", event.Data.RoomID)
	}

	if event.Data.SenderID != "user-1" {
		t.Fatalf("expected user-1, got %q", event.Data.SenderID)
	}

	if event.Data.SenderUsername != "alice" {
		t.Fatalf("expected alice, got %q", event.Data.SenderUsername)
	}

	if !event.Data.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, event.Data.CreatedAt)
	}
}

func TestDecodeMessageCreatedEventRejectsUnsupportedType(t *testing.T) {
	payload, err := json.Marshal(MessageCreatedEvent{
		EventID:   "event-1",
		EventType: "unknown",
		Data: MessageCreatedEventData{
			RoomID:    "room-1",
			SenderID:  "user-1",
			MessageID: "message-1",
		},
	})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	_, err = DecodeMessageCreatedEvent(payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildMessageNotificationInputsSkipsSenderAndTruncatesBody(t *testing.T) {
	event := &MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Data: MessageCreatedEventData{
			RoomID:         "room-1",
			MessageID:      "message-1",
			SenderID:       "user-1",
			SenderUsername: "alice",
			Content:        strings.Repeat("a", 120),
		},
	}

	notifications := BuildMessageNotificationInputs(event, []*roomv1.RoomMember{
		{UserId: "user-1", RoomId: "room-1"},
		{UserId: "user-2", RoomId: "room-1"},
		{UserId: "user-3", RoomId: "room-1"},
	})

	if len(notifications) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notifications))
	}

	for _, notification := range notifications {
		if notification.UserID == event.Data.SenderID {
			t.Fatalf("sender should not receive notification: %+v", notification)
		}

		if notification.MessageID != "message-1" {
			t.Fatalf("expected message id message-1, got %q", notification.MessageID)
		}

		if notification.Type != notificationservice.NotificationTypeMessage {
			t.Fatalf("expected message notification type, got %q", notification.Type)
		}

		if notification.Title != "New message from alice" {
			t.Fatalf("expected title with sender username, got %q", notification.Title)
		}

		if len([]rune(notification.Body)) != maxNotificationBodyRunes {
			t.Fatalf("expected body to be truncated to %d runes, got %d", maxNotificationBodyRunes, len([]rune(notification.Body)))
		}
	}
}

func TestHandleMessageCreatesNotificationsForRoomMembersExceptSender(t *testing.T) {
	payload := mustMessageCreatedPayload(t, MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Timestamp: time.Now().UTC(),
		Data: MessageCreatedEventData{
			MessageID:      "message-1",
			RoomID:         "room-1",
			SenderID:       "user-1",
			SenderUsername: "alice",
			Content:        "hello",
			Type:           "text",
			CreatedAt:      time.Now().UTC(),
		},
	})

	sentTo := make([]string, 0)

	consumer := &Consumer{
		roomClient: fakeRoomClient{
			getRoomMembersFunc: func(
				ctx context.Context,
				req *connect.Request[roomv1.GetRoomMembersRequest],
			) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
				if req.Msg.GetRoomId() != "room-1" {
					t.Fatalf("expected room-1, got %q", req.Msg.GetRoomId())
				}

				return connect.NewResponse(&roomv1.GetRoomMembersResponse{
					Members: []*roomv1.RoomMember{
						{UserId: "user-1", RoomId: "room-1"},
						{UserId: "user-2", RoomId: "room-1"},
						{UserId: "user-3", RoomId: "room-1"},
					},
				}), nil
			},
		},
		notificationSvc: fakeNotificationSender{
			sendNotificationFunc: func(
				ctx context.Context,
				input notificationservice.SendNotificationInput,
			) (*notificationservice.Notification, error) {
				sentTo = append(sentTo, input.UserID)

				if input.RoomID != "room-1" {
					t.Fatalf("expected room-1, got %q", input.RoomID)
				}

				if input.MessageID != "message-1" {
					t.Fatalf("expected message id message-1, got %q", input.MessageID)
				}

				if input.Type != notificationservice.NotificationTypeMessage {
					t.Fatalf("expected message notification type, got %q", input.Type)
				}

				if input.Title != "New message from alice" {
					t.Fatalf("expected title with sender username, got %q", input.Title)
				}

				return &notificationservice.Notification{
					ID:     "notification-" + input.UserID,
					UserID: input.UserID,
				}, nil
			},
		},
	}

	err := consumer.HandleMessage(context.Background(), kafkago.Message{
		Value: payload,
	})
	if err != nil {
		t.Fatalf("handle message: %v", err)
	}

	if len(sentTo) != 2 {
		t.Fatalf("expected 2 notifications, got %d: %+v", len(sentTo), sentTo)
	}

	if sentTo[0] != "user-2" || sentTo[1] != "user-3" {
		t.Fatalf("expected notifications for user-2 and user-3, got %+v", sentTo)
	}
}

func TestStartCommitsMessageAfterSuccessfulHandling(t *testing.T) {
	payload := mustMessageCreatedPayload(t, MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Timestamp: time.Now().UTC(),
		Data: MessageCreatedEventData{
			MessageID:      "message-1",
			RoomID:         "room-1",
			SenderID:       "user-1",
			SenderUsername: "alice",
			Content:        "hello",
			Type:           "text",
			CreatedAt:      time.Now().UTC(),
		},
	})

	reader := &fakeMessageReader{
		messages: []kafkago.Message{
			{
				Topic:     MessageEventsTopic,
				Partition: 0,
				Offset:    10,
				Value:     payload,
			},
		},
	}

	consumer := &Consumer{
		reader: reader,
		roomClient: fakeRoomClient{
			getRoomMembersFunc: func(
				ctx context.Context,
				req *connect.Request[roomv1.GetRoomMembersRequest],
			) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
				return connect.NewResponse(&roomv1.GetRoomMembersResponse{
					Members: []*roomv1.RoomMember{
						{UserId: "user-1", RoomId: "room-1"},
						{UserId: "user-2", RoomId: "room-1"},
					},
				}), nil
			},
		},
		notificationSvc: fakeNotificationSender{
			sendNotificationFunc: func(
				ctx context.Context,
				input notificationservice.SendNotificationInput,
			) (*notificationservice.Notification, error) {
				return &notificationservice.Notification{
					ID:     "notification-" + input.UserID,
					UserID: input.UserID,
				}, nil
			},
		},
	}

	if err := consumer.Start(context.Background()); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	if len(reader.commits) != 1 {
		t.Fatalf("expected 1 committed message, got %d", len(reader.commits))
	}

	if reader.commits[0].Offset != 10 {
		t.Fatalf("expected committed offset 10, got %d", reader.commits[0].Offset)
	}
}

func TestStartDoesNotCommitWhenHandleFails(t *testing.T) {
	reader := &fakeMessageReader{
		messages: []kafkago.Message{
			{
				Topic:     MessageEventsTopic,
				Partition: 0,
				Offset:    10,
				Value:     []byte(`{"event_type":"unknown"}`),
			},
		},
	}

	consumer := &Consumer{
		reader:          reader,
		roomClient:      fakeRoomClient{},
		notificationSvc: fakeNotificationSender{},
	}

	if err := consumer.Start(context.Background()); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	if len(reader.commits) != 0 {
		t.Fatalf("expected no committed messages, got %d", len(reader.commits))
	}
}

func TestCloseClosesReader(t *testing.T) {
	reader := &fakeMessageReader{}
	consumer := &Consumer{reader: reader}

	if err := consumer.Close(); err != nil {
		t.Fatalf("close consumer: %v", err)
	}

	if !reader.closed {
		t.Fatal("expected reader to be closed")
	}
}

func mustMessageCreatedPayload(t *testing.T, event MessageCreatedEvent) []byte {
	t.Helper()

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	return payload
}

type fakeRoomClient struct {
	getRoomMembersFunc func(
		ctx context.Context,
		req *connect.Request[roomv1.GetRoomMembersRequest],
	) (*connect.Response[roomv1.GetRoomMembersResponse], error)
}

type fakeNotificationSender struct {
	sendNotificationFunc func(
		ctx context.Context,
		input notificationservice.SendNotificationInput,
	) (*notificationservice.Notification, error)
}

func (c fakeRoomClient) GetRoomMembers(
	ctx context.Context,
	req *connect.Request[roomv1.GetRoomMembersRequest],
) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
	if c.getRoomMembersFunc == nil {
		return connect.NewResponse(&roomv1.GetRoomMembersResponse{}), nil
	}

	return c.getRoomMembersFunc(ctx, req)
}

func (s fakeNotificationSender) SendNotification(
	ctx context.Context,
	input notificationservice.SendNotificationInput,
) (*notificationservice.Notification, error) {
	if s.sendNotificationFunc == nil {
		return &notificationservice.Notification{}, nil
	}

	return s.sendNotificationFunc(ctx, input)
}

var errTest = errors.New("test error")

func TestStartContinuesWhenCommitFails(t *testing.T) {
	payload := mustMessageCreatedPayload(t, MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Timestamp: time.Now().UTC(),
		Data: MessageCreatedEventData{
			MessageID:      "message-1",
			RoomID:         "room-1",
			SenderID:       "user-1",
			SenderUsername: "alice",
			Content:        "hello",
			Type:           "text",
			CreatedAt:      time.Now().UTC(),
		},
	})

	reader := &fakeMessageReader{
		messages:  []kafkago.Message{{Value: payload}},
		commitErr: errTest,
	}

	consumer := &Consumer{
		reader: reader,
		roomClient: fakeRoomClient{
			getRoomMembersFunc: func(
				ctx context.Context,
				req *connect.Request[roomv1.GetRoomMembersRequest],
			) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
				return connect.NewResponse(&roomv1.GetRoomMembersResponse{
					Members: []*roomv1.RoomMember{
						{UserId: "user-1", RoomId: "room-1"},
					},
				}), nil
			},
		},
		notificationSvc: fakeNotificationSender{},
	}

	if err := consumer.Start(context.Background()); err != nil {
		t.Fatalf("expected graceful exit after commit failure and shutdown, got %v", err)
	}
}

func TestNewConsumerRequiresNotificationService(t *testing.T) {
	_, err := NewConsumer([]string{"localhost:9092"}, MessageEventsTopic, NotificationConsumerGroupID, fakeRoomClient{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewReaderConfigRejectsEmptyTopic(t *testing.T) {
	_, err := newReaderConfig([]string{"localhost:9092"}, " ", NotificationConsumerGroupID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStartRequiresReader(t *testing.T) {
	var consumer *Consumer
	if err := consumer.Start(context.Background()); err == nil {
		t.Fatal("expected nil consumer error")
	}

	consumer = &Consumer{}
	if err := consumer.Start(context.Background()); err == nil {
		t.Fatal("expected missing reader error")
	}
}

func TestCloseHandlesNilConsumerAndReader(t *testing.T) {
	var consumer *Consumer
	if err := consumer.Close(); err != nil {
		t.Fatalf("expected nil consumer close to be no-op, got %v", err)
	}

	consumer = &Consumer{}
	if err := consumer.Close(); err != nil {
		t.Fatalf("expected missing reader close to be no-op, got %v", err)
	}
}

func TestCloseWrapsReaderError(t *testing.T) {
	consumer := &Consumer{
		reader: &fakeMessageReader{closeErr: errTest},
	}

	err := consumer.Close()
	if err == nil {
		t.Fatal("expected close error, got nil")
	}

	if !strings.Contains(err.Error(), "close kafka reader") {
		t.Fatalf("expected wrapped close error, got %v", err)
	}
}

func TestDecodeMessageCreatedEventRejectsInvalidJSON(t *testing.T) {
	_, err := DecodeMessageCreatedEvent([]byte(`{invalid-json`))
	if err == nil {
		t.Fatal("expected decode error, got nil")
	}
}

func TestDecodeMessageCreatedEventRequiresFields(t *testing.T) {
	tests := []struct {
		name    string
		event   MessageCreatedEvent
		wantErr string
	}{
		{
			name: "event id",
			event: MessageCreatedEvent{
				EventType: MessageCreatedEventType,
				Data: MessageCreatedEventData{
					RoomID:    "room-1",
					SenderID:  "user-1",
					MessageID: "message-1",
				},
			},
			wantErr: "message event id is required",
		},
		{
			name: "room id",
			event: MessageCreatedEvent{
				EventID:   "event-1",
				EventType: MessageCreatedEventType,
				Data: MessageCreatedEventData{
					SenderID:  "user-1",
					MessageID: "message-1",
				},
			},
			wantErr: "message event room id is required",
		},
		{
			name: "sender id",
			event: MessageCreatedEvent{
				EventID:   "event-1",
				EventType: MessageCreatedEventType,
				Data: MessageCreatedEventData{
					RoomID:    "room-1",
					MessageID: "message-1",
				},
			},
			wantErr: "message event sender id is required",
		},
		{
			name: "message id",
			event: MessageCreatedEvent{
				EventID:   "event-1",
				EventType: MessageCreatedEventType,
				Data: MessageCreatedEventData{
					RoomID:   "room-1",
					SenderID: "user-1",
				},
			},
			wantErr: "message event message id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeMessageCreatedEvent(mustMessageCreatedPayload(t, tt.event))
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestHandleMessageRequiresDependencies(t *testing.T) {
	payload := mustMessageCreatedPayload(t, MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Data: MessageCreatedEventData{
			MessageID: "message-1",
			RoomID:    "room-1",
			SenderID:  "user-1",
		},
	})

	var consumer *Consumer
	if err := consumer.HandleMessage(context.Background(), kafkago.Message{Value: payload}); err == nil {
		t.Fatal("expected nil consumer error")
	}

	consumer = &Consumer{notificationSvc: fakeNotificationSender{}}
	if err := consumer.HandleMessage(context.Background(), kafkago.Message{Value: payload}); err == nil {
		t.Fatal("expected missing room client error")
	}

	consumer = &Consumer{roomClient: fakeRoomClient{}}
	if err := consumer.HandleMessage(context.Background(), kafkago.Message{Value: payload}); err == nil {
		t.Fatal("expected missing notification service error")
	}
}

func TestHandleMessageWrapsRoomClientError(t *testing.T) {
	payload := mustMessageCreatedPayload(t, MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Data: MessageCreatedEventData{
			MessageID: "message-1",
			RoomID:    "room-1",
			SenderID:  "user-1",
		},
	})

	consumer := &Consumer{
		roomClient: fakeRoomClient{
			getRoomMembersFunc: func(
				ctx context.Context,
				req *connect.Request[roomv1.GetRoomMembersRequest],
			) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
				return nil, errTest
			},
		},
		notificationSvc: fakeNotificationSender{},
	}

	err := consumer.HandleMessage(context.Background(), kafkago.Message{Value: payload})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "get room members") {
		t.Fatalf("expected wrapped room client error, got %v", err)
	}
}

func TestHandleMessageWrapsNotificationError(t *testing.T) {
	payload := mustMessageCreatedPayload(t, MessageCreatedEvent{
		EventID:   "event-1",
		EventType: MessageCreatedEventType,
		Data: MessageCreatedEventData{
			MessageID:      "message-1",
			RoomID:         "room-1",
			SenderID:       "user-1",
			SenderUsername: "alice",
			Content:        "hello",
		},
	})

	consumer := &Consumer{
		roomClient: fakeRoomClient{
			getRoomMembersFunc: func(
				ctx context.Context,
				req *connect.Request[roomv1.GetRoomMembersRequest],
			) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
				return connect.NewResponse(&roomv1.GetRoomMembersResponse{
					Members: []*roomv1.RoomMember{
						{UserId: "user-1", RoomId: "room-1"},
						{UserId: "user-2", RoomId: "room-1"},
					},
				}), nil
			},
		},
		notificationSvc: fakeNotificationSender{
			sendNotificationFunc: func(
				ctx context.Context,
				input notificationservice.SendNotificationInput,
			) (*notificationservice.Notification, error) {
				return nil, errTest
			},
		},
	}

	err := consumer.HandleMessage(context.Background(), kafkago.Message{Value: payload})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "send notification to user user-2") {
		t.Fatalf("expected wrapped notification error, got %v", err)
	}
}
