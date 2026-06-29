package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	chatv1 "github.com/kodokbakar/pylon/gen/pylon/chat/v1"
	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	roomv1connect "github.com/kodokbakar/pylon/gen/pylon/room/v1/roomv1connect"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
	"google.golang.org/protobuf/types/known/emptypb"
)

const testJWTSecret = "test-secret"

type fakeRoomClient struct {
	createRoomFunc     func(context.Context, *connect.Request[roomv1.CreateRoomRequest]) (*connect.Response[roomv1.CreateRoomResponse], error)
	getRoomFunc        func(context.Context, *connect.Request[roomv1.GetRoomRequest]) (*connect.Response[roomv1.GetRoomResponse], error)
	listRoomsFunc      func(context.Context, *connect.Request[roomv1.ListRoomsRequest]) (*connect.Response[roomv1.ListRoomsResponse], error)
	joinRoomFunc       func(context.Context, *connect.Request[roomv1.JoinRoomRequest]) (*connect.Response[emptypb.Empty], error)
	leaveRoomFunc      func(context.Context, *connect.Request[roomv1.LeaveRoomRequest]) (*connect.Response[emptypb.Empty], error)
	getRoomMembersFunc func(context.Context, *connect.Request[roomv1.GetRoomMembersRequest]) (*connect.Response[roomv1.GetRoomMembersResponse], error)
}

func (c fakeRoomClient) CreateRoom(
	ctx context.Context,
	req *connect.Request[roomv1.CreateRoomRequest],
) (*connect.Response[roomv1.CreateRoomResponse], error) {
	if c.createRoomFunc != nil {
		return c.createRoomFunc(ctx, req)
	}

	return connect.NewResponse(&roomv1.CreateRoomResponse{}), nil
}

func (c fakeRoomClient) GetRoom(
	ctx context.Context,
	req *connect.Request[roomv1.GetRoomRequest],
) (*connect.Response[roomv1.GetRoomResponse], error) {
	if c.getRoomFunc != nil {
		return c.getRoomFunc(ctx, req)
	}

	return connect.NewResponse(&roomv1.GetRoomResponse{}), nil
}

func (c fakeRoomClient) GetRoomMembers(
	ctx context.Context,
	req *connect.Request[roomv1.GetRoomMembersRequest],
) (*connect.Response[roomv1.GetRoomMembersResponse], error) {
	if c.getRoomMembersFunc != nil {
		return c.getRoomMembersFunc(ctx, req)
	}

	return connect.NewResponse(&roomv1.GetRoomMembersResponse{}), nil
}

func (c fakeRoomClient) ListRooms(
	ctx context.Context,
	req *connect.Request[roomv1.ListRoomsRequest],
) (*connect.Response[roomv1.ListRoomsResponse], error) {
	if c.listRoomsFunc != nil {
		return c.listRoomsFunc(ctx, req)
	}

	return connect.NewResponse(&roomv1.ListRoomsResponse{}), nil
}

