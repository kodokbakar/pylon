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
	"github.com/kodokbakar/pylon/internal/metrics"
	notificationservice "github.com/kodokbakar/pylon/services/notification-service/service"
)

const (
	MessageEventsTopic          = "message-events"
	NotificationConsumerGroupID = "notification-consumer-group"
	MessageCreatedEventType     = "message_created"
	maxNotificationBodyRunes    = 100
)

type MessageCreatedEvent struct {
	EventID   string                  `json:"event_id"`
	EventType string                  `json:"event_type"`
	Timestamp time.Time               `json:"timestamp"`
	Data      MessageCreatedEventData `json:"data"`
}

type MessageCreatedEventData struct {
	MessageID      string    `json:"message_id"`
	RoomID         string    `json:"room_id"`
	SenderID       string    `json:"sender_id"`
	SenderUsername string    `json:"sender_username"`
	Content        string    `json:"content"`
	Type           string    `json:"type"`
	CreatedAt      time.Time `json:"created_at"`
}

type RoomMembersClient interface {
	GetRoomMembers(context.Context, *connect.Request[roomv1.GetRoomMembersRequest]) (*connect.Response[roomv1.GetRoomMembersResponse], error)
}

type NotificationSender interface {
	SendNotification(ctx context.Context, input notificationservice.SendNotificationInput) (*notificationservice.Notification, error)
}

type messageReader interface {
	FetchMessage(ctx context.Context) (kafkago.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafkago.Message) error
	Close() error
}

type Consumer struct {
	reader          messageReader
	roomClient      RoomMembersClient
	notificationSvc NotificationSender
	topic           string
	consumerGroup   string
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	roomClient RoomMembersClient,
	notificationSvc NotificationSender,
) (*Consumer, error) {
	if roomClient == nil {
		return nil, fmt.Errorf("room client is required")
	}

	if notificationSvc == nil {
		return nil, fmt.Errorf("notification service is required")
	}

	readerConfig, err := newReaderConfig(brokers, topic, groupID)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		reader:          kafkago.NewReader(readerConfig),
		roomClient:      roomClient,
		notificationSvc: notificationSvc,
		topic:           readerConfig.Topic,
		consumerGroup:   readerConfig.GroupID,
	}, nil
}

func newReaderConfig(brokers []string, topic, groupID string) (kafkago.ReaderConfig, error) {
	if len(brokers) == 0 {
		return kafkago.ReaderConfig{}, fmt.Errorf("kafka brokers are required")
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return kafkago.ReaderConfig{}, fmt.Errorf("kafka topic is required")
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		groupID = NotificationConsumerGroupID
	}

	return kafkago.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafkago.FirstOffset,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	if c == nil || c.reader == nil {
		return fmt.Errorf("kafka reader is required")
	}

	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				return nil
			}

			log.Printf("failed to fetch message event: %v", err)
			continue
		}

		if err := c.HandleMessage(ctx, message); err != nil {
			log.Printf("handle kafka message: %v", err)
			continue
		}

		metrics.RecordKafkaMessageConsumed(c.metricTopic(), c.metricConsumerGroup())

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			log.Printf("failed to commit message event: %v", err)
		}
	}
}

func (c *Consumer) HandleMessage(ctx context.Context, message kafkago.Message) error {
	if c == nil {
		return fmt.Errorf("consumer is required")
	}

	if c.roomClient == nil {
		return fmt.Errorf("room client is required")
	}

	if c.notificationSvc == nil {
		return fmt.Errorf("notification service is required")
	}

	event, err := DecodeMessageCreatedEvent(message.Value)
	if err != nil {
		return err
	}

	membersRes, err := c.roomClient.GetRoomMembers(ctx, connect.NewRequest(&roomv1.GetRoomMembersRequest{
		RoomId: event.Data.RoomID,
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

func (c *Consumer) metricTopic() string {
	if c == nil {
		return MessageEventsTopic
	}

	topic := strings.TrimSpace(c.topic)
	if topic == "" {
		return MessageEventsTopic
	}

	return topic
}

func (c *Consumer) metricConsumerGroup() string {
	if c == nil {
		return NotificationConsumerGroupID
	}

	consumerGroup := strings.TrimSpace(c.consumerGroup)
	if consumerGroup == "" {
		return NotificationConsumerGroupID
	}

	return consumerGroup
}

func DecodeMessageCreatedEvent(payload []byte) (*MessageCreatedEvent, error) {
	var event MessageCreatedEvent

	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, fmt.Errorf("decode message created event: %w", err)
	}

	if event.EventType != MessageCreatedEventType {
		return nil, fmt.Errorf("unsupported message event type %q", event.EventType)
	}

	if strings.TrimSpace(event.EventID) == "" {
		return nil, fmt.Errorf("message event id is required")
	}

	if strings.TrimSpace(event.Data.RoomID) == "" {
		return nil, fmt.Errorf("message event room id is required")
	}

	if strings.TrimSpace(event.Data.SenderID) == "" {
		return nil, fmt.Errorf("message event sender id is required")
	}

	if strings.TrimSpace(event.Data.MessageID) == "" {
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
		if userID == "" || userID == event.Data.SenderID {
			continue
		}

		notifications = append(notifications, notificationservice.SendNotificationInput{
			UserID:    userID,
			Type:      notificationservice.NotificationTypeMessage,
			Title:     messageNotificationTitle(event),
			Body:      truncateRunes(event.Data.Content, maxNotificationBodyRunes),
			RoomID:    event.Data.RoomID,
			MessageID: event.Data.MessageID,
		})
	}

	return notifications
}

func messageNotificationTitle(event *MessageCreatedEvent) string {
	senderName := strings.TrimSpace(event.Data.SenderUsername)
	if senderName == "" {
		senderName = event.Data.SenderID
	}

	return fmt.Sprintf("New message from %s", senderName)
}

func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit])
}
