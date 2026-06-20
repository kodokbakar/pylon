package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MembershipRepository struct {
	db *pgxpool.Pool
}

const isRoomMemberQuery = `
	SELECT EXISTS (
		SELECT 1
		FROM room_members
		WHERE room_id = $1
		  AND user_id = $2
	)
`

func NewMembershipRepository(db *pgxpool.Pool) (*MembershipRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres pool is required")
	}

	return &MembershipRepository{db: db}, nil
}

func (r *MembershipRepository) IsMember(ctx context.Context, roomID, userID string) (bool, error) {
	var exists bool

	if err := r.db.QueryRow(ctx, isRoomMemberQuery, roomID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check room membership: %w", err)
	}

	return exists, nil
}
