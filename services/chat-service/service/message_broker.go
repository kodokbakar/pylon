package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

const DefaultMessageBrokerBuffer = 64

type MessageBroker interface {
	Subscribe(ctx context.Context, roomID string) (<-chan *Message, error)
	Publish(roomID string, msg *Message) int
}

type InMemoryMessageBroker struct {
	mu          sync.RWMutex
	buffer      int
	subscribers map[string]map[chan *Message]struct{}
}

func NewInMemoryMessageBroker(buffer int) *InMemoryMessageBroker {
	if buffer <= 0 {
		buffer = DefaultMessageBrokerBuffer
	}

	return &InMemoryMessageBroker{
		buffer:      buffer,
		subscribers: make(map[string]map[chan *Message]struct{}),
	}
}

func (b *InMemoryMessageBroker) Subscribe(ctx context.Context, roomID string) (<-chan *Message, error) {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return nil, fmt.Errorf("%w: room id is required", ErrInvalidInput)
	}

	ch := make(chan *Message, b.buffer)

	b.mu.Lock()
	if b.subscribers[roomID] == nil {
		b.subscribers[roomID] = make(map[chan *Message]struct{})
	}
	b.subscribers[roomID][ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()

		b.mu.Lock()
		delete(b.subscribers[roomID], ch)
		if len(b.subscribers[roomID]) == 0 {
			delete(b.subscribers, roomID)
		}
		close(ch)
		b.mu.Unlock()
	}()

	return ch, nil
}

// Publish returns the number of subscribers that could not receive the message
// because their channel buffer was full.
func (b *InMemoryMessageBroker) Publish(roomID string, msg *Message) int {
	if msg == nil {
		return 0
	}

	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return 0
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	dropped := 0
	for ch := range b.subscribers[roomID] {
		select {
		case ch <- msg:
		default:
			dropped++
		}
	}

	return dropped
}
