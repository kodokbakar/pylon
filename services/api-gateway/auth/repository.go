package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const createUserQuery = `
	INSERT INTO users (username, email, password_hash)
	VALUES ($1, $2, $3)
	RETURNING id::text, username, email, COALESCE(display_name, ''), COALESCE(avatar_url, ''), created_at
`

const findUserByEmailQuery = `
	SELECT id::text, username, email, password_hash, COALESCE(display_name, ''), COALESCE(avatar_url, ''), created_at
	FROM users
	WHERE email = $1
`

const storeRefreshTokenQuery = `
	INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)
`

const findRefreshTokenQuery = `
	SELECT id::text, user_id::text, token_hash, expires_at, revoked_at, created_at
	FROM refresh_tokens
	WHERE token_hash = $1
`

const revokeRefreshTokenQuery = `
	UPDATE refresh_tokens
	SET revoked_at = NOW()
	WHERE token_hash = $1
	  AND revoked_at IS NULL
`

type RepositoryPostgres struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) (*RepositoryPostgres, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres pool is required")
	}

	return &RepositoryPostgres{db: db}, nil
}

func (r *RepositoryPostgres) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	var user User

	err := r.db.QueryRow(ctx, createUserQuery, input.Username, input.Email, input.PasswordHash).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, mapPostgresError(err)
	}

	return &user, nil
}

func (r *RepositoryPostgres) FindUserByEmail(ctx context.Context, email string) (*UserWithPassword, error) {
	var user UserWithPassword

	err := r.db.QueryRow(ctx, findUserByEmailQuery, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: invalid email or password", ErrInvalidCredentials)
	}
	if err != nil {
		return nil, mapPostgresError(err)
	}

	return &user, nil
}

func (r *RepositoryPostgres) StoreRefreshToken(ctx context.Context, input StoreRefreshTokenInput) error {
	if _, err := r.db.Exec(ctx, storeRefreshTokenQuery, input.UserID, input.TokenHash, input.ExpiresAt); err != nil {
		return mapPostgresError(err)
	}

	return nil
}

func (r *RepositoryPostgres) FindRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var refreshToken RefreshToken
	var revokedAt sql.NullTime

	err := r.db.QueryRow(ctx, findRefreshTokenQuery, tokenHash).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.TokenHash,
		&refreshToken.ExpiresAt,
		&revokedAt,
		&refreshToken.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: invalid refresh token", ErrInvalidCredentials)
	}
	if err != nil {
		return nil, mapPostgresError(err)
	}

	if revokedAt.Valid {
		refreshToken.RevokedAt = &revokedAt.Time
	}

	return &refreshToken, nil
}

func (r *RepositoryPostgres) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	tag, err := r.db.Exec(ctx, revokeRefreshTokenQuery, tokenHash)
	if err != nil {
		return mapPostgresError(err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: refresh token is already revoked or not found", ErrInvalidCredentials)
	}

	return nil
}

func mapPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: username or email already exists", ErrAlreadyExists)
		case "23502", "23514":
			return fmt.Errorf("%w: invalid user data", ErrInvalidInput)
		}
	}

	return err
}
