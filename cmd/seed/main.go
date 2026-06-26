package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/kodokbakar/pylon/internal/config"
	"github.com/kodokbakar/pylon/internal/database"
)

const defaultDemoPassword = "password123"

type demoUser struct {
	Username    string
	Email       string
	DisplayName string
}

type seededUser struct {
	ID          string
	Username    string
	Email       string
	DisplayName string
}

type demoRoom struct {
	Name    string
	Type    string
	Creator string
	Members []string
}

type seededRoom struct {
	ID   string
	Name string
	Type string
}

type demoMessage struct {
	Sender  string
	Content string
}

type demoRoomMessages struct {
	RoomName string
	Messages []demoMessage
}

type insertedMessage struct {
	ID       string
	RoomID   string
	RoomName string
	Sender   string
}

type demoNotification struct {
	User      string
	Type      string
	Title     string
	Body      string
	RoomName  string
	MessageID string
	Read      bool
	CreatedAt time.Time
}

type seedStats struct {
	UsersCreated          int
	UsersUpdated          int
	RoomsCreated          int
	RoomsExisting         int
	RoomMembersInserted   int
	MessagesInserted      int
	NotificationsInserted int
}

var demoUsers = []demoUser{
	{Username: "alice", Email: "alice@example.com", DisplayName: "Alice"},
	{Username: "bob", Email: "bob@example.com", DisplayName: "Bob"},
	{Username: "charlie", Email: "charlie@example.com", DisplayName: "Charlie"},
	{Username: "diana", Email: "diana@example.com", DisplayName: "Diana"},
	{Username: "eve", Email: "eve@example.com", DisplayName: "Eve"},
}

var demoRooms = []demoRoom{
	{
		Name:    "General",
		Type:    "channel",
		Creator: "alice",
		Members: []string{"alice", "bob", "charlie", "diana", "eve"},
	},
	{
		Name:    "Random",
		Type:    "channel",
		Creator: "bob",
		Members: []string{"alice", "bob", "charlie", "diana", "eve"},
	},
	{
		Name:    "Alice & Bob",
		Type:    "direct",
		Creator: "alice",
		Members: []string{"alice", "bob"},
	},
	{
		Name:    "Dev Team",
		Type:    "group",
		Creator: "alice",
		Members: []string{"alice", "bob", "charlie"},
	},
}

