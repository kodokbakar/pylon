package apigateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"

	"github.com/kodokbakar/pylon/internal/config"
)

func TestHealthCheckReturnsOK(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected health response body, got %q", rec.Body.String())
	}
}

func TestProtectedEndpointBlocksUnauthenticatedRequest(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestProtectedEndpointAllowsAuthenticatedRequest(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-123", "test-secret"))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status 501, got %d", rec.Code)
	}
}

func TestAuthEndpointIsReachable(t *testing.T) {
	server := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected status 501, got %d", rec.Code)
	}
}

func TestWebSocketRejectsUnknownOrigin(t *testing.T) {
	server := newTestServer(t)

	testServer := httptest.NewServer(server.Handler())
	defer testServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws"

	conn, resp, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + testJWT(t, "user-123", "test-secret")},
			"Origin":        []string{"http://evil.example.com"},
		},
	})
	if conn != nil {
		_ = conn.Close(websocket.StatusNormalClosure, "test cleanup")
	}

	if err == nil {
		t.Fatal("expected websocket dial error, got nil")
	}

	if resp == nil {
		t.Fatal("expected websocket response, got nil")
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", resp.StatusCode)
	}
}

func TestWebSocketAcceptsAllowedOriginAndEchoesMessage(t *testing.T) {
	server := newTestServer(t)

	testServer := httptest.NewServer(server.Handler())
	defer testServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/ws"

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Authorization": []string{"Bearer " + testJWT(t, "user-123", "test-secret")},
			"Origin":        []string{"http://localhost:5173"},
		},
	})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() {
		_ = conn.CloseNow()
	}()

	if err := conn.Write(ctx, websocket.MessageText, []byte("hello")); err != nil {
		t.Fatalf("write websocket message: %v", err)
	}

	messageType, payload, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read websocket message: %v", err)
	}

	if messageType != websocket.MessageText {
		t.Fatalf("expected text message, got %v", messageType)
	}

	if string(payload) != "hello" {
		t.Fatalf("expected echoed payload hello, got %q", string(payload))
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	server, err := New(&config.Config{
		App: config.AppConfig{
			Env:         "test",
			Port:        "0",
			LogLevel:    "debug",
			CORSOrigins: []string{"http://localhost:5173"},
		},
		JWT: config.JWTConfig{
			Secret:        "test-secret",
			Expiry:        time.Hour,
			RefreshExpiry: 24 * time.Hour,
		},
		WebSocket: config.WebSocketConfig{
			MaxConnections:     100,
			ReadBufferSize:     4096,
			WriteBufferSize:    4096,
			InsecureSkipVerify: false,
		},
		Services: config.ServicesConfig{
			ChatURL:         "http://localhost:9001",
			PresenceURL:     "http://localhost:9002",
			RoomURL:         "http://localhost:9003",
			NotificationURL: "http://localhost:9004",
		},
	})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	return server
}

func testJWT(t *testing.T, subject, secret string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": subject,
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return tokenString
}
