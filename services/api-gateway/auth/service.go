package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/kodokbakar/pylon/internal/config"
)

var ErrInvalidInput = errors.New("invalid input")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrAlreadyExists = errors.New("already exists")
var ErrUnavailable = errors.New("auth service unavailable")

const (
	minUsernameLen   = 3
	maxUsernameLen   = 50
	minPasswordLen   = 8
	maxEmailLen      = 255
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type User struct {
	ID          string
	Username    string
	Email       string
	DisplayName string
	AvatarURL   string
	CreatedAt   time.Time
}

type UserWithPassword struct {
	User
	PasswordHash string
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RegisterInput struct {
	Username string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type AuthResult struct {
	User             User
	Token            string
	RefreshToken     string
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
}

type RefreshResult struct {
	Token            string
	RefreshToken     string
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
}

type Repository interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*UserWithPassword, error)
	StoreRefreshToken(ctx context.Context, input StoreRefreshTokenInput) error
	FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}

type CreateUserInput struct {
	Username     string
	Email        string
	PasswordHash string
}

type StoreRefreshTokenInput struct {
	UserID    string
	TokenHash string
	ExpiresAt time.Time
}

type Service struct {
	repo       Repository
	secret     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewService(repo Repository, jwtCfg config.JWTConfig) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("auth repository is required")
	}

	secret := strings.TrimSpace(jwtCfg.Secret)
	if secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}

	accessTTL := jwtCfg.Expiry
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL := jwtCfg.RefreshExpiry
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	return &Service{
		repo:       repo,
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	username := strings.TrimSpace(input.Username)
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)

	if err := validateRegisterInput(username, email, password); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, CreateUserInput{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return s.issueAuthResult(ctx, *user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)

	if email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	if password == "" {
		return nil, fmt.Errorf("%w: password is required", ErrInvalidInput)
	}

	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, err
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("%w: invalid email or password", ErrInvalidCredentials)
	}

	return s.issueAuthResult(ctx, user.User)
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (*RefreshResult, error) {
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("%w: refresh token is required", ErrInvalidInput)
	}

	user, err := s.userFromRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	tokenHash := hashRefreshToken(refreshToken)

	storedToken, err := s.repo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid refresh token", ErrInvalidCredentials)
	}

	if storedToken.UserID != user.ID {
		return nil, fmt.Errorf("%w: refresh token user mismatch", ErrInvalidCredentials)
	}

	if storedToken.RevokedAt != nil {
		return nil, fmt.Errorf("%w: refresh token has been revoked", ErrInvalidCredentials)
	}

	if !storedToken.ExpiresAt.After(time.Now().UTC()) {
		return nil, fmt.Errorf("%w: refresh token has expired", ErrInvalidCredentials)
	}

	if err := s.repo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("revoke refresh token: %w", err)
	}

	result, err := s.issueAuthResult(ctx, user)
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		Token:            result.Token,
		RefreshToken:     result.RefreshToken,
		ExpiresAt:        result.ExpiresAt,
		RefreshExpiresAt: result.RefreshExpiresAt,
	}, nil
}

func (s *Service) issueAuthResult(ctx context.Context, user User) (*AuthResult, error) {
	accessToken, accessExpiresAt, err := s.issueToken(user, tokenTypeAccess, s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.issueToken(user, tokenTypeRefresh, s.refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	refreshTokenHash := hashRefreshToken(refreshToken)
	if err := s.repo.StoreRefreshToken(ctx, StoreRefreshTokenInput{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &AuthResult{
		User:             user,
		Token:            accessToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *Service) issueToken(user User, tokenType string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(ttl)
	tokenID, err := newTokenID()
	if err != nil {
		return "", time.Time{}, fmt.Errorf("generate token id: %w", err)
	}

	claims := jwt.MapClaims{
		"sub":        user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"token_type": tokenType,
		"jti":        tokenID,
		"iat":        now.Unix(),
		"exp":        expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt token: %w", err)
	}

	return tokenString, expiresAt, nil
}

func (s *Service) userFromRefreshToken(tokenString string) (User, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}

		return []byte(s.secret), nil
	})
	if err != nil {
		return User{}, fmt.Errorf("%w: invalid refresh token", ErrInvalidCredentials)
	}

	if !token.Valid {
		return User{}, fmt.Errorf("%w: invalid refresh token", ErrInvalidCredentials)
	}

	if stringClaim(claims, "token_type") != tokenTypeRefresh {
		return User{}, fmt.Errorf("%w: token is not a refresh token", ErrInvalidCredentials)
	}

	userID, err := claims.GetSubject()
	if err != nil {
		return User{}, fmt.Errorf("%w: refresh token subject is required", ErrInvalidCredentials)
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return User{}, fmt.Errorf("%w: refresh token subject is required", ErrInvalidCredentials)
	}

	return User{
		ID:       userID,
		Username: stringClaim(claims, "username"),
		Email:    stringClaim(claims, "email"),
	}, nil
}

func newTokenID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return hex.EncodeToString(bytes[:]), nil
}

func hashRefreshToken(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}

func validateRegisterInput(username, email, password string) error {
	if username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidInput)
	}

	if len([]rune(username)) < minUsernameLen {
		return fmt.Errorf("%w: username must be at least %d characters", ErrInvalidInput, minUsernameLen)
	}

	if len([]rune(username)) > maxUsernameLen {
		return fmt.Errorf("%w: username must be at most %d characters", ErrInvalidInput, maxUsernameLen)
	}

	if email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidInput)
	}

	if len(email) > maxEmailLen {
		return fmt.Errorf("%w: email must be at most %d characters", ErrInvalidInput, maxEmailLen)
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("%w: email is invalid", ErrInvalidInput)
	}

	if len([]rune(password)) < minPasswordLen {
		return fmt.Errorf("%w: password must be at least %d characters", ErrInvalidInput, minPasswordLen)
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func stringClaim(claims jwt.MapClaims, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}
