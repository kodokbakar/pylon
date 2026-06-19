package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kodokbakar/pylon/internal/config"
)

type fakeRepository struct {
	createUserFunc      func(ctx context.Context, input CreateUserInput) (*User, error)
	findUserByEmailFunc func(ctx context.Context, email string) (*UserWithPassword, error)
}

func (r fakeRepository) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	if r.createUserFunc != nil {
		return r.createUserFunc(ctx, input)
	}

	return &User{
		ID:        "user-1",
		Username:  input.Username,
		Email:     input.Email,
		CreatedAt: time.Now(),
	}, nil
}

func (r fakeRepository) FindUserByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	if r.findUserByEmailFunc != nil {
		return r.findUserByEmailFunc(ctx, email)
	}

	return nil, ErrInvalidCredentials
}

func TestRegisterHashesPasswordAndReturnsTokens(t *testing.T) {
	svc := newTestService(t, fakeRepository{
		createUserFunc: func(ctx context.Context, input CreateUserInput) (*User, error) {
			if input.Username != "alice" {
				t.Fatalf("expected username alice, got %q", input.Username)
			}

			if input.Email != "alice@example.com" {
				t.Fatalf("expected normalized email, got %q", input.Email)
			}

			if input.PasswordHash == "password123" {
				t.Fatal("expected password to be hashed")
			}

			if err := bcrypt.CompareHashAndPassword([]byte(input.PasswordHash), []byte("password123")); err != nil {
				t.Fatalf("password hash mismatch: %v", err)
			}

			return &User{
				ID:        "user-1",
				Username:  input.Username,
				Email:     input.Email,
				CreatedAt: time.Now(),
			}, nil
		},
	})

	result, err := svc.Register(context.Background(), RegisterInput{
		Username: " alice ",
		Email:    " ALICE@example.com ",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if result.Token == "" {
		t.Fatal("expected access token")
	}

	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
}

func TestLoginRejectsInvalidPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	svc := newTestService(t, fakeRepository{
		findUserByEmailFunc: func(ctx context.Context, email string) (*UserWithPassword, error) {
			return &UserWithPassword{
				User: User{
					ID:       "user-1",
					Username: "alice",
					Email:    email,
				},
				PasswordHash: string(hash),
			}, nil
		},
	})

	_, err = svc.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestLoginReturnsTokens(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	svc := newTestService(t, fakeRepository{
		findUserByEmailFunc: func(ctx context.Context, email string) (*UserWithPassword, error) {
			return &UserWithPassword{
				User: User{
					ID:       "user-1",
					Username: "alice",
					Email:    email,
				},
				PasswordHash: string(hash),
			}, nil
		},
	})

	result, err := svc.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if result.Token == "" || result.RefreshToken == "" {
		t.Fatalf("expected tokens, got %#v", result)
	}
}

func TestRefreshRejectsAccessToken(t *testing.T) {
	svc := newTestService(t, fakeRepository{})

	result, err := svc.issueAuthResult(User{
		ID:       "user-1",
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	_, err = svc.Refresh(context.Background(), RefreshInput{
		RefreshToken: result.Token,
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestRefreshReturnsNewTokens(t *testing.T) {
	svc := newTestService(t, fakeRepository{})

	result, err := svc.issueAuthResult(User{
		ID:       "user-1",
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	refreshed, err := svc.Refresh(context.Background(), RefreshInput{
		RefreshToken: result.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if refreshed.Token == "" || refreshed.RefreshToken == "" {
		t.Fatalf("expected refreshed tokens, got %#v", refreshed)
	}
}

func newTestService(t *testing.T, repo Repository) *Service {
	t.Helper()

	svc, err := NewService(repo, config.JWTConfig{
		Secret:        "test-secret",
		Expiry:        15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("create auth service: %v", err)
	}

	return svc
}
