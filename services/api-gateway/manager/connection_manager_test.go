package manager

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestConnectionManagerAddRemoveAndCount(t *testing.T) {
	manager := NewConnectionManager(10)

	if manager.Count() != 0 {
		t.Fatalf("expected initial count 0, got %d", manager.Count())
	}

	firstConn := testConnection("user-1")
	secondConn := testConnection("user-1")

	if !manager.Add(firstConn) {
		t.Fatal("expected first connection to be added")
	}

	if !manager.Add(secondConn) {
		t.Fatal("expected second connection to be added")
	}

	if manager.Count() != 2 {
		t.Fatalf("expected count 2, got %d", manager.Count())
	}

	if manager.UserConnectionCount("user-1") != 2 {
		t.Fatalf("expected user connection count 2, got %d", manager.UserConnectionCount("user-1"))
	}

	manager.Remove(firstConn)

	if manager.Count() != 1 {
		t.Fatalf("expected count 1 after removing first connection, got %d", manager.Count())
	}

	manager.Remove(secondConn)

	if manager.Count() != 0 {
		t.Fatalf("expected count 0 after removing all connections, got %d", manager.Count())
	}
}

func TestConnectionManagerAddWithUserLimitRejectsSecondConnectionForSameUser(t *testing.T) {
	manager := NewConnectionManager(10)

	if !manager.AddWithUserLimit(testConnection("user-1"), 1) {
		t.Fatal("expected first connection to be added")
	}

	if manager.AddWithUserLimit(testConnection("user-1"), 1) {
		t.Fatal("expected second connection for same user to be rejected")
	}

	if !manager.AddWithUserLimit(testConnection("user-2"), 1) {
		t.Fatal("expected different user connection to be added")
	}

	if manager.Count() != 2 {
		t.Fatalf("expected count 2, got %d", manager.Count())
	}
}

func TestConnectionManagerRejectsOverLimit(t *testing.T) {
	manager := NewConnectionManager(1)

	if !manager.Add(testConnection("user-1")) {
		t.Fatal("expected first connection to be added")
	}

	if manager.Add(testConnection("user-2")) {
		t.Fatal("expected second connection to be rejected")
	}

	if manager.Count() != 1 {
		t.Fatalf("expected count 1, got %d", manager.Count())
	}
}

func TestConnectionManagerAddIsConcurrencySafe(t *testing.T) {
	manager := NewConnectionManager(100)

	var accepted int32
	var wg sync.WaitGroup

	for i := range 250 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			if manager.Add(testConnection("user-1")) {
				atomic.AddInt32(&accepted, 1)
			}
		}(i)
	}

	wg.Wait()

	if accepted != 100 {
		t.Fatalf("expected exactly 100 accepted connections, got %d", accepted)
	}

	if manager.Count() != 100 {
		t.Fatalf("expected manager count 100, got %d", manager.Count())
	}
}

func TestConnectionManagerJoinLeaveRoom(t *testing.T) {
	manager := NewConnectionManager(10)
	conn := testConnection("user-1")

	if !manager.Add(conn) {
		t.Fatal("expected connection to be added")
	}

	if !manager.JoinRoom(conn, "room-1") {
		t.Fatal("expected connection to join room")
	}

	if !manager.IsInRoom(conn, "room-1") {
		t.Fatal("expected connection to be in room")
	}

	if manager.RoomConnectionCount("room-1") != 1 {
		t.Fatalf("expected room connection count 1, got %d", manager.RoomConnectionCount("room-1"))
	}

	manager.LeaveRoom(conn, "room-1")

	if manager.IsInRoom(conn, "room-1") {
		t.Fatal("expected connection to leave room")
	}

	if manager.RoomConnectionCount("room-1") != 0 {
		t.Fatalf("expected room connection count 0, got %d", manager.RoomConnectionCount("room-1"))
	}
}

func TestConnectionManagerBroadcastToRoom(t *testing.T) {
	manager := NewConnectionManager(10)

	firstConn := testConnection("user-1")
	secondConn := testConnection("user-2")
	otherRoomConn := testConnection("user-3")

	manager.Add(firstConn)
	manager.Add(secondConn)
	manager.Add(otherRoomConn)

	manager.JoinRoom(firstConn, "room-1")
	manager.JoinRoom(secondConn, "room-1")
	manager.JoinRoom(otherRoomConn, "room-2")

	dropped := manager.BroadcastToRoom("room-1", []byte("hello"))
	if dropped != 0 {
		t.Fatalf("expected no dropped messages, got %d", dropped)
	}

	assertReceived(t, firstConn, "hello")
	assertReceived(t, secondConn, "hello")
	assertNoMessage(t, otherRoomConn)
}

