package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	chatv1 "github.com/kodokbakar/pylon/gen/pylon/chat/v1"
	gatewaymanager "github.com/kodokbakar/pylon/services/api-gateway/manager"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNormalizeOriginPatternsTrimsAndAddsHostPatterns(t *testing.T) {
	got := normalizeOriginPatterns([]string{
		" http://localhost:3000 ",
		"",
		" http://localhost:5173 ",
	})

	expected := []string{
		"http://localhost:3000",
		"localhost:3000",
		"http://localhost:5173",
		"localhost:5173",
	}

	if len(got) != len(expected) {
		t.Fatalf("expected %d origin patterns, got %#v", len(expected), got)
	}

	for i, pattern := range expected {
		if got[i] != pattern {
			t.Fatalf("expected origin pattern %d to be %q, got %q", i, pattern, got[i])
		}
	}
}

func TestClientMessageTypeToProtoDefaultsToText(t *testing.T) {
	got, err := clientMessageTypeToProto("")
	if err != nil {
		t.Fatalf("convert message type: %v", err)
	}

	if got != chatv1.MessageType_MESSAGE_TYPE_TEXT {
		t.Fatalf("expected text type, got %v", got)
	}
}

func TestClientMessageTypeToProtoSupportsFile(t *testing.T) {
	got, err := clientMessageTypeToProto("file")
	if err != nil {
		t.Fatalf("convert message type: %v", err)
	}

	if got != chatv1.MessageType_MESSAGE_TYPE_FILE {
		t.Fatalf("expected file type, got %v", got)
	}
}

func TestClientMessageTypeToProtoRejectsUnsupportedType(t *testing.T) {
	_, err := clientMessageTypeToProto("video")
	if err == nil {
		t.Fatal("expected unsupported message type error")
	}
}

func TestTypingEnvelopeUsesConnectionUsername(t *testing.T) {
	got := typingEnvelope(&gatewaymanager.Connection{
		UserID:   "user-1",
		Username: "alice",
	}, "room-1")

	if got.Type != "typing" {
		t.Fatalf("expected typing type, got %q", got.Type)
	}

	if got.RoomID != "room-1" {
		t.Fatalf("expected room-1, got %q", got.RoomID)
	}

	if got.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", got.UserID)
	}

	if got.Username != "alice" {
		t.Fatalf("expected username alice, got %q", got.Username)
	}
}

func TestTypingEnvelopeDoesNotUseUserIDAsUsernameWhenUsernameMissing(t *testing.T) {
	got := typingEnvelope(&gatewaymanager.Connection{
		UserID: "user-1",
	}, "room-1")

	if got.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", got.UserID)
	}

	if got.Username != "" {
		t.Fatalf("expected empty username, got %q", got.Username)
	}
}

func TestUserJoinedEnvelopeUsesUserJoinedType(t *testing.T) {
	got := userJoinedEnvelope(&gatewaymanager.Connection{
		UserID:   "user-1",
		Username: "alice",
	}, "room-1")

	if got.Type != "user_joined" {
		t.Fatalf("expected user_joined type, got %q", got.Type)
	}

	if got.RoomID != "room-1" {
		t.Fatalf("expected room-1, got %q", got.RoomID)
	}

	if got.User == nil {
		t.Fatal("expected user payload")
	}

	if got.User.ID != "user-1" {
		t.Fatalf("expected user-1, got %q", got.User.ID)
	}

	if got.UserID != "" {
		t.Fatalf("expected empty top-level user_id, got %q", got.UserID)
	}

	if got.Username != "" {
		t.Fatalf("expected empty top-level username, got %q", got.Username)
	}
}

type fakeWebSocketChatClient struct {
	sendMessageFunc func(context.Context, *connect.Request[chatv1.SendMessageRequest]) (*connect.Response[chatv1.SendMessageResponse], error)
	getMessagesFunc func(context.Context, *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error)
}

func (c fakeWebSocketChatClient) SendMessage(
	ctx context.Context,
	req *connect.Request[chatv1.SendMessageRequest],
) (*connect.Response[chatv1.SendMessageResponse], error) {
	if c.sendMessageFunc != nil {
		return c.sendMessageFunc(ctx, req)
	}

	return connect.NewResponse(&chatv1.SendMessageResponse{
		Message: &chatv1.Message{
			Id:       "message-1",
			RoomId:   req.Msg.GetRoomId(),
			SenderId: req.Msg.GetSenderId(),
			Content:  req.Msg.GetContent(),
			Type:     req.Msg.GetType(),
		},
	}), nil
}

