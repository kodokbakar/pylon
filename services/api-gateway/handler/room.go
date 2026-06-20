package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	roomv1 "github.com/kodokbakar/pylon/gen/pylon/room/v1"
	"github.com/kodokbakar/pylon/internal/response"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxJSONBodyBytes = 1 << 20 // 1 MiB

type RoomServiceClient interface {
	CreateRoom(context.Context, *connect.Request[roomv1.CreateRoomRequest]) (*connect.Response[roomv1.CreateRoomResponse], error)
	ListRooms(context.Context, *connect.Request[roomv1.ListRoomsRequest]) (*connect.Response[roomv1.ListRoomsResponse], error)
	JoinRoom(context.Context, *connect.Request[roomv1.JoinRoomRequest]) (*connect.Response[emptypb.Empty], error)
	LeaveRoom(context.Context, *connect.Request[roomv1.LeaveRoomRequest]) (*connect.Response[emptypb.Empty], error)
}

type RoomHandler struct {
	client RoomServiceClient
}

type createRoomRequest struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	MemberIDs []string `json:"member_ids"`
}

func NewRoomHandler(client RoomServiceClient) *RoomHandler {
	return &RoomHandler{client: client}
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	userID, ok := gatewaymiddleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "user id is required")
		return
	}

	res, err := h.client.ListRooms(r.Context(), connect.NewRequest(&roomv1.ListRoomsRequest{
		UserId: userID,
	}))
	if err != nil {
		writeConnectError(w, err)
		return
	}

	response.Success(w, http.StatusOK, res.Msg)
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := gatewaymiddleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "user id is required")
		return
	}

	var body createRoomRequest
	if err := decodeJSON(w, r, &body); err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	res, err := h.client.CreateRoom(r.Context(), connect.NewRequest(&roomv1.CreateRoomRequest{
		Name:      strings.TrimSpace(body.Name),
		Type:      roomTypeToProto(body.Type),
		CreatorId: userID,
		MemberIds: normalizeStringSlice(body.MemberIDs),
	}))
	if err != nil {
		writeConnectError(w, err)
		return
	}

	response.Success(w, http.StatusCreated, res.Msg)
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimSpace(r.PathValue("id"))
	if roomID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "room id is required")
		return
	}

	userID, ok := gatewaymiddleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "user id is required")
		return
	}

	_, err := h.client.JoinRoom(r.Context(), connect.NewRequest(&roomv1.JoinRoomRequest{
		RoomId: roomID,
		UserId: userID,
	}))
	if err != nil {
		writeConnectError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"joined": true,
	})
}

func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimSpace(r.PathValue("id"))
	if roomID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "room id is required")
		return
	}

	userID, ok := gatewaymiddleware.UserIDFromContext(r.Context())
	if !ok || strings.TrimSpace(userID) == "" {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "user id is required")
		return
	}

	_, err := h.client.LeaveRoom(r.Context(), connect.NewRequest(&roomv1.LeaveRoomRequest{
		RoomId: roomID,
		UserId: userID,
	}))
	if err != nil {
		writeConnectError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"left": true,
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) (err error) {
	if r.Body == nil {
		return fmt.Errorf("request body is required")
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)

	defer func() {
		closeErr := r.Body.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close request body: %w", closeErr)
		}
	}()

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid json body: %w", err)
	}

	if decoder.More() {
		return fmt.Errorf("invalid json body: multiple json values are not allowed")
	}

	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err == nil && len(bytes.TrimSpace(trailing)) > 0 {
		return fmt.Errorf("invalid json body: multiple json values are not allowed")
	}

	return nil
}

func normalizeStringSlice(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	return normalized
}

func roomTypeToProto(roomType string) roomv1.RoomType {
	switch strings.TrimSpace(strings.ToLower(roomType)) {
	case "direct", "dm":
		return roomv1.RoomType_ROOM_TYPE_DIRECT
	case "group":
		return roomv1.RoomType_ROOM_TYPE_GROUP
	case "channel":
		return roomv1.RoomType_ROOM_TYPE_CHANNEL
	default:
		return roomv1.RoomType_ROOM_TYPE_UNSPECIFIED
	}
}
