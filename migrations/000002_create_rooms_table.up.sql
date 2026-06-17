CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'group',
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT rooms_name_not_empty CHECK (length(trim(name)) > 0),
    CONSTRAINT rooms_type_valid CHECK (type IN ('direct', 'group', 'channel'))
);

CREATE TABLE IF NOT EXISTS room_members (
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (room_id, user_id),
    CONSTRAINT room_members_role_valid CHECK (role IN ('owner', 'admin', 'member'))
);

CREATE INDEX IF NOT EXISTS idx_rooms_created_by
    ON rooms(created_by);

CREATE INDEX IF NOT EXISTS idx_room_members_user_id
    ON room_members(user_id);

CREATE INDEX IF NOT EXISTS idx_room_members_room_id
    ON room_members(room_id);