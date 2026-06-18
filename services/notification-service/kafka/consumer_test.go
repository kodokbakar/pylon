package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"connectrpc.com/connect"
	kafkago "github.com/segmentio/kafka-go"

	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	notificationservice "github.com/kodokbakar/pylon/services/notification-service/service"
)

func TestNewConsumerRequiresBrokers(t *testing.T) {
	_, err := NewConsumer(nil, MessageEventsTopic, "group", fakeRoomClient{}, fakeNotificationSender{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDecodeMessageCreatedEvent(t *testing.T) {
	payload, err := json.Marshal(MessageCreatedEvent{
		Version:   "1.0",
		EventID:   "message.created.message-1",
		Type:      "message.created",
		RoomID:    "room-1",
		SenderID:  "user-1",
		MessageID: "message-1",
		Content:   "hello",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	event, err := DecodeMessageCreatedEvent(payload)
	if err != nil {
		t.Fatalf("decode event: %v", err)
	}

	if event.RoomID != "room-1" {
		t.Fatalf("expected room-1, got %q", event.RoomID)
	}

	if event.SenderID != "user-1" {
		t.Fatalf("expected user-1, got %q", event.SenderID)
	}
}

func TestDecodeMessageCreatedEventRejectsUnsupportedType(t *testing.T) {
	payload, err := json.Marshal(MessageCreatedEvent{
		Type:      "unknown",
		RoomID:    "room-1",
		SenderID:  "user-1",
		MessageID: "message-1",
	})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	_, err = DecodeMessageCreatedEvent(payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildMessageNotificationInputsSkipsSender(t *testing.T) {
	event := &MessageCreatedEvent{
		Type:     "message.created",
		RoomID:   "room-1",
		SenderID: "user-1",
		Content:  "hello",
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
		if notification.UserID == event.SenderID {
			t.Fatalf("sender should not receive notification: %+v", notification)
		}

		if notification.Type != notificationservice.NotificationTypeMessage {
			t.Fatalf("expected message notification type, got %q", notification.Type)
		}
	}
}

func TestHandleMessageCreatesNotificationsForRoomMembersExceptSender(t *testing.T) {
	payload, err := json.Marshal(MessageCreatedEvent{
		Version:   "1.0",
		EventID:   "message.created.message-1",
		Type:      "message.created",
		RoomID:    "room-1",
		SenderID:  "user-1",
		MessageID: "message-1",
		Content:   "hello",
		Timestamp: time.Now(),
	})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

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

				if input.Type != notificationservice.NotificationTypeMessage {
					t.Fatalf("expected message notification type, got %q", input.Type)
				}

				return &notificationservice.Notification{
					ID:     "notification-" + input.UserID,
					UserID: input.UserID,
				}, nil
			},
		},
	}

	err = consumer.HandleMessage(context.Background(), kafkago.Message{
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
