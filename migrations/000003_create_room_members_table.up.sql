CREATE TABLE IF NOT EXISTS room_members (
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (room_id, user_id),
    CONSTRAINT room_members_role_valid CHECK (role IN ('owner', 'admin', 'member'))
);

CREATE INDEX IF NOT EXISTS idx_room_members_user_id
    ON room_members(user_id);

CREATE INDEX IF NOT EXISTS idx_room_members_room_id
    ON room_members(room_id);

CREATE INDEX IF NOT EXISTS idx_room_members_joined_at
    ON room_members(joined_at);