CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'text',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT messages_content_not_empty CHECK (length(trim(content)) > 0),
    CONSTRAINT messages_content_max_length CHECK (char_length(content) <= 10000),
    CONSTRAINT messages_type_valid CHECK (type IN ('text', 'image', 'system'))
);

CREATE INDEX IF NOT EXISTS idx_messages_room_id
    ON messages(room_id);

CREATE INDEX IF NOT EXISTS idx_messages_created_at
    ON messages(created_at);

CREATE INDEX IF NOT EXISTS idx_messages_room_created
    ON messages(room_id, created_at DESC, id DESC);