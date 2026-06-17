CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'group',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT rooms_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT rooms_type_valid CHECK (type IN ('direct', 'group', 'channel')),
    CONSTRAINT rooms_name_max_length CHECK (char_length(name) <= 255)
);

CREATE INDEX IF NOT EXISTS idx_rooms_created_by
    ON rooms(created_by);

CREATE INDEX IF NOT EXISTS idx_rooms_created_at
    ON rooms(created_at DESC);