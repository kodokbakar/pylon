package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gatewayauth "github.com/kodokbakar/pylon/services/api-gateway/auth"
)

type fakeAuthService struct {
	registerFunc func(ctx context.Context, input gatewayauth.RegisterInput) (*gatewayauth.AuthResult, error)
	loginFunc    func(ctx context.Context, input gatewayauth.LoginInput) (*gatewayauth.AuthResult, error)
	refreshFunc  func(ctx context.Context, input gatewayauth.RefreshInput) (*gatewayauth.RefreshResult, error)
}

func (s fakeAuthService) Register(ctx context.Context, input gatewayauth.RegisterInput) (*gatewayauth.AuthResult, error) {
	if s.registerFunc != nil {
		return s.registerFunc(ctx, input)
	}

	return sampleAuthResult(), nil
}

func (s fakeAuthService) Login(ctx context.Context, input gatewayauth.LoginInput) (*gatewayauth.AuthResult, error) {
	if s.loginFunc != nil {
		return s.loginFunc(ctx, input)
	}

	return sampleAuthResult(), nil
}

func (s fakeAuthService) Refresh(ctx context.Context, input gatewayauth.RefreshInput) (*gatewayauth.RefreshResult, error) {
	if s.refreshFunc != nil {
		return s.refreshFunc(ctx, input)
	}

	return &gatewayauth.RefreshResult{
		Token:            "new-access-token",
		RefreshToken:     "new-refresh-token",
		ExpiresAt:        time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
		RefreshExpiresAt: time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC),
	}, nil
}

func TestAuthHandlerRegisterReturnsCreatedResponse(t *testing.T) {
	handler := NewAuthHandler(fakeAuthService{
		registerFunc: func(ctx context.Context, input gatewayauth.RegisterInput) (*gatewayauth.AuthResult, error) {
			if input.Username != "alice" {
				t.Fatalf("expected alice, got %q", input.Username)
			}

			if input.Email != "alice@example.com" {
				t.Fatalf("expected alice@example.com, got %q", input.Email)
			}

			if input.Password != "password123" {
				t.Fatalf("expected password to be forwarded")
			}

			return sampleAuthResult(), nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{
		"username": "alice",
		"email": "alice@example.com",
		"password": "password123"
	}`))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"token":"access-token"`) {
		t.Fatalf("expected access token response, got %s", body)
	}

	if !strings.Contains(body, `"refresh_token":"refresh-token"`) {
		t.Fatalf("expected refresh token response, got %s", body)
	}
}

func TestAuthHandlerLoginMapsInvalidCredentials(t *testing.T) {
	handler := NewAuthHandler(fakeAuthService{
		loginFunc: func(ctx context.Context, input gatewayauth.LoginInput) (*gatewayauth.AuthResult, error) {
			return nil, gatewayauth.ErrInvalidCredentials
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{
		"email": "alice@example.com",
		"password": "wrong-password"
	}`))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	if strings.Contains(rec.Body.String(), "wrong-password") {
		t.Fatalf("expected credential details to be hidden, got %s", rec.Body.String())
	}
}

func TestAuthHandlerRefreshReturnsRotatedTokens(t *testing.T) {
	handler := NewAuthHandler(fakeAuthService{
		refreshFunc: func(ctx context.Context, input gatewayauth.RefreshInput) (*gatewayauth.RefreshResult, error) {
			if input.RefreshToken != "old-refresh-token" {
				t.Fatalf("expected old refresh token, got %q", input.RefreshToken)
			}

			return &gatewayauth.RefreshResult{
				Token:            "new-access-token",
				RefreshToken:     "new-refresh-token",
				ExpiresAt:        time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
				RefreshExpiresAt: time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC),
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(`{
		"refresh_token": "old-refresh-token"
	}`))
	rec := httptest.NewRecorder()

	handler.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"token":"new-access-token"`) {
		t.Fatalf("expected new access token, got %s", body)
	}

	if !strings.Contains(body, `"refresh_token":"new-refresh-token"`) {
		t.Fatalf("expected new refresh token, got %s", body)
	}
}

func TestWriteAuthErrorMapsUnavailable(t *testing.T) {
	rec := httptest.NewRecorder()

	writeAuthError(rec, gatewayauth.ErrUnavailable)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
}

func TestAuthResponseFormatting(t *testing.T) {
	result := authResultToResponse(sampleAuthResult())

	if result.User.ID != "user-1" {
		t.Fatalf("expected user-1, got %q", result.User.ID)
	}

	if result.User.DisplayName != "Alice" {
		t.Fatalf("expected display name Alice, got %q", result.User.DisplayName)
	}

	if result.ExpiresAt == "" {
		t.Fatal("expected expires_at")
	}
}

func TestRefreshResponseFormatting(t *testing.T) {
	result := refreshResultToResponse(&gatewayauth.RefreshResult{
		Token:            "access-token",
		RefreshToken:     "refresh-token",
		ExpiresAt:        time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
		RefreshExpiresAt: time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC),
	})

	if result.Token != "access-token" {
		t.Fatalf("expected access token, got %q", result.Token)
	}

	if result.RefreshToken != "refresh-token" {
		t.Fatalf("expected refresh token, got %q", result.RefreshToken)
	}
}

func TestWriteAuthErrorMapsUnexpectedToGenericInternalError(t *testing.T) {
	rec := httptest.NewRecorder()

	writeAuthError(rec, errors.New("database password leaked"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	if strings.Contains(rec.Body.String(), "database password leaked") {
		t.Fatalf("expected internal error details to be hidden, got %s", rec.Body.String())
	}
}

func sampleAuthResult() *gatewayauth.AuthResult {
	return &gatewayauth.AuthResult{
		User: gatewayauth.User{
			ID:          "user-1",
			Username:    "alice",
			Email:       "alice@example.com",
			DisplayName: " Alice ",
			AvatarURL:   " https://example.com/avatar.png ",
			CreatedAt:   time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC),
		},
		Token:            "access-token",
		RefreshToken:     "refresh-token",
		ExpiresAt:        time.Date(2026, 6, 19, 12, 15, 0, 0, time.UTC),
		RefreshExpiresAt: time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC),
	}
}
