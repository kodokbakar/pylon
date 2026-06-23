//go:build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/coder/websocket"

	chatv1 "github.com/kodokbakar/pylon/gen/pylon/chat/v1"
	chatv1connect "github.com/kodokbakar/pylon/gen/pylon/chat/v1/chatv1connect"
	notificationv1 "github.com/kodokbakar/pylon/gen/pylon/notification/v1"
	notificationv1connect "github.com/kodokbakar/pylon/gen/pylon/notification/v1/notificationv1connect"
	presencev1 "github.com/kodokbakar/pylon/gen/pylon/presence/v1"
	presencev1connect "github.com/kodokbakar/pylon/gen/pylon/presence/v1/presencev1connect"
	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
)

type apiEnvelope[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data"`
	Error   *apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type testUser struct {
	ID           string
	Username     string
	Email        string
	Token        string
	RefreshToken string
}

type authData struct {
	User         userData `json:"user"`
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token"`
}

type userData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type createRoomData struct {
	Room roomData `json:"room"`
}

type listRoomsData struct {
	Rooms []roomData `json:"rooms"`
}

type roomData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type websocketEnvelope struct {
	Type      string `json:"type"`
	Code      string `json:"code"`
	RoomID    string `json:"room_id"`
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
	UserID    string `json:"user_id"`
}

func TestIntegrationEndToEndFlows(t *testing.T) {
	suite := setupIntegrationSuite(t)
	ctx := context.Background()

	roomClient := roomv1connect.NewRoomServiceClient(suite.HTTPClient, suite.RoomURL)
	chatClient := chatv1connect.NewChatServiceClient(suite.HTTPClient, suite.ChatURL)
	presenceClient := presencev1connect.NewPresenceServiceClient(suite.HTTPClient, suite.PresenceURL)
	notificationClient := notificationv1connect.NewNotificationServiceClient(suite.HTTPClient, suite.NotificationURL)

	user1 := suite.registerUser(t, "integration-user-1")
	user1Login := suite.loginUser(t, user1.Email, "password123")
	if user1Login.Token == "" {
		t.Fatal("expected login token")
	}
	user1.Token = user1Login.Token

	user2 := suite.registerUser(t, "integration-user-2")

	t.Run("auth flow protects rooms endpoint", func(t *testing.T) {
		_ = suite.listRooms(t, user1.Token)
	})

	var room roomData

	t.Run("room creation joining and member verification", func(t *testing.T) {
		room = suite.createRoom(t, user1.Token, "Integration Channel", "channel", nil)
		if room.ID == "" {
			t.Fatal("expected room id")
		}

		suite.joinRoom(t, user2.Token, room.ID)

		members, err := roomClient.GetRoomMembers(ctx, connect.NewRequest(&roomv1.GetRoomMembersRequest{
			RoomId: room.ID,
		}))
		if err != nil {
			t.Fatalf("get room members: %v", err)
		}

		if len(members.Msg.GetMembers()) != 2 {
			t.Fatalf("expected 2 room members, got %d", len(members.Msg.GetMembers()))
		}
	})

	t.Run("presence tracking works", func(t *testing.T) {
		_, err := presenceClient.SetOnline(ctx, connect.NewRequest(&presencev1.SetOnlineRequest{
			UserId: user2.ID,
			RoomId: room.ID,
		}))
		if err != nil {
			t.Fatalf("set online: %v", err)
		}

		presence, err := presenceClient.GetPresence(ctx, connect.NewRequest(&presencev1.GetPresenceRequest{
			UserId: user2.ID,
		}))
		if err != nil {
			t.Fatalf("get presence: %v", err)
		}

		if presence.Msg.GetStatus() != presencev1.PresenceStatus_PRESENCE_STATUS_ONLINE {
			t.Fatalf("expected online status, got %v", presence.Msg.GetStatus())
		}
	})

	t.Run("websocket messaging persists message and creates notification", func(t *testing.T) {
		wsUser2 := suite.openWebSocket(t, user2.Token)
		defer closeWebSocket(t, wsUser2)

		suite.writeWebSocketJSON(t, wsUser2, map[string]any{
			"type":    "join",
			"room_id": room.ID,
		})
		_ = suite.readWebSocketUntil(t, wsUser2, func(envelope websocketEnvelope) bool {
			return envelope.Type == "user_joined"
		})

		wsUser1 := suite.openWebSocket(t, user1.Token)
		defer closeWebSocket(t, wsUser1)

		suite.writeWebSocketJSON(t, wsUser1, map[string]any{
			"type":    "join",
			"room_id": room.ID,
		})
		_ = suite.readWebSocketUntil(t, wsUser1, func(envelope websocketEnvelope) bool {
			return envelope.Type == "user_joined"
		})

		messageContent := "hello from integration websocket"
		suite.writeWebSocketJSON(t, wsUser1, map[string]any{
			"type":    "message",
			"room_id": room.ID,
			"content": messageContent,
		})

		received := suite.readWebSocketUntil(t, wsUser2, func(envelope websocketEnvelope) bool {
			return envelope.Type == "message" && envelope.Content == messageContent
		})

		if received.MessageID == "" {
			t.Fatalf("expected message id in websocket envelope, got %#v", received)
		}

		messages, err := chatClient.GetMessages(ctx, connect.NewRequest(&chatv1.GetMessagesRequest{
			RoomId: room.ID,
			UserId: user1.ID,
			Limit:  10,
		}))
		if err != nil {
			t.Fatalf("get messages: %v", err)
		}

		if !containsMessage(messages.Msg.GetMessages(), messageContent) {
			t.Fatalf("expected persisted message %q, got %#v", messageContent, messages.Msg.GetMessages())
		}

		notifications := suite.waitForNotifications(t, notificationClient, user2.ID)
		if len(notifications.GetNotifications()) == 0 {
			t.Fatal("expected notification for user2")
		}
	})

	t.Run("rate limiting returns 429", func(t *testing.T) {
		ip := fmt.Sprintf("198.51.100.%d", time.Now().UnixNano()%200+1)

		limited := false

		for i := 0; i < 150; i++ {
			resp := suite.doJSONWithHeaders(t, http.MethodPost, "/api/v1/auth/login", map[string]any{
				"email":    fmt.Sprintf("missing-%d@example.com", i),
				"password": "wrong-password",
			}, map[string]string{
				"X-Forwarded-For": ip,
			})
			body := readBodyAndClose(t, resp)

			if resp.StatusCode == http.StatusTooManyRequests {
				if resp.Header.Get("Retry-After") == "" {
					t.Fatal("expected Retry-After header")
				}

				limited = true
				break
			}

			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected status 401 before rate limit, got %d body=%s", resp.StatusCode, body)
			}
		}

		if !limited {
			t.Fatal("expected rate limiter to eventually return 429")
		}
	})
}