func (c fakeWebSocketChatClient) GetMessages(
	ctx context.Context,
	req *connect.Request[chatv1.GetMessagesRequest],
) (*connect.Response[chatv1.GetMessagesResponse], error) {
	if c.getMessagesFunc != nil {
		return c.getMessagesFunc(ctx, req)
	}

	return connect.NewResponse(&chatv1.GetMessagesResponse{}), nil
}

func TestNewWebSocketHandlerRequiresChatClient(t *testing.T) {
	_, err := NewWebSocketHandler(10, nil, false, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewWebSocketHandlerInitializesManagerAndOrigins(t *testing.T) {
	handler, err := NewWebSocketHandler(
		10,
		[]string{"http://localhost:5173"},
		false,
		fakeWebSocketChatClient{},
	)
	if err != nil {
		t.Fatalf("create websocket handler: %v", err)
	}

	if handler.manager == nil {
		t.Fatal("expected manager")
	}

	if handler.chatClient == nil {
		t.Fatal("expected chat client")
	}

	if len(handler.originPatterns) != 2 {
		t.Fatalf("expected origin plus host pattern, got %#v", handler.originPatterns)
	}
}

func TestWebSocketHandlerShutdownClosesManagerConnections(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})

	handler.Shutdown()

	if handler.manager.Count() != 0 {
		t.Fatalf("expected manager count 0, got %d", handler.manager.Count())
	}

	assertClosedSendChannel(t, conn)
}

func TestHandleClientMessageRejectsInvalidJSON(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{invalid-json`)); err != nil {
		t.Fatalf("handle message: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Code != "INVALID_JSON" {
		t.Fatalf("expected INVALID_JSON, got %#v", got)
	}
}

func TestHandleClientMessageRespondsToPing(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{"type":"ping"}`)); err != nil {
		t.Fatalf("handle ping: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Type != "pong" {
		t.Fatalf("expected pong, got %#v", got)
	}
}

func TestHandleClientMessageRejectsUnknownType(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{"type":"unknown"}`)); err != nil {
		t.Fatalf("handle message: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Code != "UNKNOWN_TYPE" {
		t.Fatalf("expected UNKNOWN_TYPE, got %#v", got)
	}
}

func TestWebSocketHandleJoinBroadcastsUserJoined(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{
		getMessagesFunc: func(ctx context.Context, req *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error) {
			if req.Msg.GetRoomId() != "room-1" {
				t.Fatalf("expected room-1, got %q", req.Msg.GetRoomId())
			}

			if req.Msg.GetUserId() != "user-1" {
				t.Fatalf("expected user-1, got %q", req.Msg.GetUserId())
			}

			if req.Msg.GetLimit() != 1 {
				t.Fatalf("expected access check limit 1, got %d", req.Msg.GetLimit())
			}

			return connect.NewResponse(&chatv1.GetMessagesResponse{}), nil
		},
	})

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{"type":"join","room_id":" room-1 "}`)); err != nil {
		t.Fatalf("handle join: %v", err)
	}

	if !handler.manager.IsInRoom(conn, "room-1") {
		t.Fatal("expected connection to join room")
	}

	got := readServerEnvelope(t, conn)
	if got.Type != "user_joined" {
		t.Fatalf("expected user_joined, got %#v", got)
	}

	if got.User == nil || got.User.ID != "user-1" {
		t.Fatalf("expected joined user payload, got %#v", got.User)
	}
}

func TestWebSocketHandleJoinMapsAccessError(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{
		getMessagesFunc: func(ctx context.Context, req *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error) {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("not a room member"))
		},
	})

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{"type":"join","room_id":"room-1"}`)); err != nil {
		t.Fatalf("handle join: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Code != "FORBIDDEN" {
		t.Fatalf("expected FORBIDDEN, got %#v", got)
	}
}

func TestWebSocketHandleLeaveRequiresJoinedRoom(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{"type":"leave","room_id":"room-1"}`)); err != nil {
		t.Fatalf("handle leave: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Code != "NOT_JOINED" {
		t.Fatalf("expected NOT_JOINED, got %#v", got)
	}
}