func (c fakeRoomClient) JoinRoom(
	ctx context.Context,
	req *connect.Request[roomv1.JoinRoomRequest],
) (*connect.Response[emptypb.Empty], error) {
	if c.joinRoomFunc != nil {
		return c.joinRoomFunc(ctx, req)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (c fakeRoomClient) LeaveRoom(
	ctx context.Context,
	req *connect.Request[roomv1.LeaveRoomRequest],
) (*connect.Response[emptypb.Empty], error) {
	if c.leaveRoomFunc != nil {
		return c.leaveRoomFunc(ctx, req)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

type fakeMessageClient struct {
	getMessagesFunc func(context.Context, *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error)
}

func (c fakeMessageClient) GetMessages(
	ctx context.Context,
	req *connect.Request[chatv1.GetMessagesRequest],
) (*connect.Response[chatv1.GetMessagesResponse], error) {
	if c.getMessagesFunc != nil {
		return c.getMessagesFunc(ctx, req)
	}

	return connect.NewResponse(&chatv1.GetMessagesResponse{}), nil
}

func TestRoomHandlerListRoomsForwardsAuthenticatedUserID(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{
		listRoomsFunc: func(ctx context.Context, req *connect.Request[roomv1.ListRoomsRequest]) (*connect.Response[roomv1.ListRoomsResponse], error) {
			if req.Msg.GetUserId() != "user-1" {
				t.Fatalf("expected user-1, got %q", req.Msg.GetUserId())
			}

			return connect.NewResponse(&roomv1.ListRoomsResponse{}), nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.ListRooms))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestRoomConnectHandlerListRoomsUsesAuthenticatedUserID(t *testing.T) {
	roomHandler, err := NewRoomConnectHandler(fakeRoomClient{
		listRoomsFunc: func(ctx context.Context, req *connect.Request[roomv1.ListRoomsRequest]) (*connect.Response[roomv1.ListRoomsResponse], error) {
			if req.Msg.GetUserId() != "user-1" {
				t.Fatalf("expected authenticated user-1, got %q", req.Msg.GetUserId())
			}

			return connect.NewResponse(&roomv1.ListRoomsResponse{}), nil
		},
	})
	if err != nil {
		t.Fatalf("create room connect handler: %v", err)
	}

	handler := requireAuth(t, roomv1connectHandler(t, roomHandler))

	req := httptest.NewRequest(
		http.MethodPost,
		"/pylon.room.v1.RoomService/ListRooms",
		strings.NewReader(`{"userId":"spoofed-user"}`),
	)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRoomConnectHandlerCreateRoomUsesAuthenticatedCreatorID(t *testing.T) {
	roomHandler, err := NewRoomConnectHandler(fakeRoomClient{
		createRoomFunc: func(ctx context.Context, req *connect.Request[roomv1.CreateRoomRequest]) (*connect.Response[roomv1.CreateRoomResponse], error) {
			if req.Msg.GetCreatorId() != "user-1" {
				t.Fatalf("expected authenticated creator user-1, got %q", req.Msg.GetCreatorId())
			}

			if req.Msg.GetName() != "general" {
				t.Fatalf("expected general, got %q", req.Msg.GetName())
			}

			return connect.NewResponse(&roomv1.CreateRoomResponse{
				Room: &roomv1.Room{
					Id:   "room-1",
					Name: "general",
				},
			}), nil
		},
	})
	if err != nil {
		t.Fatalf("create room connect handler: %v", err)
	}

	handler := requireAuth(t, roomv1connectHandler(t, roomHandler))

	req := httptest.NewRequest(
		http.MethodPost,
		"/pylon.room.v1.RoomService/CreateRoom",
		strings.NewReader(`{"name":"general","creatorId":"spoofed-user"}`),
	)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMessageHandlerListMessagesForwardsRoomAndUserID(t *testing.T) {
	messageHandler := NewMessageHandler(fakeMessageClient{
		getMessagesFunc: func(ctx context.Context, req *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error) {
			if req.Msg.GetRoomId() != "room-1" {
				t.Fatalf("expected room-1, got %q", req.Msg.GetRoomId())
			}

			if req.Msg.GetUserId() != "user-1" {
				t.Fatalf("expected user-1, got %q", req.Msg.GetUserId())
			}

			if req.Msg.GetLimit() != 25 {
				t.Fatalf("expected limit 25, got %d", req.Msg.GetLimit())
			}

			return connect.NewResponse(&chatv1.GetMessagesResponse{}), nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(messageHandler.ListMessages))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/room-1/messages?limit=25", nil)
	req.SetPathValue("id", "room-1")
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestRoomHandlerCreateRoomForwardsAuthenticatedUserIDAndBody(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{
		createRoomFunc: func(ctx context.Context, req *connect.Request[roomv1.CreateRoomRequest]) (*connect.Response[roomv1.CreateRoomResponse], error) {
			if req.Msg.GetCreatorId() != "user-1" {
				t.Fatalf("expected creator user-1, got %q", req.Msg.GetCreatorId())
			}

			if req.Msg.GetName() != "general" {
				t.Fatalf("expected room name general, got %q", req.Msg.GetName())
			}

			if req.Msg.GetType() != roomv1.RoomType_ROOM_TYPE_CHANNEL {
				t.Fatalf("expected channel room type, got %v", req.Msg.GetType())
			}

			if len(req.Msg.GetMemberIds()) != 2 {
				t.Fatalf("expected 2 normalized members, got %#v", req.Msg.GetMemberIds())
			}

			return connect.NewResponse(&roomv1.CreateRoomResponse{
				Room: &roomv1.Room{
					Id:   "room-1",
					Name: "general",
					Type: roomv1.RoomType_ROOM_TYPE_CHANNEL,
				},
			}), nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.CreateRoom))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(`{
		"name": " general ",
		"type": "channel",
		"member_ids": [" user-2 ", "user-2", "", "user-3"]
	}`))
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"id":"room-1"`) {
		t.Fatalf("expected room response, got %s", rec.Body.String())
	}
}

func TestRoomHandlerJoinRoomForwardsRoomAndUserID(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{
		joinRoomFunc: func(ctx context.Context, req *connect.Request[roomv1.JoinRoomRequest]) (*connect.Response[emptypb.Empty], error) {
			if req.Msg.GetRoomId() != "room-1" {
				t.Fatalf("expected room-1, got %q", req.Msg.GetRoomId())
			}

			if req.Msg.GetUserId() != "user-1" {
				t.Fatalf("expected user-1, got %q", req.Msg.GetUserId())
			}

			return connect.NewResponse(&emptypb.Empty{}), nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.JoinRoom))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms/room-1/join", nil)
	req.SetPathValue("id", "room-1")
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"joined":true`) {
		t.Fatalf("expected joined response, got %s", rec.Body.String())
	}
}

func TestRoomHandlerLeaveRoomForwardsRoomAndUserID(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{
		leaveRoomFunc: func(ctx context.Context, req *connect.Request[roomv1.LeaveRoomRequest]) (*connect.Response[emptypb.Empty], error) {
			if req.Msg.GetRoomId() != "room-1" {
				t.Fatalf("expected room-1, got %q", req.Msg.GetRoomId())
			}

			if req.Msg.GetUserId() != "user-1" {
				t.Fatalf("expected user-1, got %q", req.Msg.GetUserId())
			}

			return connect.NewResponse(&emptypb.Empty{}), nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.LeaveRoom))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms/room-1/leave", nil)
	req.SetPathValue("id", "room-1")
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"left":true`) {
		t.Fatalf("expected left response, got %s", rec.Body.String())
	}
}

func TestRoomHandlerCreateRoomMapsConnectError(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{
		createRoomFunc: func(ctx context.Context, req *connect.Request[roomv1.CreateRoomRequest]) (*connect.Response[roomv1.CreateRoomResponse], error) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("room name is required"))
		},
	})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.CreateRoom))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(`{"name": ""}`))
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "room name is required") {
		t.Fatalf("expected upstream validation message, got %s", rec.Body.String())
	}
}

func TestMessageHandlerListMessagesRejectsInvalidLimit(t *testing.T) {
	messageHandler := NewMessageHandler(fakeMessageClient{
		getMessagesFunc: func(ctx context.Context, req *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error) {
			t.Fatal("message client should not be called for invalid limit")
			return nil, nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(messageHandler.ListMessages))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/room-1/messages?limit=invalid", nil)
	req.SetPathValue("id", "room-1")
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMessageHandlerListMessagesMapsNotFound(t *testing.T) {
	messageHandler := NewMessageHandler(fakeMessageClient{
		getMessagesFunc: func(ctx context.Context, req *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("room not found"))
		},
	})

	handler := requireAuth(t, http.HandlerFunc(messageHandler.ListMessages))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/room-1/messages", nil)
	req.SetPathValue("id", "room-1")
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d body=%s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), `"code":"not_found"`) {
		t.Fatalf("expected not_found response, got %s", rec.Body.String())
	}
}

func TestRoomHandlerListRoomsRequiresUserID(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{
		listRoomsFunc: func(ctx context.Context, req *connect.Request[roomv1.ListRoomsRequest]) (*connect.Response[roomv1.ListRoomsResponse], error) {
			t.Fatal("room client should not be called without user id")
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms", nil)
	rec := httptest.NewRecorder()

	roomHandler.ListRooms(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRoomHandlerCreateRoomRejectsInvalidJSON(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{
		createRoomFunc: func(ctx context.Context, req *connect.Request[roomv1.CreateRoomRequest]) (*connect.Response[roomv1.CreateRoomResponse], error) {
			t.Fatal("room client should not be called for invalid json")
			return nil, nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.CreateRoom))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(`{invalid-json`))
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRoomHandlerJoinRoomRequiresRoomID(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.JoinRoom))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms//join", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRoomHandlerLeaveRoomRequiresRoomID(t *testing.T) {
	roomHandler := NewRoomHandler(fakeRoomClient{})

	handler := requireAuth(t, http.HandlerFunc(roomHandler.LeaveRoom))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms//leave", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMessageHandlerListMessagesRequiresRoomID(t *testing.T) {
	messageHandler := NewMessageHandler(fakeMessageClient{
		getMessagesFunc: func(ctx context.Context, req *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error) {
			t.Fatal("message client should not be called without room id")
			return nil, nil
		},
	})

	handler := requireAuth(t, http.HandlerFunc(messageHandler.ListMessages))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms//messages", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, "user-1"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestWriteConnectErrorMapsAdditionalCodes(t *testing.T) {
	tests := []struct {
		name string
		code connect.Code
		want int
	}{
		{name: "unauthenticated", code: connect.CodeUnauthenticated, want: http.StatusUnauthorized},
		{name: "permission denied", code: connect.CodePermissionDenied, want: http.StatusForbidden},
		{name: "already exists", code: connect.CodeAlreadyExists, want: http.StatusConflict},
		{name: "failed precondition", code: connect.CodeFailedPrecondition, want: http.StatusPreconditionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			writeConnectError(rec, connect.NewError(tt.code, errors.New("upstream message")))

			if rec.Code != tt.want {
				t.Fatalf("expected status %d, got %d body=%s", tt.want, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRoomTypeToProtoSupportsAliasesAndDefault(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want roomv1.RoomType
	}{
		{name: "direct", in: "direct", want: roomv1.RoomType_ROOM_TYPE_DIRECT},
		{name: "dm", in: "dm", want: roomv1.RoomType_ROOM_TYPE_DIRECT},
		{name: "group", in: " group ", want: roomv1.RoomType_ROOM_TYPE_GROUP},
		{name: "unknown", in: "unknown", want: roomv1.RoomType_ROOM_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roomTypeToProto(tt.in); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func requireAuth(t *testing.T, next http.Handler) http.Handler {
	t.Helper()

	middleware, err := gatewaymiddleware.NewAuthMiddleware(testJWTSecret)
	if err != nil {
		t.Fatalf("create auth middleware: %v", err)
	}

	return middleware.RequireAuth(next)
}

func roomv1connectHandler(t *testing.T, roomHandler *RoomConnectHandler) http.Handler {
	t.Helper()

	path, handler := roomv1connect.NewRoomServiceHandler(roomHandler)
	if path != "/pylon.room.v1.RoomService/" {
		t.Fatalf("unexpected room connect path %q", path)
	}

	return handler
}

func testJWT(t *testing.T, userID string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      userID,
		"username": "alice",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return tokenString
}

func TestWriteConnectErrorDoesNotLeakDefaultConnectErrorMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	writeConnectError(rec, connect.NewError(connect.CodeInternal, errors.New("sql: password=secret")))

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}

	body := rec.Body.String()
	if strings.Contains(body, "password=secret") {
		t.Fatalf("expected internal message to be hidden, got body %q", body)
	}

	if !strings.Contains(body, "internal service error") {
		t.Fatalf("expected generic error message, got body %q", body)
	}
}

func TestWriteConnectErrorDoesNotLeakNonConnectErrorMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	writeConnectError(rec, errors.New("dial tcp internal-db:5432 failed"))

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}

	body := rec.Body.String()
	if strings.Contains(body, "internal-db") {
		t.Fatalf("expected non-connect error details to be hidden, got body %q", body)
	}

	if !strings.Contains(body, "internal service error") {
		t.Fatalf("expected generic error message, got body %q", body)
	}
}

func TestWriteConnectErrorUsesGenericUnavailableMessage(t *testing.T) {
	rec := httptest.NewRecorder()

	writeConnectError(rec, connect.NewError(connect.CodeUnavailable, errors.New("chat-service pod 10.0.0.9 refused connection")))

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", rec.Code)
	}

	body := rec.Body.String()
	if strings.Contains(body, "10.0.0.9") {
		t.Fatalf("expected unavailable details to be hidden, got body %q", body)
	}

	if !strings.Contains(body, "internal service unavailable") {
		t.Fatalf("expected generic unavailable message, got body %q", body)
	}
}

func TestRoomTypeToProtoSupportsChannel(t *testing.T) {
	got := roomTypeToProto("channel")
	if got != roomv1.RoomType_ROOM_TYPE_CHANNEL {
		t.Fatalf("expected channel type, got %v", got)
	}
}

func TestDecodeJSONAcceptsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(`{
		"name": "general",
		"type": "channel",
		"member_ids": ["user-1"],
		"future_field": "ignored"
	}`))
	rec := httptest.NewRecorder()

	var body createRoomRequest
	if err := decodeJSON(rec, req, &body); err != nil {
		t.Fatalf("decode json: %v", err)
	}

	if body.Name != "general" {
		t.Fatalf("expected room name general, got %q", body.Name)
	}
}

func TestDecodeJSONRejectsBodyLargerThanLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", strings.NewReader(`{"name":"`+strings.Repeat("a", maxJSONBodyBytes)+`"}`))
	rec := httptest.NewRecorder()

	var body createRoomRequest
	err := decodeJSON(rec, req, &body)
	if err == nil {
		t.Fatal("expected max body size error, got nil")
	}

	if !strings.Contains(err.Error(), "http: request body too large") {
		t.Fatalf("expected request body too large error, got %v", err)
	}
}
