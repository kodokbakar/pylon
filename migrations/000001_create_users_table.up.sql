CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    avatar_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_username_not_empty CHECK (length(trim(username)) > 0),
    CONSTRAINT users_email_not_empty CHECK (length(trim(email)) > 0),
    CONSTRAINT users_password_hash_not_empty CHECK (length(trim(password_hash)) > 0),
    CONSTRAINT users_username_max_length CHECK (char_length(username) <= 50),
    CONSTRAINT users_email_max_length CHECK (char_length(email) <= 255),
    CONSTRAINT users_display_name_max_length CHECK (display_name IS NULL OR char_length(display_name) <= 100),
    CONSTRAINT users_avatar_url_max_length CHECK (avatar_url IS NULL OR char_length(avatar_url) <= 500)
);

CREATE INDEX IF NOT EXISTS idx_users_username
    ON users(username);

CREATE INDEX IF NOT EXISTS idx_users_email
    ON users(email);