var demoMessages = []demoRoomMessages{
	{
		RoomName: "General",
		Messages: []demoMessage{
			{Sender: "alice", Content: "Welcome to General! This room is for day-to-day project updates."},
			{Sender: "bob", Content: "Morning everyone, I pushed the latest Docker fixes."},
			{Sender: "charlie", Content: "Nice. I will check the API Gateway routes today."},
			{Sender: "diana", Content: "Frontend can start testing auth and room list with this data."},
			{Sender: "eve", Content: "I am checking notification behavior after messages are sent."},
			{Sender: "alice", Content: "Reminder: keep logs clean and errors wrapped."},
			{Sender: "bob", Content: "PostgreSQL and Redis look stable locally."},
			{Sender: "charlie", Content: "Kafka is also up. Consumer lag is zero on my machine."},
			{Sender: "diana", Content: "The UI looks much better when there is real message history."},
			{Sender: "eve", Content: "I added a few notes for QA verification."},
			{Sender: "alice", Content: "Great work. Let's keep the demo data predictable."},
			{Sender: "bob", Content: "Agreed. Predictable seeds make debugging faster."},
		},
	},
	{
		RoomName: "Random",
		Messages: []demoMessage{
			{Sender: "bob", Content: "Random room is officially open."},
			{Sender: "diana", Content: "Important question: tabs or spaces?"},
			{Sender: "charlie", Content: "The formatter already chose for us."},
			{Sender: "eve", Content: "That is the safest answer."},
			{Sender: "alice", Content: "I brought virtual coffee for everyone."},
			{Sender: "bob", Content: "Virtual coffee accepted."},
			{Sender: "diana", Content: "The demo chat should feel alive now."},
			{Sender: "charlie", Content: "Someone should add a meme endpoint someday."},
			{Sender: "eve", Content: "Only if it has tests."},
			{Sender: "alice", Content: "Everything gets tests."},
		},
	},
	{
		RoomName: "Alice & Bob",
		Messages: []demoMessage{
			{Sender: "alice", Content: "Hey Bob, can you review the room membership flow?"},
			{Sender: "bob", Content: "Sure, I will check direct room behavior first."},
			{Sender: "alice", Content: "Direct rooms should stay between exactly two users."},
			{Sender: "bob", Content: "Makes sense. I will also check duplicate room prevention."},
			{Sender: "alice", Content: "Thanks. The seed should help frontend test this conversation."},
			{Sender: "bob", Content: "Yes, this is much better than empty state testing."},
			{Sender: "alice", Content: "Ping me when you finish the review."},
			{Sender: "bob", Content: "Will do."},
			{Sender: "alice", Content: "Also check notification counts if you can."},
			{Sender: "bob", Content: "On it."},
		},
	},
	{
		RoomName: "Dev Team",
		Messages: []demoMessage{
			{Sender: "alice", Content: "Dev Team room is for engineering notes."},
			{Sender: "bob", Content: "Current focus: seed data and final polish."},
			{Sender: "charlie", Content: "I will validate message history pagination."},
			{Sender: "alice", Content: "Please also test room member listing."},
			{Sender: "bob", Content: "The seeder should add owner and member roles correctly."},
			{Sender: "charlie", Content: "Good. That will help test access control."},
			{Sender: "alice", Content: "Messages use varied timestamps so sorting can be tested."},
			{Sender: "bob", Content: "Nice detail."},
			{Sender: "charlie", Content: "Notification examples should cover message and mention types."},
			{Sender: "alice", Content: "Agreed. Keep it realistic but simple."},
			{Sender: "bob", Content: "After this, frontend can test room list and message history quickly."},
			{Sender: "charlie", Content: "Seeder accepted."},
		},
	},
}

func main() {
	log.SetFlags(0)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if strings.TrimSpace(cfg.Database.URL) == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	seedPassword := getSeedPassword()

	stats, err := seed(ctx, db, seedPassword)
	if err != nil {
		log.Fatalf("seed database: %v", err)
	}

	fmt.Println("Pylon demo data seeded successfully.")
	fmt.Printf("users: %d created, %d updated\n", stats.UsersCreated, stats.UsersUpdated)
	fmt.Printf("rooms: %d created, %d existing\n", stats.RoomsCreated, stats.RoomsExisting)
	fmt.Printf("room_members: %d inserted\n", stats.RoomMembersInserted)
	fmt.Printf("messages: %d inserted\n", stats.MessagesInserted)
	fmt.Printf("notifications: %d inserted\n", stats.NotificationsInserted)
	fmt.Println("note: users and rooms are idempotent; messages and notifications are appended on each run.")
}

func getSeedPassword() string {
	password := strings.TrimSpace(os.Getenv("SEED_PASSWORD"))
	if password == "" {
		return defaultDemoPassword
	}

	return password
}

func seed(ctx context.Context, db *pgxpool.Pool, seedPassword string) (*seedStats, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin seed transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	stats := &seedStats{}

	users, err := seedUsers(ctx, tx, stats, seedPassword)
	if err != nil {
		return nil, err
	}

	rooms, err := seedRooms(ctx, tx, users, stats)
	if err != nil {
		return nil, err
	}

	messages, err := seedMessages(ctx, tx, users, rooms, stats)
	if err != nil {
		return nil, err
	}

	if err := seedNotifications(ctx, tx, users, rooms, messages, stats); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit seed transaction: %w", err)
	}

	committed = true

	return stats, nil
}

