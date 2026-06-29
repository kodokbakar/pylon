package handler

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"

	authv1 "github.com/kodokbakar/pylon/gen/pylon/auth/v1"
	gatewayauth "github.com/kodokbakar/pylon/services/api-gateway/auth"
)

func TestAuthConnectRegisterReturnsTypedResponse(t *testing.T) {
	handler := NewAuthConnectHandler(fakeAuthService{
		registerFunc: func(ctx context.Context, input gatewayauth.RegisterInput) (*gatewayauth.AuthResult, error) {
			if input.Username != "alice" {
				t.Fatalf("expected username alice, got %q", input.Username)
			}

			if input.Email != "alice@example.com" {
				t.Fatalf("expected email alice@example.com, got %q", input.Email)
			}

			if input.Password != "password123" {
				t.Fatalf("expected password password123, got %q", input.Password)
			}

			return sampleAuthResult(), nil
		},
	})

	res, err := handler.Register(context.Background(), connect.NewRequest(&authv1.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	}))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if res.Msg.GetToken() != "access-token" {
		t.Fatalf("expected access token, got %q", res.Msg.GetToken())
	}

	if res.Msg.GetUser().GetUsername() != "alice" {
		t.Fatalf("expected username alice, got %q", res.Msg.GetUser().GetUsername())
	}
}

func TestAuthConnectLoginReturnsTypedResponse(t *testing.T) {
	handler := NewAuthConnectHandler(fakeAuthService{
		loginFunc: func(ctx context.Context, input gatewayauth.LoginInput) (*gatewayauth.AuthResult, error) {
			if input.Email != "alice@example.com" {
				t.Fatalf("expected email alice@example.com, got %q", input.Email)
			}

			if input.Password != "password123" {
				t.Fatalf("expected password password123, got %q", input.Password)
			}

			return sampleAuthResult(), nil
		},
	})

	res, err := handler.Login(context.Background(), connect.NewRequest(&authv1.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	}))
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if res.Msg.GetToken() != "access-token" {
		t.Fatalf("expected access token, got %q", res.Msg.GetToken())
	}

	if res.Msg.GetRefreshToken() != "refresh-token" {
		t.Fatalf("expected refresh token, got %q", res.Msg.GetRefreshToken())
	}
}

func TestAuthConnectMapsInvalidCredentials(t *testing.T) {
	handler := NewAuthConnectHandler(fakeAuthService{
		loginFunc: func(ctx context.Context, input gatewayauth.LoginInput) (*gatewayauth.AuthResult, error) {
			return nil, gatewayauth.ErrInvalidCredentials
		},
	})

	_, err := handler.Login(context.Background(), connect.NewRequest(&authv1.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrong-password",
	}))
	if err == nil {
		t.Fatal("expected error")
	}

	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("expected unauthenticated code, got %v", connect.CodeOf(err))
	}
}

func TestAuthConnectMapsDuplicateRegister(t *testing.T) {
	handler := NewAuthConnectHandler(fakeAuthService{
		registerFunc: func(ctx context.Context, input gatewayauth.RegisterInput) (*gatewayauth.AuthResult, error) {
			return nil, gatewayauth.ErrAlreadyExists
		},
	})

	_, err := handler.Register(context.Background(), connect.NewRequest(&authv1.RegisterRequest{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	}))
	if err == nil {
		t.Fatal("expected error")
	}

	if connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("expected already exists code, got %v", connect.CodeOf(err))
	}
}

func TestAuthConnectHidesInternalErrors(t *testing.T) {
	handler := NewAuthConnectHandler(fakeAuthService{
		loginFunc: func(ctx context.Context, input gatewayauth.LoginInput) (*gatewayauth.AuthResult, error) {
			return nil, errors.New("database password leaked")
		},
	})

	_, err := handler.Login(context.Background(), connect.NewRequest(&authv1.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	}))
	if err == nil {
		t.Fatal("expected error")
	}

	if connect.CodeOf(err) != connect.CodeInternal {
		t.Fatalf("expected internal code, got %v", connect.CodeOf(err))
	}

	if connect.NewError(connect.CodeInternal, errors.New("internal server error")).Message() == "database password leaked" {
		t.Fatal("internal error leaked details")
	}
}