func TestWebSocketHandleLeaveBroadcastsUserLeft(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})

	otherConn := gatewaymanager.NewConnection("user-2", nil, 8)
	otherConn.Username = "bob"
	if !handler.manager.Add(otherConn) {
		t.Fatal("expected other connection to be added")
	}

	if !handler.manager.JoinRoom(conn, "room-1") {
		t.Fatal("expected leaving connection to join room setup")
	}

	if !handler.manager.JoinRoom(otherConn, "room-1") {
		t.Fatal("expected other connection to join room setup")
	}

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{"type":"leave","room_id":"room-1"}`)); err != nil {
		t.Fatalf("handle leave: %v", err)
	}

	if handler.manager.IsInRoom(conn, "room-1") {
		t.Fatal("expected leaving connection to leave room")
	}

	if !handler.manager.IsInRoom(otherConn, "room-1") {
		t.Fatal("expected other connection to stay in room")
	}

	got := readServerEnvelope(t, otherConn)
	if got.Type != "user_left" {
		t.Fatalf("expected user_left, got %#v", got)
	}

	if got.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", got.UserID)
	}
}

func TestWebSocketHandleMessageRequiresJoinedRoom(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{
		"type":"message",
		"room_id":"room-1",
		"content":"hello"
	}`)); err != nil {
		t.Fatalf("handle message: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Code != "NOT_JOINED" {
		t.Fatalf("expected NOT_JOINED, got %#v", got)
	}
}

func TestWebSocketHandleMessageBroadcastsChatMessage(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)

	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{
		sendMessageFunc: func(ctx context.Context, req *connect.Request[chatv1.SendMessageRequest]) (*connect.Response[chatv1.SendMessageResponse], error) {
			if req.Msg.GetRoomId() != "room-1" {
				t.Fatalf("expected room-1, got %q", req.Msg.GetRoomId())
			}

			if req.Msg.GetSenderId() != "user-1" {
				t.Fatalf("expected user-1, got %q", req.Msg.GetSenderId())
			}

			if req.Msg.GetContent() != "hello" {
				t.Fatalf("expected hello, got %q", req.Msg.GetContent())
			}

			if req.Msg.GetType() != chatv1.MessageType_MESSAGE_TYPE_FILE {
				t.Fatalf("expected file message type, got %v", req.Msg.GetType())
			}

			return connect.NewResponse(&chatv1.SendMessageResponse{
				Message: &chatv1.Message{
					Id:                "message-1",
					RoomId:            "room-1",
					SenderId:          "user-1",
					SenderUsername:    "alice",
					SenderDisplayName: "Alice",
					SenderAvatarUrl:   "https://example.com/avatar.png",
					Content:           "hello",
					Type:              chatv1.MessageType_MESSAGE_TYPE_FILE,
					CreatedAt:         timestamppb.New(now),
				},
			}), nil
		},
	})
	if !handler.manager.JoinRoom(conn, "room-1") {
		t.Fatal("expected join room setup to succeed")
	}

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{
		"type":"message",
		"room_id":" room-1 ",
		"content":" hello ",
		"msg_type":"file"
	}`)); err != nil {
		t.Fatalf("handle message: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Type != "message" {
		t.Fatalf("expected message envelope, got %#v", got)
	}

	if got.MessageID != "message-1" {
		t.Fatalf("expected message-1, got %q", got.MessageID)
	}

	if got.Sender == nil || got.Sender.Username != "alice" {
		t.Fatalf("expected sender username alice, got %#v", got.Sender)
	}

	if got.MsgType != "file" {
		t.Fatalf("expected file msg_type, got %q", got.MsgType)
	}

	if got.CreatedAt == "" {
		t.Fatal("expected created_at")
	}
}

func TestWebSocketHandleMessageMapsChatServiceError(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{
		sendMessageFunc: func(ctx context.Context, req *connect.Request[chatv1.SendMessageRequest]) (*connect.Response[chatv1.SendMessageResponse], error) {
			return nil, connect.NewError(connect.CodeUnavailable, errors.New("pod unavailable"))
		},
	})
	if !handler.manager.JoinRoom(conn, "room-1") {
		t.Fatal("expected join room setup to succeed")
	}

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{
		"type":"message",
		"room_id":"room-1",
		"content":"hello"
	}`)); err != nil {
		t.Fatalf("handle message: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Code != "CHAT_SERVICE_UNAVAILABLE" {
		t.Fatalf("expected CHAT_SERVICE_UNAVAILABLE, got %#v", got)
	}
}

func TestWebSocketHandleTypingBroadcastsTyping(t *testing.T) {
	handler, conn := newTestWebSocketHandler(t, fakeWebSocketChatClient{})
	if !handler.manager.JoinRoom(conn, "room-1") {
		t.Fatal("expected join room setup to succeed")
	}

	if err := handler.handleClientMessage(context.Background(), conn, []byte(`{
		"type":"typing",
		"room_id":"room-1"
	}`)); err != nil {
		t.Fatalf("handle typing: %v", err)
	}

	got := readServerEnvelope(t, conn)
	if got.Type != "typing" {
		t.Fatalf("expected typing, got %#v", got)
	}

	if got.UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", got.UserID)
	}
}

