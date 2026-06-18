package manager

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const DefaultSendBuffer = 64

type Connection struct {
	ID      string
	UserID  string
	RoomIDs map[string]bool
	Conn    *websocket.Conn
	Send    chan []byte

	closed bool
}

type ConnectionManager struct {
	mu             sync.RWMutex
	maxConnections int
	connections    map[string]map[*Connection]struct{}
	rooms          map[string]map[*Connection]struct{}
}

func NewConnection(userID string, conn *websocket.Conn, sendBuffer int) *Connection {
	userID = strings.TrimSpace(userID)
	if sendBuffer <= 0 {
		sendBuffer = DefaultSendBuffer
	}

	return &Connection{
		ID:      newConnectionID(),
		UserID:  userID,
		RoomIDs: make(map[string]bool),
		Conn:    conn,
		Send:    make(chan []byte, sendBuffer),
	}
}

func NewConnectionManager(maxConnections int) *ConnectionManager {
	if maxConnections <= 0 {
		maxConnections = 1000
	}

	return &ConnectionManager{
		maxConnections: maxConnections,
		connections:    make(map[string]map[*Connection]struct{}),
		rooms:          make(map[string]map[*Connection]struct{}),
	}
}

func (m *ConnectionManager) Add(conn *Connection) bool {
	if conn == nil || strings.TrimSpace(conn.UserID) == "" {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.countLocked() >= m.maxConnections {
		return false
	}

	conn.UserID = strings.TrimSpace(conn.UserID)
	if conn.ID == "" {
		conn.ID = newConnectionID()
	}
	if conn.RoomIDs == nil {
		conn.RoomIDs = make(map[string]bool)
	}
	if conn.Send == nil {
		conn.Send = make(chan []byte, DefaultSendBuffer)
	}

	if _, exists := m.connections[conn.UserID]; !exists {
		m.connections[conn.UserID] = make(map[*Connection]struct{})
	}
	m.connections[conn.UserID][conn] = struct{}{}

	return true
}

func (m *ConnectionManager) Remove(conn *Connection) {
	if conn == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if userConnections, exists := m.connections[conn.UserID]; exists {
		delete(userConnections, conn)
		if len(userConnections) == 0 {
			delete(m.connections, conn.UserID)
		}
	}

	for roomID := range conn.RoomIDs {
		m.leaveRoomLocked(conn, roomID)
	}

	m.closeSendLocked(conn)
}

func (m *ConnectionManager) JoinRoom(conn *Connection, roomID string) bool {
	if conn == nil {
		return false
	}

	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if conn.closed {
		return false
	}

	if conn.RoomIDs == nil {
		conn.RoomIDs = make(map[string]bool)
	}

	if _, exists := m.rooms[roomID]; !exists {
		m.rooms[roomID] = make(map[*Connection]struct{})
	}

	conn.RoomIDs[roomID] = true
	m.rooms[roomID][conn] = struct{}{}

	return true
}

func (m *ConnectionManager) LeaveRoom(conn *Connection, roomID string) {
	if conn == nil {
		return
	}

	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.leaveRoomLocked(conn, roomID)
}

func (m *ConnectionManager) IsInRoom(conn *Connection, roomID string) bool {
	if conn == nil {
		return false
	}

	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return conn.RoomIDs[roomID]
}

func (m *ConnectionManager) BroadcastToRoom(roomID string, message []byte) int {
	roomID = strings.TrimSpace(roomID)
	if roomID == "" || len(message) == 0 {
		return 0
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	dropped := 0
	for conn := range m.rooms[roomID] {
		if !enqueue(conn, message) {
			dropped++
		}
	}

	return dropped
}

func (m *ConnectionManager) SendToUser(userID string, message []byte) int {
	userID = strings.TrimSpace(userID)
	if userID == "" || len(message) == 0 {
		return 0
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	dropped := 0
	for conn := range m.connections[userID] {
		if !enqueue(conn, message) {
			dropped++
		}
	}

	return dropped
}

func (m *ConnectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.countLocked()
}

func (m *ConnectionManager) UserConnectionCount(userID string) int {
	userID = strings.TrimSpace(userID)

	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.connections[userID])
}

func (m *ConnectionManager) RoomConnectionCount(roomID string) int {
	roomID = strings.TrimSpace(roomID)

	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.rooms[roomID])
}

func (m *ConnectionManager) Shutdown() {
	m.mu.Lock()

	connections := make([]*Connection, 0, m.countLocked())
	for _, userConnections := range m.connections {
		for conn := range userConnections {
			connections = append(connections, conn)
		}
	}

	m.connections = make(map[string]map[*Connection]struct{})
	m.rooms = make(map[string]map[*Connection]struct{})

	for _, conn := range connections {
		conn.RoomIDs = make(map[string]bool)
		m.closeSendLocked(conn)
	}

	m.mu.Unlock()

	for _, conn := range connections {
		if conn.Conn != nil {
			_ = conn.Conn.Close(websocket.StatusGoingAway, "server shutting down")
		}
	}
}

func (m *ConnectionManager) leaveRoomLocked(conn *Connection, roomID string) {
	delete(conn.RoomIDs, roomID)

	roomConnections, exists := m.rooms[roomID]
	if !exists {
		return
	}

	delete(roomConnections, conn)
	if len(roomConnections) == 0 {
		delete(m.rooms, roomID)
	}
}

func (m *ConnectionManager) closeSendLocked(conn *Connection) {
	if conn.closed {
		return
	}

	conn.closed = true
	close(conn.Send)
}

func (m *ConnectionManager) countLocked() int {
	total := 0
	for _, userConnections := range m.connections {
		total += len(userConnections)
	}

	return total
}

func enqueue(conn *Connection, message []byte) bool {
	if conn == nil || conn.closed {
		return false
	}

	select {
	case conn.Send <- message:
		return true
	default:
		return false
	}
}

func newConnectionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}

	return fmt.Sprintf("%d", time.Now().UnixNano())
}
