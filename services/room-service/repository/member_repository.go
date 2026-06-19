package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	roomservice "github.com/kodokbakar/pylon/services/room-service/service"
)

type MemberRepository struct {
	db *pgxpool.Pool
}

const addRoomMemberQuery = `
	INSERT INTO room_members (room_id, user_id, role)
	VALUES ($1, $2, $3)
	ON CONFLICT (room_id, user_id) DO NOTHING
`

const removeRoomMemberQuery = `
	DELETE FROM room_members
	WHERE room_id = $1
	  AND user_id = $2
`

const listRoomMembersQuery = `
	SELECT
		rm.room_id,
		rm.user_id,
		rm.role,
		rm.joined_at,
		u.username,
		COALESCE(u.display_name, ''),
		COALESCE(u.avatar_url, '')
	FROM room_members rm
	JOIN users u ON u.id = rm.user_id
	WHERE rm.room_id = $1
	ORDER BY rm.joined_at ASC, rm.user_id ASC
`

const getRoomMemberRoleQuery = `
	SELECT role
	FROM room_members
	WHERE room_id = $1
	  AND user_id = $2
`

func NewMemberRepository(db *pgxpool.Pool) (*MemberRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres pool is required")
	}

	return &MemberRepository{db: db}, nil
}

func (r *MemberRepository) Add(ctx context.Context, input roomservice.AddRoomMemberInput) error {
	if _, err := r.db.Exec(ctx, addRoomMemberQuery, input.RoomID, input.UserID, input.Role); err != nil {
		return fmt.Errorf("insert room member: %w", err)
	}

	return nil
}

func (r *MemberRepository) Remove(ctx context.Context, input roomservice.RemoveRoomMemberInput) error {
	if _, err := r.db.Exec(ctx, removeRoomMemberQuery, input.RoomID, input.UserID); err != nil {
		return fmt.Errorf("delete room member: %w", err)
	}

	return nil
}

func (r *MemberRepository) ListByRoomID(ctx context.Context, roomID string) ([]roomservice.RoomMember, error) {
	rows, err := r.db.Query(ctx, listRoomMembersQuery, roomID)
	if err != nil {
		return nil, fmt.Errorf("query room members: %w", err)
	}
	defer rows.Close()

	members := make([]roomservice.RoomMember, 0)
	for rows.Next() {
		var member roomservice.RoomMember

		if err := rows.Scan(
			&member.RoomID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.Username,
			&member.DisplayName,
			&member.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf("scan room member: %w", err)
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room members: %w", err)
	}

	return members, nil
}

func (r *MemberRepository) GetRole(ctx context.Context, roomID, userID string) (string, error) {
	var role string

	err := r.db.QueryRow(ctx, getRoomMemberRoleQuery, roomID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("%w: room member %s/%s", roomservice.ErrNotFound, roomID, userID)
	}
	if err != nil {
		return "", fmt.Errorf("select room member role: %w", err)
	}

	return role, nil
}
