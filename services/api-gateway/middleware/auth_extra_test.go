package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestRequireAuthStoresClaimsInContext(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok || userID != "user-1" {
			t.Fatalf("expected user-1 in context, got %q", userID)
		}

		username, ok := UsernameFromContext(r.Context())
		if !ok || username != "alice" {
			t.Fatalf("expected alice in context, got %q", username)
		}

		email, ok := EmailFromContext(r.Context())
		if !ok || email != "alice@example.com" {
			t.Fatalf("expected alice@example.com in context, got %q", email)
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	req.Header.Set("Authorization", "Bearer "+signedMiddlewareToken(t, jwt.MapClaims{
		"sub":        "user-1",
		"username":   "alice",
		"email":      "alice@example.com",
		"token_type": "access",
		"exp":        time.Now().Add(time.Hour).Unix(),
	}))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRequireAuthRejectsMissingBearerToken(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestRequireAuthRejectsInvalidToken(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestValidateTokenUsesPreferredUsernameFallback(t *testing.T) {
	middleware, err := NewAuthMiddleware("test-secret")
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	token := signedMiddlewareToken(t, jwt.MapClaims{
		"sub":                "user-1",
		"preferred_username": "preferred-alice",
		"token_type":         "access",
		"exp":                time.Now().Add(time.Hour).Unix(),
	})

	userID, username, _, err := middleware.validateToken(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	if userID != "user-1" {
		t.Fatalf("expected user-1, got %q", userID)
	}

	if username != "preferred-alice" {
		t.Fatalf("expected preferred username fallback, got %q", username)
	}
}

func signedMiddlewareToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return tokenString
}