func (s *integrationSuite) registerUser(t *testing.T, usernamePrefix string) testUser {
	t.Helper()

	suffix := time.Now().UnixNano()
	username := fmt.Sprintf("%s-%d", usernamePrefix, suffix)
	email := fmt.Sprintf("%s-%d@example.com", usernamePrefix, suffix)

	resp := s.doJSON(t, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username": username,
		"email":    email,
		"password": "password123",
	})
	defer closeResponseBody(t, resp)

	data := decodeAPIResponse[authData](t, resp, http.StatusCreated)
	if data.Token == "" {
		t.Fatal("expected register token")
	}
	if data.RefreshToken == "" {
		t.Fatal("expected register refresh token")
	}
	if data.User.ID == "" {
		t.Fatal("expected registered user id")
	}

	return testUser{
		ID:           data.User.ID,
		Username:     data.User.Username,
		Email:        data.User.Email,
		Token:        data.Token,
		RefreshToken: data.RefreshToken,
	}
}

func (s *integrationSuite) loginUser(t *testing.T, email, password string) testUser {
	t.Helper()

	resp := s.doJSON(t, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email":    email,
		"password": password,
	})
	defer closeResponseBody(t, resp)

	data := decodeAPIResponse[authData](t, resp, http.StatusOK)
	if data.Token == "" {
		t.Fatal("expected login token")
	}

	return testUser{
		ID:           data.User.ID,
		Username:     data.User.Username,
		Email:        data.User.Email,
		Token:        data.Token,
		RefreshToken: data.RefreshToken,
	}
}

func (s *integrationSuite) listRooms(t *testing.T, token string) listRoomsData {
	t.Helper()

	resp := s.doJSONWithHeaders(t, http.MethodGet, "/api/v1/rooms", nil, authHeaders(token))
	defer closeResponseBody(t, resp)

	return decodeAPIResponse[listRoomsData](t, resp, http.StatusOK)
}

