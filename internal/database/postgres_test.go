package database

import (
	"context"
	"testing"

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
