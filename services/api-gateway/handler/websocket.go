package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/coder/websocket"

	"github.com/kodokbakar/pylon/internal/response"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
)

type WebSocketHandler struct {
	manager            *ConnectionManager
	originPatterns     []string
	insecureSkipVerify bool
}

type managedConnection interface {
	Close(code websocket.StatusCode, reason string) error
}

type ConnectionManager struct {
	mu             sync.RWMutex
	maxConnections int
	connections    map[string]map[managedConnection]struct{}
}

func NewWebSocketHandler(maxConnections int, originPatterns []string, insecureSkipVerify bool) *WebSocketHandler {
	return &WebSocketHandler{
		manager:            NewConnectionManager(maxConnections),
		originPatterns:     normalizeOriginPatterns(originPatterns),
		insecureSkipVerify: insecureSkipVerify,
	}
}

func NewConnectionManager(maxConnections int) *ConnectionManager {
	if maxConnections <= 0 {
		maxConnections = 1000
	}

	return &ConnectionManager{
		maxConnections: maxConnections,
		connections:    make(map[string]map[managedConnection]struct{}),
	}
}

func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := gatewaymiddleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "user id is required")
		return
	}

	acceptOptions := &websocket.AcceptOptions{
		InsecureSkipVerify: h.insecureSkipVerify,
	}

	if !h.insecureSkipVerify {
		acceptOptions.OriginPatterns = h.originPatterns
	}

	conn, err := websocket.Accept(w, r, acceptOptions)
	if err != nil {
		log.Printf(
			"accept websocket: origin=%q host=%q patterns=%v insecure_skip_verify=%t error=%v",
			r.Header.Get("Origin"),
			r.Host,
			acceptOptions.OriginPatterns,
			acceptOptions.InsecureSkipVerify,
			err,
		)
		return
	}

	if !h.manager.TryAdd(userID, conn) {
		_ = conn.Close(websocket.StatusPolicyViolation, "websocket connection limit reached")
		return
	}

	defer h.manager.Remove(userID, conn)
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	if err := h.readLoop(r.Context(), conn); err != nil {
		_ = conn.Close(websocket.StatusInternalError, err.Error())
	}
}

func (h *WebSocketHandler) Shutdown() {
	h.manager.Shutdown()
}

func (h *WebSocketHandler) readLoop(ctx context.Context, conn *websocket.Conn) error {
	for {
		messageType, payload, err := conn.Read(ctx)
		if err != nil {
			return nil
		}

		if messageType != websocket.MessageText && messageType != websocket.MessageBinary {
			continue
		}

		if err := conn.Write(ctx, messageType, payload); err != nil {
			return fmt.Errorf("write websocket message: %w", err)
		}
	}
}

func (m *ConnectionManager) TryAdd(userID string, conn managedConnection) bool {
	userID = strings.TrimSpace(userID)
	if userID == "" || conn == nil {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.countLocked() >= m.maxConnections {
		return false
	}

	if _, exists := m.connections[userID]; !exists {
		m.connections[userID] = make(map[managedConnection]struct{})
	}

	m.connections[userID][conn] = struct{}{}
	return true
}

func (m *ConnectionManager) Remove(userID string, conn managedConnection) {
	userID = strings.TrimSpace(userID)
	if userID == "" || conn == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	userConnections, exists := m.connections[userID]
	if !exists {
		return
	}

	delete(userConnections, conn)

	if len(userConnections) == 0 {
		delete(m.connections, userID)
	}
}

func (m *ConnectionManager) Shutdown() {
	m.mu.Lock()

	connections := make([]managedConnection, 0, m.countLocked())
	for _, userConnections := range m.connections {
		for conn := range userConnections {
			connections = append(connections, conn)
		}
	}

	m.connections = make(map[string]map[managedConnection]struct{})

	m.mu.Unlock()

	for _, conn := range connections {
		_ = conn.Close(websocket.StatusGoingAway, "server shutting down")
	}
}

func (m *ConnectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.countLocked()
}

func (m *ConnectionManager) countLocked() int {
	total := 0
	for _, userConnections := range m.connections {
		total += len(userConnections)
	}

	return total
}

func normalizeOriginPatterns(patterns []string) []string {
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(patterns))

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		addOriginPattern(&normalized, seen, pattern)

		scheme, host, ok := strings.Cut(pattern, "://")
		if ok {
			_ = scheme
			addOriginPattern(&normalized, seen, host)
		}
	}

	return normalized
}

func addOriginPattern(patterns *[]string, seen map[string]struct{}, pattern string) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return
	}

	if _, exists := seen[pattern]; exists {
		return
	}

	seen[pattern] = struct{}{}
	*patterns = append(*patterns, pattern)
}
