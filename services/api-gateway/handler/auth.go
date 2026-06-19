package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	gatewayauth "github.com/kodokbakar/pylon/services/api-gateway/auth"

	"github.com/kodokbakar/pylon/internal/response"
)

type AuthService interface {
	Register(ctx context.Context, input gatewayauth.RegisterInput) (*gatewayauth.AuthResult, error)
	Login(ctx context.Context, input gatewayauth.LoginInput) (*gatewayauth.AuthResult, error)
	Refresh(ctx context.Context, input gatewayauth.RefreshInput) (*gatewayauth.RefreshResult, error)
}

type AuthHandler struct {
	service AuthService
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authResponse struct {
	User             userResponse `json:"user"`
	Token            string       `json:"token"`
	RefreshToken     string       `json:"refresh_token"`
	ExpiresAt        string       `json:"expires_at"`
	RefreshExpiresAt string       `json:"refresh_expires_at"`
}

type refreshResponse struct {
	Token            string `json:"token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresAt        string `json:"expires_at"`
	RefreshExpiresAt string `json:"refresh_expires_at"`
}

type userResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

func NewAuthHandler(service AuthService) *AuthHandler {
	if service == nil {
		service = gatewayauth.NewUnavailableService("auth service is not configured")
	}

	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body registerRequest
	if err := decodeJSON(w, r, &body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.service.Register(r.Context(), gatewayauth.RegisterInput{
		Username: body.Username,
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, authResultToResponse(result))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginRequest
	if err := decodeJSON(w, r, &body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.service.Login(r.Context(), gatewayauth.LoginInput{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	response.Success(w, http.StatusOK, authResultToResponse(result))
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := decodeJSON(w, r, &body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	result, err := h.service.Refresh(r.Context(), gatewayauth.RefreshInput{
		RefreshToken: body.RefreshToken,
	})
	if err != nil {
		writeAuthError(w, err)
		return
	}

	response.Success(w, http.StatusOK, refreshResultToResponse(result))
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gatewayauth.ErrInvalidInput):
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
	case errors.Is(err, gatewayauth.ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
	case errors.Is(err, gatewayauth.ErrAlreadyExists):
		response.Error(w, http.StatusConflict, "already_exists", "username or email already exists")
	case errors.Is(err, gatewayauth.ErrUnavailable):
		response.Error(w, http.StatusServiceUnavailable, "service_unavailable", "auth service unavailable")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func authResultToResponse(result *gatewayauth.AuthResult) authResponse {
	if result == nil {
		return authResponse{}
	}

	return authResponse{
		User:             userToResponse(result.User),
		Token:            result.Token,
		RefreshToken:     result.RefreshToken,
		ExpiresAt:        formatTime(result.ExpiresAt),
		RefreshExpiresAt: formatTime(result.RefreshExpiresAt),
	}
}

func refreshResultToResponse(result *gatewayauth.RefreshResult) refreshResponse {
	if result == nil {
		return refreshResponse{}
	}

	return refreshResponse{
		Token:            result.Token,
		RefreshToken:     result.RefreshToken,
		ExpiresAt:        formatTime(result.ExpiresAt),
		RefreshExpiresAt: formatTime(result.RefreshExpiresAt),
	}
}

func userToResponse(user gatewayauth.User) userResponse {
	return userResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: strings.TrimSpace(user.DisplayName),
		AvatarURL:   strings.TrimSpace(user.AvatarURL),
		CreatedAt:   formatTime(user.CreatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.UTC().Format(time.RFC3339)
}
