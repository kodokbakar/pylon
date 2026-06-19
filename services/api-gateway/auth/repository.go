package auth

import (
	"context"
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
