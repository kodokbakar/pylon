package database

import (
	"context"
	"strings"
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

func TestNewRedisClientAppliesConfigBeforePing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewRedisClient(ctx, config.RedisConfig{
		URL:      "redis://localhost:6379",
		Password: "secret",
		DB:       2,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "ping redis") {
		t.Fatalf("expected ping redis error, got %v", err)
	}
}
