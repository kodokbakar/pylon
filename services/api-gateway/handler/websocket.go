package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/coder/websocket"

	chatv1 "github.com/kodokbakar/pylon/gen/pylon/chat/v1"
	"github.com/kodokbakar/pylon/internal/response"
	gatewaymanager "github.com/kodokbakar/pylon/services/api-gateway/manager"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
)

const (
	webSocketSendBuffer = 64

	wsTypeJoin    = "join"
	wsTypeLeave   = "leave"
	wsTypeMessage = "message"
	wsTypeTyping  = "typing"
	wsTypePing    = "ping"
)

type ChatServiceClient interface {
	SendMessage(context.Context, *connect.Request[chatv1.SendMessageRequest]) (*connect.Response[chatv1.SendMessageResponse], error)
	GetMessages(context.Context, *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error)
}

type WebSocketHandler struct {
	manager            *gatewaymanager.ConnectionManager
	chatClient         ChatServiceClient
	originPatterns     []string
	insecureSkipVerify bool
}

type clientEnvelope struct {
	Type    string `json:"type"`
	RoomID  string `json:"room_id,omitempty"`
	Content string `json:"content,omitempty"`
	MsgType string `json:"msg_type,omitempty"`
}

type serverEnvelope struct {
	Type      string         `json:"type"`
	MessageID string         `json:"message_id,omitempty"`
	RoomID    string         `json:"room_id,omitempty"`
	Sender    *senderPayload `json:"sender,omitempty"`
	User      *userPayload   `json:"user,omitempty"`
	UserID    string         `json:"user_id,omitempty"`
	Username  string         `json:"username,omitempty"`
	Content   string         `json:"content,omitempty"`
	MsgType   string         `json:"msg_type,omitempty"`
	CreatedAt string         `json:"created_at,omitempty"`
	Code      string         `json:"code,omitempty"`
	Message   string         `json:"message,omitempty"`
}

