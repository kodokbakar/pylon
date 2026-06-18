package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"connectrpc.com/connect"
	kafkago "github.com/segmentio/kafka-go"

	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	notificationservice "github.com/kodokbakar/pylon/services/notification-service/service"
)

const MessageEventsTopic = "message-events"

const defaultConsumerGroupID = "pylon-notification-service"

type MessageCreatedEvent struct {
	Version   string    `json:"version"`
	EventID   string    `json:"event_id"`
	Type      string    `json:"type"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	MessageID string    `json:"message_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type RoomMembersClient interface {
	GetRoomMembers(context.Context, *connect.Request[roomv1.GetRoomMembersRequest]) (*connect.Response[roomv1.GetRoomMembersResponse], error)
}

type NotificationSender interface {
	SendNotification(ctx context.Context, input notificationservice.SendNotificationInput) (*notificationservice.Notification, error)
}

type Consumer struct {
	reader          *kafkago.Reader
	roomClient      RoomMembersClient
	notificationSvc NotificationSender
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	roomClient RoomMembersClient,
	notificationSvc NotificationSender,
) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		groupID = defaultConsumerGroupID
	}

	if roomClient == nil {
		return nil, fmt.Errorf("room client is required")
	}

	if notificationSvc == nil {
		return nil, fmt.Errorf("notification service is required")
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:          reader,
		roomClient:      roomClient,
		notificationSvc: notificationSvc,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return fmt.Errorf("read message event: %w", err)
		}

		if err := c.HandleMessage(ctx, message); err != nil {
			log.Printf("handle message event: %v", err)
			continue
		}
	}
}

func (c *Consumer) HandleMessage(ctx context.Context, message kafkago.Message) error {
	event, err := DecodeMessageCreatedEvent(message.Value)
	if err != nil {
		return err
	}

	membersRes, err := c.roomClient.GetRoomMembers(ctx, connect.NewRequest(&roomv1.GetRoomMembersRequest{
		RoomId: event.RoomID,
	}))
	if err != nil {
		return fmt.Errorf("get room members: %w", err)
	}

	notifications := BuildMessageNotificationInputs(event, membersRes.Msg.GetMembers())
	for _, notification := range notifications {
		if _, err := c.notificationSvc.SendNotification(ctx, notification); err != nil {
			return fmt.Errorf("send notification to user %s: %w", notification.UserID, err)
		}
	}

	return nil
}

func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}

	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("close kafka reader: %w", err)
	}

	return nil
}

func DecodeMessageCreatedEvent(payload []byte) (*MessageCreatedEvent, error) {
	var event MessageCreatedEvent

	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode message created event: %w", err)
	}

	if event.Type != "message.created" {
		return nil, fmt.Errorf("unsupported message event type %q", event.Type)
	}

	if strings.TrimSpace(event.RoomID) == "" {
		return nil, fmt.Errorf("message event room id is required")
	}

	if strings.TrimSpace(event.SenderID) == "" {
		return nil, fmt.Errorf("message event sender id is required")
	}

	if strings.TrimSpace(event.MessageID) == "" {
		return nil, fmt.Errorf("message event message id is required")
	}

	return &event, nil
}

func BuildMessageNotificationInputs(event *MessageCreatedEvent, members []*roomv1.RoomMember) []notificationservice.SendNotificationInput {
	if event == nil {
		return nil
	}

	notifications := make([]notificationservice.SendNotificationInput, 0, len(members))
	for _, member := range members {
		if member == nil {
			continue
		}

		userID := strings.TrimSpace(member.GetUserId())
		if userID == "" || userID == event.SenderID {
			continue
		}

		notifications = append(notifications, notificationservice.SendNotificationInput{
			UserID: userID,
			Type:   notificationservice.NotificationTypeMessage,
			Title:  "New message",
			Body:   event.Content,
			RoomID: event.RoomID,
		})
	}

	return notifications
}
