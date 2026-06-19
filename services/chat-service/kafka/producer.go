package kafka

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/kodokbakar/pylon/internal/metrics"
	chatservice "github.com/kodokbakar/pylon/services/chat-service/service"
)

const (
	MessageEventsTopic       = "message-events"
	MessageCreatedEventType  = "message_created"
	defaultProducerClientID  = "pylon-chat-service"
	defaultProducerBatchSize = 100
	defaultBatchTimeout      = 10 * time.Millisecond
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

type messageWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Producer struct {
	writer messageWriter
	topic  string
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
		clientID = defaultProducerClientID
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		BatchTimeout:           defaultBatchTimeout,
		BatchSize:              defaultProducerBatchSize,
		RequiredAcks:           kafka.RequireAll,
		Async:                  false,
		AllowAutoTopicCreation: true,
		Transport: &kafka.Transport{
			ClientID: clientID,
		},
	}

	return &Producer{
		writer: writer,
		topic:  topic,
	}, nil
}

func NewMessageCreatedEvent(msg *chatservice.Message) (*MessageCreatedEvent, error) {
	if msg == nil {
		return nil, fmt.Errorf("message is required")
	}

	eventID, err := newEventID()
	if err != nil {
		return nil, fmt.Errorf("generate event id: %w", err)
	}

	return &MessageCreatedEvent{
		EventID:   eventID,
		EventType: MessageCreatedEventType,
		Timestamp: time.Now().UTC(),
		Data: MessageCreatedEventData{
			MessageID:      msg.ID,
			RoomID:         msg.RoomID,
			SenderID:       msg.SenderID,
			SenderUsername: msg.SenderUsername,
			Content:        msg.Content,
			Type:           string(msg.Type),
			CreatedAt:      msg.CreatedAt.UTC(),
		},
	}, nil
}

func (p *Producer) PublishMessageCreated(ctx context.Context, msg *chatservice.Message) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("kafka writer is required")
	}

	event, err := NewMessageCreatedEvent(msg)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal message created event: %w", err)
	}

	kafkaMessage := kafka.Message{
		Key:   []byte(event.Data.RoomID),
		Value: payload,
		Time:  event.Timestamp,
		Headers: []kafka.Header{
			{
				Key:   "event_type",
				Value: []byte(MessageCreatedEventType),
			},
		},
	}

	topic := strings.TrimSpace(p.topic)
	if topic == "" {
		topic = MessageEventsTopic
	}

	startedAt := time.Now()
	if err := p.writer.WriteMessages(ctx, kafkaMessage); err != nil {
		metrics.ObserveKafkaPublish(topic, time.Since(startedAt))
		return fmt.Errorf("write message created event: %w", err)
	}

	metrics.ObserveKafkaPublish(topic, time.Since(startedAt))
	metrics.RecordKafkaMessagePublished(topic)

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

func newEventID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	), nil
}