func TestConnectionManagerBroadcastReportsDroppedMessages(t *testing.T) {
	manager := NewConnectionManager(10)
	conn := &Connection{
		ID:      "conn-1",
		UserID:  "user-1",
		RoomIDs: make(map[string]bool),
		Send:    make(chan []byte, 1),
	}

	manager.Add(conn)
	manager.JoinRoom(conn, "room-1")

	if dropped := manager.BroadcastToRoom("room-1", []byte("first")); dropped != 0 {
		t.Fatalf("expected no dropped messages, got %d", dropped)
	}

	if dropped := manager.BroadcastToRoom("room-1", []byte("second")); dropped != 1 {
		t.Fatalf("expected 1 dropped message, got %d", dropped)
	}
}

func TestConnectionManagerSendToUser(t *testing.T) {
	manager := NewConnectionManager(10)
	conn := testConnection("user-1")

	manager.Add(conn)

	dropped := manager.SendToUser("user-1", []byte("direct"))
	if dropped != 0 {
		t.Fatalf("expected no dropped messages, got %d", dropped)
	}

	assertReceived(t, conn, "direct")
}

func TestConnectionManagerRemoveCleansRoomsAndClosesSendChannel(t *testing.T) {
	manager := NewConnectionManager(10)
	conn := testConnection("user-1")

	manager.Add(conn)
	manager.JoinRoom(conn, "room-1")
	manager.Remove(conn)

	if manager.Count() != 0 {
		t.Fatalf("expected count 0, got %d", manager.Count())
	}

	if manager.RoomConnectionCount("room-1") != 0 {
		t.Fatalf("expected room connection count 0, got %d", manager.RoomConnectionCount("room-1"))
	}

	if _, ok := <-conn.Send; ok {
		t.Fatal("expected send channel to be closed")
	}
}

func TestConnectionManagerShutdownClosesConnections(t *testing.T) {
	manager := NewConnectionManager(10)

	firstConn := testConnection("user-1")
	secondConn := testConnection("user-2")

	manager.Add(firstConn)
	manager.Add(secondConn)
	manager.JoinRoom(firstConn, "room-1")
	manager.JoinRoom(secondConn, "room-1")

	manager.Shutdown()

	if manager.Count() != 0 {
		t.Fatalf("expected count 0 after shutdown, got %d", manager.Count())
	}

	if manager.RoomConnectionCount("room-1") != 0 {
		t.Fatalf("expected room count 0 after shutdown, got %d", manager.RoomConnectionCount("room-1"))
	}

	if _, ok := <-firstConn.Send; ok {
		t.Fatal("expected first send channel to be closed")
	}

	if _, ok := <-secondConn.Send; ok {
		t.Fatal("expected second send channel to be closed")
	}
}

func testConnection(userID string) *Connection {
	return &Connection{
		ID:      userID + "-conn",
		UserID:  userID,
		RoomIDs: make(map[string]bool),
		Send:    make(chan []byte, 8),
	}
}

func assertReceived(t *testing.T, conn *Connection, expected string) {
	t.Helper()

	select {
	case got := <-conn.Send:
		if string(got) != expected {
			t.Fatalf("expected message %q, got %q", expected, string(got))
		}
	default:
		t.Fatalf("expected message %q", expected)
	}
}

func assertNoMessage(t *testing.T, conn *Connection) {
	t.Helper()

	select {
	case got := <-conn.Send:
		t.Fatalf("expected no message, got %q", string(got))
	default:
	}
}

func TestNewConnectionTrimsUserIDAndUsesDefaultBuffer(t *testing.T) {
	conn := NewConnection(" user-1 ", nil, 0)

	if conn.UserID != "user-1" {
		t.Fatalf("expected trimmed user id, got %q", conn.UserID)
	}

	if conn.ID == "" {
		t.Fatal("expected generated connection id")
	}

	if conn.RoomIDs == nil {
		t.Fatal("expected room ids map")
	}

	if conn.Send == nil {
		t.Fatal("expected send channel")
	}

	if cap(conn.Send) != DefaultSendBuffer {
		t.Fatalf("expected default send buffer %d, got %d", DefaultSendBuffer, cap(conn.Send))
	}
}

func TestNewConnectionManagerUsesDefaultMaxConnections(t *testing.T) {
	manager := NewConnectionManager(0)

	if manager.maxConnections != 1000 {
		t.Fatalf("expected default max connections 1000, got %d", manager.maxConnections)
	}
}

func TestConnectionManagerRejectsNilAndEmptyUserConnection(t *testing.T) {
	manager := NewConnectionManager(10)

	if manager.Add(nil) {
		t.Fatal("expected nil connection to be rejected")
	}

	if manager.Add(&Connection{UserID: " "}) {
		t.Fatal("expected empty user id connection to be rejected")
	}
}
