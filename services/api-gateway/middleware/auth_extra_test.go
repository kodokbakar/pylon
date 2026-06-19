package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAuthMiddlewareRequiresSecret(t *testing.T) {
	_, err := NewAuthMiddleware(" ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBearerTokenRejectsMalformedAuthorizationHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic token-value")

	token, ok := bearerToken(req)
	if ok {
		t.Fatalf("expected malformed authorization header to be rejected, got token %q", token)
	}
}

func TestBearerTokenRejectsEmptyBearerToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")

	token, ok := bearerToken(req)
	if ok {
		t.Fatalf("expected empty bearer token to be rejected, got token %q", token)
	}
}

func TestBearerTokenUsesAccessTokenQueryParam(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws?access_token=query-token", nil)

	token, ok := bearerToken(req)
	if !ok {
		t.Fatal("expected token from query param")
	}

	if token != "query-token" {
		t.Fatalf("expected query-token, got %q", token)
	}
}

func TestBearerTokenUsesTokenQueryParamFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws?token=query-token", nil)

	token, ok := bearerToken(req)
	if !ok {
		t.Fatal("expected token from fallback query param")
	}

	if token != "query-token" {
		t.Fatalf("expected query-token, got %q", token)
	}
}

func TestContextAccessorsReturnFalseWhenMissing(t *testing.T) {
	ctx := context.Background()

	if userID, ok := UserIDFromContext(ctx); ok {
		t.Fatalf("expected missing user id, got %q", userID)
	}

	if username, ok := UsernameFromContext(ctx); ok {
		t.Fatalf("expected missing username, got %q", username)
	}

	if email, ok := EmailFromContext(ctx); ok {
		t.Fatalf("expected missing email, got %q", email)
	}
}
