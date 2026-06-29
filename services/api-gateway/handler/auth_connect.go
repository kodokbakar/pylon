package handler

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	authv1 "github.com/kodokbakar/pylon/gen/pylon/auth/v1"
	gatewayauth "github.com/kodokbakar/pylon/services/api-gateway/auth"
)

type AuthConnectHandler struct {
	service AuthService
}

func NewAuthConnectHandler(service AuthService) *AuthConnectHandler {
	if service == nil {
		service = gatewayauth.NewUnavailableService("auth service is not configured")
	}

	return &AuthConnectHandler{service: service}
}

func (h *AuthConnectHandler) Register(
	ctx context.Context,
	req *connect.Request[authv1.RegisterRequest],
) (*connect.Response[authv1.RegisterResponse], error) {
	result, err := h.service.Register(ctx, gatewayauth.RegisterInput{
		Username: req.Msg.GetUsername(),
		Email:    req.Msg.GetEmail(),
		Password: req.Msg.GetPassword(),
	})
	if err != nil {
		return nil, authErrorToConnect(err)
	}

	return connect.NewResponse(authResultToProtoRegister(result)), nil
}

func (h *AuthConnectHandler) Login(
	ctx context.Context,
	req *connect.Request[authv1.LoginRequest],
) (*connect.Response[authv1.LoginResponse], error) {
	result, err := h.service.Login(ctx, gatewayauth.LoginInput{
		Email:    req.Msg.GetEmail(),
		Password: req.Msg.GetPassword(),
	})
	if err != nil {
		return nil, authErrorToConnect(err)
	}

	return connect.NewResponse(authResultToProtoLogin(result)), nil
}

func (h *AuthConnectHandler) RefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RefreshTokenRequest],
) (*connect.Response[authv1.RefreshTokenResponse], error) {
	result, err := h.service.Refresh(ctx, gatewayauth.RefreshInput{
		RefreshToken: req.Msg.GetRefreshToken(),
	})
	if err != nil {
		return nil, authErrorToConnect(err)
	}

	return connect.NewResponse(&authv1.RefreshTokenResponse{
		Token:            result.Token,
		RefreshToken:     result.RefreshToken,
		ExpiresAt:        formatTime(result.ExpiresAt),
		RefreshExpiresAt: formatTime(result.RefreshExpiresAt),
	}), nil
}

func authResultToProtoRegister(result *gatewayauth.AuthResult) *authv1.RegisterResponse {
	if result == nil {
		return &authv1.RegisterResponse{}
	}

	return &authv1.RegisterResponse{
		User:             userToProto(result.User),
		Token:            result.Token,
		RefreshToken:     result.RefreshToken,
		ExpiresAt:        formatTime(result.ExpiresAt),
		RefreshExpiresAt: formatTime(result.RefreshExpiresAt),
	}
}

func authResultToProtoLogin(result *gatewayauth.AuthResult) *authv1.LoginResponse {
	if result == nil {
		return &authv1.LoginResponse{}
	}

	return &authv1.LoginResponse{
		User:             userToProto(result.User),
		Token:            result.Token,
		RefreshToken:     result.RefreshToken,
		ExpiresAt:        formatTime(result.ExpiresAt),
		RefreshExpiresAt: formatTime(result.RefreshExpiresAt),
	}
}

func userToProto(user gatewayauth.User) *authv1.User {
	return &authv1.User{
		Id:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarUrl:   user.AvatarURL,
		CreatedAt:   formatTime(user.CreatedAt),
	}
}

func authErrorToConnect(err error) error {
	switch {
	case errors.Is(err, gatewayauth.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, gatewayauth.ErrInvalidCredentials):
		return connect.NewError(connect.CodeUnauthenticated, errors.New("invalid credentials"))
	case errors.Is(err, gatewayauth.ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, errors.New("username or email already exists"))
	case errors.Is(err, gatewayauth.ErrUnavailable):
		return connect.NewError(connect.CodeUnavailable, errors.New("auth service unavailable"))
	default:
		return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
}
