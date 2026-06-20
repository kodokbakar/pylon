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
	createUserFunc         func(ctx context.Context, input CreateUserInput) (*User, error)
	findUserByEmailFunc    func(ctx context.Context, email string) (*UserWithPassword, error)
	storeRefreshTokenFunc  func(ctx context.Context, input StoreRefreshTokenInput) error
	findRefreshTokenFunc   func(ctx context.Context, tokenHash string) (*RefreshToken, error)
	revokeRefreshTokenFunc func(ctx context.Context, tokenHash string) error

	refreshTokens map[string]RefreshToken
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

func (r *fakeRepository) StoreRefreshToken(ctx context.Context, input StoreRefreshTokenInput) error {
	if r.storeRefreshTokenFunc != nil {
		return r.storeRefreshTokenFunc(ctx, input)
	}

	if r.refreshTokens == nil {
		r.refreshTokens = make(map[string]RefreshToken)
	}

	r.refreshTokens[input.TokenHash] = RefreshToken{
		ID:        input.TokenHash,
		UserID:    input.UserID,
		TokenHash: input.TokenHash,
		ExpiresAt: input.ExpiresAt,
		CreatedAt: time.Now(),
	}

	return nil
}

func (r *fakeRepository) FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	if r.findRefreshTokenFunc != nil {
		return r.findRefreshTokenFunc(ctx, tokenHash)
	}

	if r.refreshTokens == nil {
		return nil, ErrInvalidCredentials
	}

	refreshToken, ok := r.refreshTokens[tokenHash]
	if !ok {
		return nil, ErrInvalidCredentials
	}

	return &refreshToken, nil
}

func (r *fakeRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if r.revokeRefreshTokenFunc != nil {
		return r.revokeRefreshTokenFunc(ctx, tokenHash)
	}

	if r.refreshTokens == nil {
		return ErrInvalidCredentials
	}

	refreshToken, ok := r.refreshTokens[tokenHash]
	if !ok {
		return ErrInvalidCredentials
	}

	now := time.Now()
	refreshToken.RevokedAt = &now
	r.refreshTokens[tokenHash] = refreshToken

	return nil
}

func TestRegisterHashesPasswordAndReturnsTokens(t *testing.T) {
	svc := newTestService(t, &fakeRepository{
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

	svc := newTestService(t, &fakeRepository{
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

	svc := newTestService(t, &fakeRepository{
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
	svc := newTestService(t, &fakeRepository{})

	result, err := svc.issueAuthResult(context.Background(), User{
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

func TestRefreshRotatesRefreshToken(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(t, repo)

	result, err := svc.issueAuthResult(context.Background(), User{
		ID:       "user-1",
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	oldHash := hashRefreshToken(result.RefreshToken)

	refreshed, err := svc.Refresh(context.Background(), RefreshInput{
		RefreshToken: result.RefreshToken,
	})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if refreshed.Token == "" || refreshed.RefreshToken == "" {
		t.Fatalf("expected refreshed tokens, got %#v", refreshed)
	}

	if refreshed.RefreshToken == result.RefreshToken {
		t.Fatal("expected refresh token to rotate")
	}

	oldToken, err := repo.FindRefreshToken(context.Background(), oldHash)
	if err != nil {
		t.Fatalf("find old refresh token: %v", err)
	}

	if oldToken.RevokedAt == nil {
		t.Fatal("expected old refresh token to be revoked")
	}

	newHash := hashRefreshToken(refreshed.RefreshToken)
	newToken, err := repo.FindRefreshToken(context.Background(), newHash)
	if err != nil {
		t.Fatalf("find new refresh token: %v", err)
	}

	if newToken.RevokedAt != nil {
		t.Fatal("expected new refresh token to be active")
	}
}

func TestRefreshRejectsRevokedRefreshToken(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(t, repo)

	result, err := svc.issueAuthResult(context.Background(), User{
		ID:       "user-1",
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	oldHash := hashRefreshToken(result.RefreshToken)
	if err := repo.RevokeRefreshToken(context.Background(), oldHash); err != nil {
		t.Fatalf("revoke token: %v", err)
	}

	_, err = svc.Refresh(context.Background(), RefreshInput{
		RefreshToken: result.RefreshToken,
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestRefreshRejectsExpiredStoredRefreshToken(t *testing.T) {
	repo := &fakeRepository{}
	svc := newTestService(t, repo)

	result, err := svc.issueAuthResult(context.Background(), User{
		ID:       "user-1",
		Username: "alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("issue tokens: %v", err)
	}

	hash := hashRefreshToken(result.RefreshToken)
	token := repo.refreshTokens[hash]
	token.ExpiresAt = time.Now().Add(-time.Minute)
	repo.refreshTokens[hash] = token

	_, err = svc.Refresh(context.Background(), RefreshInput{
		RefreshToken: result.RefreshToken,
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
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

func TestRegisterStoresRefreshTokenHash(t *testing.T) {
	var storedHash string

	svc := newTestService(t, &fakeRepository{
		createUserFunc: func(ctx context.Context, input CreateUserInput) (*User, error) {
			return &User{
				ID:        "user-1",
				Username:  input.Username,
				Email:     input.Email,
				CreatedAt: time.Now(),
			}, nil
		},
		storeRefreshTokenFunc: func(ctx context.Context, input StoreRefreshTokenInput) error {
			storedHash = input.TokenHash

			if input.UserID != "user-1" {
				t.Fatalf("expected user-1, got %q", input.UserID)
			}

			if input.TokenHash == "" {
				t.Fatal("expected refresh token hash")
			}

			if input.ExpiresAt.IsZero() {
				t.Fatal("expected refresh token expiry")
			}

			return nil
		},
	})

	result, err := svc.Register(context.Background(), RegisterInput{
		Username: "alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if storedHash == "" {
		t.Fatal("expected refresh token to be stored")
	}

	if storedHash == result.RefreshToken {
		t.Fatal("expected stored value to be a hash, got raw refresh token")
	}

	if storedHash != hashRefreshToken(result.RefreshToken) {
		t.Fatal("expected stored hash to match refresh token")
	}
}