type senderPayload struct {
	ID          string `json:"id"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type userPayload struct {
	ID string `json:"id"`
}

func NewWebSocketHandler(
	maxConnections int,
	originPatterns []string,
	insecureSkipVerify bool,
	chatClient ChatServiceClient,
) (*WebSocketHandler, error) {
	if chatClient == nil {
		return nil, fmt.Errorf("chat service client is required")
	}

	return &WebSocketHandler{
		manager:            gatewaymanager.NewConnectionManager(maxConnections),
		chatClient:         chatClient,
		originPatterns:     normalizeOriginPatterns(originPatterns),
		insecureSkipVerify: insecureSkipVerify,
	}, nil
}

func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := gatewaymiddleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "user id is required")
		return
	}

	username, _ := gatewaymiddleware.UsernameFromContext(r.Context())
	username = strings.TrimSpace(username)

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

	wsConn := gatewaymanager.NewConnection(userID, conn, webSocketSendBuffer)
	wsConn.Username = username
	if !h.manager.Add(wsConn) {
		_ = conn.Close(websocket.StatusPolicyViolation, "websocket connection limit reached")
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	defer h.manager.Remove(wsConn)
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	writeErrCh := make(chan error, 1)
	go func() {
		writeErrCh <- h.writeLoop(ctx, wsConn)
	}()

	if err := h.readLoop(ctx, wsConn); err != nil {
		_ = sendJSON(ctx, wsConn, errorEnvelope("INTERNAL", err.Error()))
		_ = conn.Close(websocket.StatusInternalError, err.Error())
	}

	cancel()

	select {
	case err := <-writeErrCh:
		if err != nil {
			log.Printf("websocket write loop stopped: user_id=%s connection_id=%s error=%v", userID, wsConn.ID, err)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

func (h *WebSocketHandler) Shutdown() {
	h.manager.Shutdown()
}

func (h *WebSocketHandler) readLoop(ctx context.Context, conn *gatewaymanager.Connection) error {
	for {
		messageType, payload, err := conn.Conn.Read(ctx)
		if err != nil {
			return nil
		}

		if messageType != websocket.MessageText {
			if err := sendJSON(ctx, conn, errorEnvelope("UNSUPPORTED_MESSAGE_TYPE", "only text websocket messages are supported")); err != nil {
				return err
			}
			continue
		}

		if err := h.handleClientMessage(ctx, conn, payload); err != nil {
			return err
		}
	}
}

func (h *WebSocketHandler) writeLoop(ctx context.Context, conn *gatewaymanager.Connection) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case payload, ok := <-conn.Send:
			if !ok {
				return nil
			}

			if err := conn.Conn.Write(ctx, websocket.MessageText, payload); err != nil {
				return fmt.Errorf("write websocket message: %w", err)
			}
		}
	}
}

func (h *WebSocketHandler) handleClientMessage(ctx context.Context, conn *gatewaymanager.Connection, payload []byte) error {
	var msg clientEnvelope
	if err := json.Unmarshal(payload, &msg); err != nil {
		return sendJSON(ctx, conn, errorEnvelope("INVALID_JSON", "invalid websocket message json"))
	}

	msg.Type = strings.TrimSpace(strings.ToLower(msg.Type))

	switch msg.Type {
	case wsTypeJoin:
		return h.handleJoin(ctx, conn, msg)
	case wsTypeLeave:
		return h.handleLeave(ctx, conn, msg)
	case wsTypeMessage:
		return h.handleMessage(ctx, conn, msg)
	case wsTypeTyping:
		return h.handleTyping(ctx, conn, msg)
	case wsTypePing:
		return sendJSON(ctx, conn, serverEnvelope{Type: "pong"})
	default:
		return sendJSON(ctx, conn, errorEnvelope("UNKNOWN_TYPE", "unknown websocket message type"))
	}
}

func (h *WebSocketHandler) handleJoin(ctx context.Context, conn *gatewaymanager.Connection, msg clientEnvelope) error {
	roomID := strings.TrimSpace(msg.RoomID)
	if roomID == "" {
		return sendJSON(ctx, conn, errorEnvelope("BAD_REQUEST", "room_id is required"))
	}

	if err := h.ensureRoomAccess(ctx, conn.UserID, roomID); err != nil {
		return sendJSON(ctx, conn, websocketErrorEnvelope(err))
	}

	if !h.manager.JoinRoom(conn, roomID) {
		return sendJSON(ctx, conn, errorEnvelope("BAD_REQUEST", "failed to join room"))
	}

	payload, err := json.Marshal(typingEnvelope(conn, roomID))
	if err != nil {
		return fmt.Errorf("marshal user joined event: %w", err)
	}

	if dropped := h.manager.BroadcastToRoom(roomID, payload); dropped > 0 {
		log.Printf("drop user_joined websocket event: room_id=%s dropped=%d", roomID, dropped)
	}

	return nil
}

func (h *WebSocketHandler) handleLeave(ctx context.Context, conn *gatewaymanager.Connection, msg clientEnvelope) error {
	roomID := strings.TrimSpace(msg.RoomID)
	if roomID == "" {
		return sendJSON(ctx, conn, errorEnvelope("BAD_REQUEST", "room_id is required"))
	}

	if !h.manager.IsInRoom(conn, roomID) {
		return sendJSON(ctx, conn, errorEnvelope("NOT_JOINED", "connection has not joined this room"))
	}

	h.manager.LeaveRoom(conn, roomID)

	payload, err := json.Marshal(serverEnvelope{
		Type:   "user_left",
		RoomID: roomID,
		UserID: conn.UserID,
	})
	if err != nil {
		return fmt.Errorf("marshal user left event: %w", err)
	}

	if dropped := h.manager.BroadcastToRoom(roomID, payload); dropped > 0 {
		log.Printf("drop user_left websocket event: room_id=%s dropped=%d", roomID, dropped)
	}

	return nil
}

func (h *WebSocketHandler) handleMessage(ctx context.Context, conn *gatewaymanager.Connection, msg clientEnvelope) error {
	roomID := strings.TrimSpace(msg.RoomID)
	content := strings.TrimSpace(msg.Content)

	if roomID == "" {
		return sendJSON(ctx, conn, errorEnvelope("BAD_REQUEST", "room_id is required"))
	}

	if content == "" {
		return sendJSON(ctx, conn, errorEnvelope("BAD_REQUEST", "content is required"))
	}

	if !h.manager.IsInRoom(conn, roomID) {
		return sendJSON(ctx, conn, errorEnvelope("NOT_JOINED", "join room before sending messages"))
	}

	messageType, err := clientMessageTypeToProto(msg.MsgType)
	if err != nil {
		return sendJSON(ctx, conn, errorEnvelope("BAD_REQUEST", err.Error()))
	}

	resp, err := h.chatClient.SendMessage(ctx, connect.NewRequest(&chatv1.SendMessageRequest{
		RoomId:   roomID,
		SenderId: conn.UserID,
		Content:  content,
		Type:     messageType,
	}))
	if err != nil {
		return sendJSON(ctx, conn, websocketErrorEnvelope(err))
	}

	if resp.Msg.GetMessage() == nil {
		return sendJSON(ctx, conn, errorEnvelope("CHAT_SERVICE_ERROR", "chat service returned empty message"))
	}

	payload, err := json.Marshal(messageToEnvelope(resp.Msg.GetMessage()))
	if err != nil {
		return fmt.Errorf("marshal message event: %w", err)
	}

	if dropped := h.manager.BroadcastToRoom(roomID, payload); dropped > 0 {
		log.Printf("drop websocket message event: room_id=%s dropped=%d", roomID, dropped)
	}

	return nil
}

func (h *WebSocketHandler) handleTyping(ctx context.Context, conn *gatewaymanager.Connection, msg clientEnvelope) error {
	roomID := strings.TrimSpace(msg.RoomID)
	if roomID == "" {
		return sendJSON(ctx, conn, errorEnvelope("BAD_REQUEST", "room_id is required"))
	}

	if !h.manager.IsInRoom(conn, roomID) {
		return sendJSON(ctx, conn, errorEnvelope("NOT_JOINED", "join room before sending typing events"))
	}

	payload, err := json.Marshal(serverEnvelope{
		Type:     "typing",
		RoomID:   roomID,
		UserID:   conn.UserID,
		Username: conn.UserID,
	})
	if err != nil {
		return fmt.Errorf("marshal typing event: %w", err)
	}

	if dropped := h.manager.BroadcastToRoom(roomID, payload); dropped > 0 {
		log.Printf("drop typing websocket event: room_id=%s dropped=%d", roomID, dropped)
	}

	return nil
}

func (h *WebSocketHandler) ensureRoomAccess(ctx context.Context, userID, roomID string) error {
	_, err := h.chatClient.GetMessages(ctx, connect.NewRequest(&chatv1.GetMessagesRequest{
		RoomId: roomID,
		UserId: userID,
		Limit:  1,
	}))
	if err != nil {
		return fmt.Errorf("validate room access: %w", err)
	}

	return nil
}

func sendJSON(ctx context.Context, conn *gatewaymanager.Connection, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal websocket payload: %w", err)
	}

	select {
	case conn.Send <- data:
		return nil
	case <-ctx.Done():
		return nil
	default:
		return fmt.Errorf("websocket send buffer is full")
	}
}

func typingEnvelope(conn *gatewaymanager.Connection, roomID string) serverEnvelope {
	if conn == nil {
		return serverEnvelope{
			Type:   "typing",
			RoomID: roomID,
		}
	}

	return serverEnvelope{
		Type:     "typing",
		RoomID:   roomID,
		UserID:   conn.UserID,
		Username: strings.TrimSpace(conn.Username),
	}
}

func messageToEnvelope(msg *chatv1.Message) serverEnvelope {
	createdAt := ""
	if msg.GetCreatedAt() != nil {
		createdAt = msg.GetCreatedAt().AsTime().Format(time.RFC3339Nano)
	}

	return serverEnvelope{
		Type:      "message",
		MessageID: msg.GetId(),
		RoomID:    msg.GetRoomId(),
		Sender: &senderPayload{
			ID:          msg.GetSenderId(),
			Username:    msg.GetSenderUsername(),
			DisplayName: msg.GetSenderDisplayName(),
			AvatarURL:   msg.GetSenderAvatarUrl(),
		},
		Content:   msg.GetContent(),
		MsgType:   protoMessageTypeToClient(msg.GetType()),
		CreatedAt: createdAt,
	}
}

func clientMessageTypeToProto(messageType string) (chatv1.MessageType, error) {
	messageType = strings.TrimSpace(strings.ToLower(messageType))
	if messageType == "" {
		return chatv1.MessageType_MESSAGE_TYPE_TEXT, nil
	}

	switch messageType {
	case "text":
		return chatv1.MessageType_MESSAGE_TYPE_TEXT, nil
	case "image":
		return chatv1.MessageType_MESSAGE_TYPE_IMAGE, nil
	case "system":
		return chatv1.MessageType_MESSAGE_TYPE_SYSTEM, nil
	case "file":
		return chatv1.MessageType_MESSAGE_TYPE_FILE, nil
	default:
		return chatv1.MessageType_MESSAGE_TYPE_UNSPECIFIED, fmt.Errorf("unsupported msg_type %q", messageType)
	}
}

func protoMessageTypeToClient(messageType chatv1.MessageType) string {
	switch messageType {
	case chatv1.MessageType_MESSAGE_TYPE_IMAGE:
		return "image"
	case chatv1.MessageType_MESSAGE_TYPE_SYSTEM:
		return "system"
	case chatv1.MessageType_MESSAGE_TYPE_FILE:
		return "file"
	default:
		return "text"
	}
}

func websocketErrorEnvelope(err error) serverEnvelope {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		switch connectErr.Code() {
		case connect.CodeUnauthenticated:
			return errorEnvelope("UNAUTHORIZED", "unauthorized")
		case connect.CodePermissionDenied:
			return errorEnvelope("FORBIDDEN", "forbidden")
		case connect.CodeInvalidArgument:
			return errorEnvelope("BAD_REQUEST", connectErr.Message())
		case connect.CodeUnavailable:
			return errorEnvelope("CHAT_SERVICE_UNAVAILABLE", "chat service unavailable")
		default:
			return errorEnvelope("CHAT_SERVICE_ERROR", connectErr.Message())
		}
	}

	return errorEnvelope("INTERNAL", err.Error())
}

func errorEnvelope(code, message string) serverEnvelope {
	return serverEnvelope{
		Type:    "error",
		Code:    code,
		Message: message,
	}
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

		_, host, ok := strings.Cut(pattern, "://")
		if ok {
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
