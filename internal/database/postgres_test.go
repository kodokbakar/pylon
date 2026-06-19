package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kodokbakar/pylon/internal/config"
)

func TestNewPostgresPoolReturnsErrorForMissingURL(t *testing.T) {
	_, err := NewPostgresPool(context.Background(), config.DatabaseConfig{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewPostgresPoolReturnsErrorForInvalidURL(t *testing.T) {
	_, err := NewPostgresPool(context.Background(), config.DatabaseConfig{
		URL: "not-a-postgres-url",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewPostgresPoolAppliesConfigBeforePing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewPostgresPool(ctx, config.DatabaseConfig{
		URL:             "postgres://user:pass@127.0.0.1:5432/pylon?sslmode=disable",
		MaxOpenConns:    2,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "ping postgres") {
		t.Fatalf("expected ping postgres error, got %v", err)
	}
}
