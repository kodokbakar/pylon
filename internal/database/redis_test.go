package database

import (
	"context"
	"testing"

	"github.com/kodokbakar/pylon/internal/config"
)

func TestNewRedisClientReturnsErrorForMissingURL(t *testing.T) {
	_, err := NewRedisClient(context.Background(), config.RedisConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewRedisClientReturnsErrorForInvalidURL(t *testing.T) {
	_, err := NewRedisClient(context.Background(), config.RedisConfig{
		URL: "not-a-redis-url",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
