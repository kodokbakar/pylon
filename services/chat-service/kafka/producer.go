package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	chatservice "github.com/kodokbakar/pylon/services/chat-service/service"
)

const MessageEventsTopic = "message-events"

const MessageCreatedEventVersion = "1.0"

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

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic, clientID string) (*Producer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers are required")
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return nil, fmt.Errorf("kafka topic is required")
	}

	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		clientID = "pylon-chat-service"
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		Transport: &kafka.Transport{
			ClientID: clientID,
		},
	}

	return &Producer{writer: writer}, nil
}

func NewMessageCreatedEvent(msg *chatservice.Message) (*MessageCreatedEvent, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is required")
	}

	return &MessageCreatedEvent{
		Version:   MessageCreatedEventVersion,
		EventID:   fmt.Sprintf("message.created.%s", msg.ID),
		Type:      "message.created",
		RoomID:    msg.RoomID,
		SenderID:  msg.SenderID,
		MessageID: msg.ID,
		Content:   msg.Content,
		Timestamp: msg.CreatedAt,
	}, nil
}

func (p *Producer) PublishMessageCreated(ctx context.Context, msg *chatservice.Message) error {
	event, err := NewMessageCreatedEvent(msg)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal message created event: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.RoomID),
		Value: payload,
		Time:  time.Now(),
	}); err != nil {
		return fmt.Errorf("write message created event: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}

	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("close kafka writer: %w", err)
	}

	return nil
}