func seedUsers(ctx context.Context, tx pgx.Tx, stats *seedStats, seedPassword string) (map[string]seededUser, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash demo password: %w", err)
	}

	users := make(map[string]seededUser, len(demoUsers))

	for _, demo := range demoUsers {
		email := strings.ToLower(strings.TrimSpace(demo.Email))
		username := strings.TrimSpace(demo.Username)
		displayName := strings.TrimSpace(demo.DisplayName)

		var existed bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM users
				WHERE email = $1
			)
		`, email).Scan(&existed); err != nil {
			return nil, fmt.Errorf("check user %s: %w", email, err)
		}

		var user seededUser
		err := tx.QueryRow(ctx, `
			INSERT INTO users (username, email, password_hash, display_name)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (email) DO UPDATE
			SET
				password_hash = EXCLUDED.password_hash,
				display_name = EXCLUDED.display_name,
				updated_at = NOW()
			RETURNING id::text, username, email, COALESCE(display_name, '')
		`, username, email, string(passwordHash), displayName).Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.DisplayName,
		)
		if err != nil {
			return nil, fmt.Errorf("upsert user %s: %w", email, err)
		}

		if existed {
			stats.UsersUpdated++
		} else {
			stats.UsersCreated++
		}

		users[demo.Username] = user
	}

	return users, nil
}

func seedRooms(ctx context.Context, tx pgx.Tx, users map[string]seededUser, stats *seedStats) (map[string]seededRoom, error) {
	rooms := make(map[string]seededRoom, len(demoRooms))

	for _, demo := range demoRooms {
		creator, ok := users[demo.Creator]
		if !ok {
			return nil, fmt.Errorf("room %s creator %s not found", demo.Name, demo.Creator)
		}

		room, created, err := findOrCreateRoom(ctx, tx, demo, creator.ID)
		if err != nil {
			return nil, err
		}

		if created {
			stats.RoomsCreated++
		} else {
			stats.RoomsExisting++
		}

		rooms[demo.Name] = room

		inserted, err := seedRoomMembers(ctx, tx, room.ID, demo, users)
		if err != nil {
			return nil, err
		}

		stats.RoomMembersInserted += inserted
	}

	return rooms, nil
}

func findOrCreateRoom(ctx context.Context, tx pgx.Tx, demo demoRoom, creatorID string) (seededRoom, bool, error) {
	var room seededRoom

	err := tx.QueryRow(ctx, `
		SELECT id::text, name, type
		FROM rooms
		WHERE name = $1
		  AND type = $2
		ORDER BY created_at ASC, id ASC
		LIMIT 1
	`, demo.Name, demo.Type).Scan(&room.ID, &room.Name, &room.Type)
	if err == nil {
		return room, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return seededRoom{}, false, fmt.Errorf("find room %s/%s: %w", demo.Name, demo.Type, err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO rooms (name, type, created_by, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, name, type
	`, demo.Name, demo.Type, creatorID, time.Now().UTC().Add(-48*time.Hour)).Scan(
		&room.ID,
		&room.Name,
		&room.Type,
	)
	if err != nil {
		return seededRoom{}, false, fmt.Errorf("insert room %s/%s: %w", demo.Name, demo.Type, err)
	}

	return room, true, nil
}

func seedRoomMembers(ctx context.Context, tx pgx.Tx, roomID string, demo demoRoom, users map[string]seededUser) (int, error) {
	inserted := 0
	seen := make(map[string]bool, len(demo.Members))

	for _, username := range demo.Members {
		if seen[username] {
			continue
		}
		seen[username] = true

		user, ok := users[username]
		if !ok {
			return 0, fmt.Errorf("room %s member %s not found", demo.Name, username)
		}

		role := "member"
		if username == demo.Creator {
			role = "owner"
		}

		tag, err := tx.Exec(ctx, `
			INSERT INTO room_members (room_id, user_id, role)
			VALUES ($1, $2, $3)
			ON CONFLICT (room_id, user_id) DO NOTHING
		`, roomID, user.ID, role)
		if err != nil {
			return 0, fmt.Errorf("insert room member %s/%s: %w", demo.Name, username, err)
		}

		inserted += int(tag.RowsAffected())
	}

	return inserted, nil
}

