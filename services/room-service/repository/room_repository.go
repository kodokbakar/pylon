package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	roomservice "github.com/kodokbakar/pylon/services/room-service/service"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

const createRoomQuery = `
	INSERT INTO rooms (name, type, created_by)
	VALUES ($1, $2, $3)
	RETURNING id, name, type, created_by, created_at
`

const getRoomByIDQuery = `
	SELECT id, name, type, created_by, created_at
	FROM rooms
	WHERE id = $1
`

const listRoomsByUserIDQuery = `
	SELECT r.id, r.name, r.type, r.created_by, r.created_at
	FROM rooms r
	JOIN room_members rm ON rm.room_id = r.id
	WHERE rm.user_id = $1
	ORDER BY r.created_at DESC, r.id DESC
`

const findDirectRoomQuery = `
	SELECT r.id, r.name, r.type, r.created_by, r.created_at
	FROM rooms r
	JOIN room_members rm1 ON rm1.room_id = r.id
	JOIN room_members rm2 ON rm2.room_id = r.id
	WHERE r.type = 'direct'
	  AND rm1.user_id = $1
	  AND rm2.user_id = $2
	LIMIT 1
`

func NewRoomRepository(db *pgxpool.Pool) (*RoomRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("postgres pool is required")
	}

	return &RoomRepository{db: db}, nil
}

func (r *RoomRepository) Create(ctx context.Context, input roomservice.CreateRoomRecordInput) (*roomservice.Room, error) {
	var room roomservice.Room
	var roomType string

	err := r.db.QueryRow(ctx, createRoomQuery, input.Name, string(input.Type), input.CreatedBy).Scan(
		&room.ID,
		&room.Name,
		&roomType,
		&room.CreatedBy,
		&room.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert room: %w", err)
	}

	room.Type = roomservice.RoomType(roomType)

	return &room, nil
}

func (r *RoomRepository) GetByID(ctx context.Context, roomID string) (*roomservice.Room, error) {
	var room roomservice.Room
	var roomType string

	err := r.db.QueryRow(ctx, getRoomByIDQuery, roomID).Scan(
		&room.ID,
		&room.Name,
		&roomType,
		&room.CreatedBy,
		&room.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: room %s", roomservice.ErrNotFound, roomID)
	}
	if err != nil {
		return nil, fmt.Errorf("select room by id: %w", err)
	}

	room.Type = roomservice.RoomType(roomType)

	return &room, nil
}

func (r *RoomRepository) ListByUserID(ctx context.Context, userID string) ([]roomservice.Room, error) {
	rows, err := r.db.Query(ctx, listRoomsByUserIDQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("query rooms by user id: %w", err)
	}
	defer rows.Close()

	rooms := make([]roomservice.Room, 0)
	for rows.Next() {
		var room roomservice.Room
		var roomType string

		if err := rows.Scan(
			&room.ID,
			&room.Name,
			&roomType,
			&room.CreatedBy,
			&room.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}

		room.Type = roomservice.RoomType(roomType)
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}

	return rooms, nil
}

func (r *RoomRepository) FindDirectRoom(ctx context.Context, userID, otherUserID string) (*roomservice.Room, error) {
	var room roomservice.Room
	var roomType string

	err := r.db.QueryRow(ctx, findDirectRoomQuery, userID, otherUserID).Scan(
		&room.ID,
		&room.Name,
		&roomType,
		&room.CreatedBy,
		&room.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: direct room for users %s and %s", roomservice.ErrNotFound, userID, otherUserID)
	}
	if err != nil {
		return nil, fmt.Errorf("select direct room: %w", err)
	}

	room.Type = roomservice.RoomType(roomType)

	return &room, nil
}