func (s *integrationSuite) createRoom(t *testing.T, token, name, roomType string, memberIDs []string) roomData {
	t.Helper()

	resp := s.doJSONWithHeaders(t, http.MethodPost, "/api/v1/rooms", map[string]any{
		"name":       name,
		"type":       roomType,
		"member_ids": memberIDs,
	}, authHeaders(token))
	defer closeResponseBody(t, resp)

	data := decodeAPIResponse[createRoomData](t, resp, http.StatusCreated)
	if data.Room.ID == "" {
		t.Fatalf("expected created room id, got %#v", data)
	}

	return data.Room
}

func (s *integrationSuite) joinRoom(t *testing.T, token, roomID string) {
	t.Helper()

	resp := s.doJSONWithHeaders(t, http.MethodPost, "/api/v1/rooms/"+roomID+"/join", nil, authHeaders(token))
	defer closeResponseBody(t, resp)

	if resp.StatusCode != http.StatusOK {
		body := readBodyAndClose(t, resp)
		t.Fatalf("expected join status 200, got %d body=%s", resp.StatusCode, body)
	}
}

func (s *integrationSuite) openWebSocket(t *testing.T, token string) *websocket.Conn {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(s.GatewayURL, "http") + "/ws?token=" + token

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Origin": []string{"http://localhost:5173"},
		},
	})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}

	return conn
}

func (s *integrationSuite) writeWebSocketJSON(t *testing.T, conn *websocket.Conn, payload any) {
	t.Helper()

	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal websocket payload: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Write(ctx, websocket.MessageText, encoded); err != nil {
		t.Fatalf("write websocket payload: %v", err)
	}
}

func (s *integrationSuite) readWebSocketUntil(
	t *testing.T,
	conn *websocket.Conn,
	match func(websocketEnvelope) bool,
) websocketEnvelope {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		_, payload, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("read websocket payload: %v", err)
		}

		var envelope websocketEnvelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			t.Fatalf("decode websocket payload %q: %v", string(payload), err)
		}

		if match(envelope) {
			return envelope
		}
	}
}

func (s *integrationSuite) waitForNotifications(
	t *testing.T,
	client notificationv1connect.NotificationServiceClient,
	userID string,
) *notificationv1.GetNotificationsResponse {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var lastErr error

	for ctx.Err() == nil {
		res, err := client.GetNotifications(ctx, connect.NewRequest(&notificationv1.GetNotificationsRequest{
			UserId: userID,
			Limit:  10,
		}))
		if err != nil {
			lastErr = err
			time.Sleep(200 * time.Millisecond)
			continue
		}

		if len(res.Msg.GetNotifications()) > 0 {
			return res.Msg
		}

		time.Sleep(200 * time.Millisecond)
	}

	if lastErr != nil {
		t.Fatalf("wait for notifications: %v", lastErr)
	}

	t.Fatal("timeout waiting for notifications")
	return nil
}

func (s *integrationSuite) doJSON(t *testing.T, method, path string, payload any) *http.Response {
	t.Helper()
	return s.doJSONWithHeaders(t, method, path, payload, nil)
}

func (s *integrationSuite) doJSONWithHeaders(
	t *testing.T,
	method string,
	path string,
	payload any,
	headers map[string]string,
) *http.Response {
	t.Helper()

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal request payload: %v", err)
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, s.GatewayURL+path, body)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("send %s %s: %v", method, path, err)
	}

	return resp
}

func decodeAPIResponse[T any](t *testing.T, resp *http.Response, expectedStatus int) T {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d body=%s", expectedStatus, resp.StatusCode, string(body))
	}

	var envelope apiEnvelope[T]
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode api response %q: %v", string(body), err)
	}

	if !envelope.Success {
		t.Fatalf("expected success response, got %#v body=%s", envelope.Error, string(body))
	}

	return envelope.Data
}

func authHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

func containsMessage(messages []*chatv1.Message, content string) bool {
	for _, message := range messages {
		if message.GetContent() == content {
			return true
		}
	}

	return false
}

func closeWebSocket(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	if err := conn.CloseNow(); err != nil {
		t.Logf("close websocket: %v", err)
	}
}

func closeResponseBody(t *testing.T, resp *http.Response) {
	t.Helper()

	if resp == nil || resp.Body == nil {
		return
	}

	if err := resp.Body.Close(); err != nil {
		t.Logf("close response body: %v", err)
	}
}

func readBodyAndClose(t *testing.T, resp *http.Response) string {
	t.Helper()

	if resp == nil || resp.Body == nil {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if err := resp.Body.Close(); err != nil {
		t.Logf("close response body: %v", err)
	}

	return string(body)
}