func seedMessages(
	ctx context.Context,
	tx pgx.Tx,
	users map[string]seededUser,
	rooms map[string]seededRoom,
	stats *seedStats,
) (map[string][]insertedMessage, error) {
	inserted := make(map[string][]insertedMessage, len(demoMessages))
	now := time.Now().UTC()

	for roomIndex, roomMessages := range demoMessages {
		roomName := roomMessages.RoomName
		messages := roomMessages.Messages

		room, ok := rooms[roomName]
		if !ok {
			return nil, fmt.Errorf("message room %s not found", roomName)
		}

		inserted[roomName] = make([]insertedMessage, 0, len(messages))

		for i, demo := range messages {
			user, ok := users[demo.Sender]
			if !ok {
				return nil, fmt.Errorf("message sender %s not found", demo.Sender)
			}

			roomOffsetHours := 24 - (roomIndex * 3)
			createdAt := now.Add(-time.Duration(roomOffsetHours) * time.Hour).Add(time.Duration(i) * 9 * time.Minute)

			var messageID string
			if err := tx.QueryRow(ctx, `
				INSERT INTO messages (room_id, sender_id, content, type, created_at)
				VALUES ($1, $2, $3, 'text', $4)
				RETURNING id::text
			`, room.ID, user.ID, demo.Content, createdAt).Scan(&messageID); err != nil {
				return nil, fmt.Errorf("insert message for room %s sender %s: %w", roomName, demo.Sender, err)
			}

			inserted[roomName] = append(inserted[roomName], insertedMessage{
				ID:       messageID,
				RoomID:   room.ID,
				RoomName: roomName,
				Sender:   demo.Sender,
			})

			stats.MessagesInserted++
		}
	}

	return inserted, nil
}

func seedNotifications(
	ctx context.Context,
	tx pgx.Tx,
	users map[string]seededUser,
	rooms map[string]seededRoom,
	messages map[string][]insertedMessage,
	stats *seedStats,
) error {
	now := time.Now().UTC()

	notifications := []demoNotification{
		{
			User:      "alice",
			Type:      "mention",
			Title:     "Bob mentioned you",
			Body:      "Bob mentioned you in General.",
			RoomName:  "General",
			MessageID: firstMessageIDBySender(messages["General"], "bob"),
			Read:      false,
			CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			User:      "alice",
			Type:      "message",
			Title:     "New Dev Team message",
			Body:      "Charlie posted an update in Dev Team.",
			RoomName:  "Dev Team",
			MessageID: firstMessageIDBySender(messages["Dev Team"], "charlie"),
			Read:      false,
			CreatedAt: now.Add(-90 * time.Minute),
		},
		{
			User:      "bob",
			Type:      "message",
			Title:     "New direct message",
			Body:      "Alice sent you a direct message.",
			RoomName:  "Alice & Bob",
			MessageID: firstMessageIDBySender(messages["Alice & Bob"], "alice"),
			Read:      false,
			CreatedAt: now.Add(-45 * time.Minute),
		},
		{
			User:      "bob",
			Type:      "invite",
			Title:     "You were added to Dev Team",
			Body:      "Alice added you to the Dev Team room.",
			RoomName:  "Dev Team",
			MessageID: "",
			Read:      true,
			CreatedAt: now.Add(-30 * time.Minute),
		},
	}

	for _, demo := range notifications {
		user, ok := users[demo.User]
		if !ok {
			return fmt.Errorf("notification user %s not found", demo.User)
		}

		room, ok := rooms[demo.RoomName]
		if !ok {
			return fmt.Errorf("notification room %s not found", demo.RoomName)
		}

		var messageID any
		if demo.MessageID != "" {
			messageID = demo.MessageID
		}

		if _, err := tx.Exec(ctx, `
	INSERT INTO notifications (user_id, type, title, body, room_id, message_id, read, created_at)
	VALUES ($1, $2, $3, $4, $5::uuid, $6::uuid, $7, $8)
`, user.ID, demo.Type, demo.Title, demo.Body, room.ID, messageID, demo.Read, demo.CreatedAt); err != nil {
			return fmt.Errorf("insert notification for %s: %w", demo.User, err)
		}

		stats.NotificationsInserted++
	}

	return nil
}

func firstMessageIDBySender(messages []insertedMessage, sender string) string {
	for _, message := range messages {
		if message.Sender == sender {
			return message.ID
		}
	}

	return ""
}