func TestSendJSONReturnsErrorWhenSendBufferIsFull(t *testing.T) {
	conn := &gatewaymanager.Connection{
		UserID: "user-1",
		Send:   make(chan []byte),
	}

	err := sendJSON(context.Background(), conn, serverEnvelope{Type: "pong"})
	if err == nil {
		t.Fatal("expected send buffer full error")
	}

	if !strings.Contains(err.Error(), "websocket send buffer is full") {
		t.Fatalf("expected send buffer error, got %v", err)
	}
}

func TestSendJSONReturnsNilWhenContextDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	conn := &gatewaymanager.Connection{
		UserID: "user-1",
		Send:   make(chan []byte),
	}

	if err := sendJSON(ctx, conn, serverEnvelope{Type: "pong"}); err != nil {
		t.Fatalf("expected nil error for canceled context, got %v", err)
	}
}

func TestMessageToEnvelopeHandlesNilCreatedAt(t *testing.T) {
	got := messageToEnvelope(&chatv1.Message{
		Id:       "message-1",
		RoomId:   "room-1",
		SenderId: "user-1",
		Content:  "hello",
		Type:     chatv1.MessageType_MESSAGE_TYPE_TEXT,
	})

	if got.Type != "message" {
		t.Fatalf("expected message type, got %q", got.Type)
	}

	if got.CreatedAt != "" {
		t.Fatalf("expected empty created_at, got %q", got.CreatedAt)
	}
}

func TestProtoMessageTypeToClientSupportsAllNonTextTypes(t *testing.T) {
	tests := []struct {
		name  string
		input chatv1.MessageType
		want  string
	}{
		{name: "image", input: chatv1.MessageType_MESSAGE_TYPE_IMAGE, want: "image"},
		{name: "system", input: chatv1.MessageType_MESSAGE_TYPE_SYSTEM, want: "system"},
		{name: "file", input: chatv1.MessageType_MESSAGE_TYPE_FILE, want: "file"},
		{name: "unspecified", input: chatv1.MessageType_MESSAGE_TYPE_UNSPECIFIED, want: "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := protoMessageTypeToClient(tt.input); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestWebSocketErrorEnvelopeMapsConnectErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code string
	}{
		{
			name: "unauthenticated",
			err:  connect.NewError(connect.CodeUnauthenticated, errors.New("bad token")),
			code: "UNAUTHORIZED",
		},
		{
			name: "invalid argument",
			err:  connect.NewError(connect.CodeInvalidArgument, errors.New("bad message")),
			code: "BAD_REQUEST",
		},
		{
			name: "generic connect",
			err:  connect.NewError(connect.CodeInternal, errors.New("internal chat error")),
			code: "CHAT_SERVICE_ERROR",
		},
		{
			name: "plain error",
			err:  errors.New("plain failure"),
			code: "INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := websocketErrorEnvelope(tt.err)
			if got.Code != tt.code {
				t.Fatalf("expected code %q, got %#v", tt.code, got)
			}

			if got.Type != "error" {
				t.Fatalf("expected error envelope type, got %q", got.Type)
			}
		})
	}
}

func newTestWebSocketHandler(t *testing.T, chatClient fakeWebSocketChatClient) (*WebSocketHandler, *gatewaymanager.Connection) {
	t.Helper()

	handler, err := NewWebSocketHandler(10, nil, true, chatClient)
	if err != nil {
		t.Fatalf("create websocket handler: %v", err)
	}

	conn := gatewaymanager.NewConnection("user-1", nil, 8)
	conn.Username = "alice"
	if !handler.manager.Add(conn) {
		t.Fatal("expected connection to be added")
	}

	return handler, conn
}

func readServerEnvelope(t *testing.T, conn *gatewaymanager.Connection) serverEnvelope {
	t.Helper()

	select {
	case payload := <-conn.Send:
		var envelope serverEnvelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			t.Fatalf("decode server envelope %q: %v", string(payload), err)
		}
		return envelope
	default:
		t.Fatal("expected websocket server envelope")
		return serverEnvelope{}
	}
}

func assertClosedSendChannel(t *testing.T, conn *gatewaymanager.Connection) {
	t.Helper()

	select {
	case _, ok := <-conn.Send:
		if ok {
			t.Fatal("expected send channel to be closed")
		}
	default:
		t.Fatal("expected send channel to be closed")
	}
}
