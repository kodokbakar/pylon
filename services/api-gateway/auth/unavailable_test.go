package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestUnavailableServiceReturnsUnavailableForAllMethods(t *testing.T) {
	service := NewUnavailableService("database url is not configured")

	_, err := service.Register(context.Background(), RegisterInput{})
	assertUnavailableError(t, err)

	_, err = service.Login(context.Background(), LoginInput{})
	assertUnavailableError(t, err)

	_, err = service.Refresh(context.Background(), RefreshInput{})
	assertUnavailableError(t, err)
}

func assertUnavailableError(t *testing.T, err error) {
	t.Helper()

	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}

	if !strings.Contains(err.Error(), "database url is not configured") {
		t.Fatalf("expected unavailable reason, got %v", err)
	}
}
