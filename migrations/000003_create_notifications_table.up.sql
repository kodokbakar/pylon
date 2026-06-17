CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    room_id UUID REFERENCES rooms(id) ON DELETE SET NULL,
    read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT notifications_type_valid CHECK (type IN ('message', 'invite', 'mention')),
    CONSTRAINT notifications_title_not_empty CHECK (length(trim(title)) > 0),
    CONSTRAINT notifications_body_not_empty CHECK (length(trim(body)) > 0),
    CONSTRAINT notifications_body_max_length CHECK (char_length(body) <= 10000),
    CONSTRAINT notifications_title_max_length CHECK (char_length(title) <= 255)
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id
    ON notifications(user_id);

CREATE INDEX IF NOT EXISTS idx_notifications_user_read
    ON notifications(user_id, read);

CREATE INDEX IF NOT EXISTS idx_notifications_created_at
    ON notifications(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created
    ON notifications(user_id, created_at DESC, id DESC);