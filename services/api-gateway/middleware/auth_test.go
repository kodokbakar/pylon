package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateTokenReturnsUsernameClaim(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      "user-1",
		"username": "alice",
		"email":    "alice@example.com",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	userID, username, email, err := middleware.validateToken(tokenString)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	if userID != "user-1" {
		t.Fatalf("expected user-1, got %q", userID)
	}

	if username != "alice" {
		t.Fatalf("expected alice, got %q", username)
	}

	if email != "alice@example.com" {
		t.Fatalf("expected alice@example.com, got %q", email)
	}
}

func TestValidateTokenAllowsMissingUsernameClaim(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	userID, username, _, err := middleware.validateToken(tokenString)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	if userID != "user-1" {
		t.Fatalf("expected user-1, got %q", userID)
	}

	if username != "" {
		t.Fatalf("expected empty username, got %q", username)
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, _, _, err = middleware.validateToken(tokenString)
	if err == nil {
		t.Fatal("expected expired token error, got nil")
	}
}

func TestValidateTokenRejectsRefreshToken(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        "user-1",
		"token_type": "refresh",
		"exp":        time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, _, _, err = middleware.validateToken(tokenString)
	if err == nil {
		t.Fatal("expected refresh token rejection, got nil")
	}
}
