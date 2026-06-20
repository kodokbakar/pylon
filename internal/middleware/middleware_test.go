package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHandlesPreflight(t *testing.T) {
	nextCalled := false
	handler := CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	if nextCalled {
		t.Fatal("expected next handler not called")
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected allow origin header, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSWithOriginsAllowsWhitelistedOrigin(t *testing.T) {
	handler := CORSWithOrigins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), []string{"https://app.example.com"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Fatalf("expected whitelisted origin header, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSWithOriginsRejectsUnknownOrigin(t *testing.T) {
	handler := CORSWithOrigins(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}), []string{"https://app.example.com"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://unknown.example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
}

func TestAuthRejectsMissingAuthorizationHeader(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthStoresBearerTokenInContext(t *testing.T) {
	handler := Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := TokenFromContext(r.Context())
		if !ok {
			t.Fatal("expected token in context")
		}

		if token != "token-value" {
			t.Fatalf("expected token-value, got %q", token)
		}

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token-value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestRecoveryHandlesPanic(t *testing.T) {
	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
}

func TestClientIPUsesFirstForwardedForValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10, 198.51.100.20, 192.0.2.30")
	req.RemoteAddr = "127.0.0.1:1234"

	got := clientIP(req)
	if got != "203.0.113.10" {
		t.Fatalf("expected first forwarded ip, got %q", got)
	}
}

func TestClientIPUsesRealIPWhenForwardedForMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.50")
	req.RemoteAddr = "127.0.0.1:1234"

	got := clientIP(req)
	if got != "203.0.113.50" {
		t.Fatalf("expected real ip, got %q", got)
	}
}
