package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	chatv1 "github.com/kodokbakar/pylon/gen/pylon/chat/v1"
	"github.com/kodokbakar/pylon/internal/response"
	gatewaymiddleware "github.com/kodokbakar/pylon/services/api-gateway/middleware"
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

type MessageServiceClient interface {
	GetMessages(context.Context, *connect.Request[chatv1.GetMessagesRequest]) (*connect.Response[chatv1.GetMessagesResponse], error)
}

type MessageHandler struct {
	client MessageServiceClient
}

func NewMessageHandler(client MessageServiceClient) *MessageHandler {
	return &MessageHandler{client: client}
}

func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
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

	limit, err := parseLimit(r.URL.Query().Get("limit"), defaultMessageLimit, maxMessageLimit)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}

	res, err := h.client.GetMessages(r.Context(), connect.NewRequest(&chatv1.GetMessagesRequest{
		RoomId:   roomID,
		UserId:   userID,
		Limit:    int32(limit),
		BeforeId: strings.TrimSpace(r.URL.Query().Get("before_id")),
	}))
	if err != nil {
		writeConnectError(w, err)
		return
	}

	response.Success(w, http.StatusOK, res.Msg)
}

func parseLimit(value string, fallback, max int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	if limit <= 0 {
		return fallback, nil
	}

	if limit > max {
		return max, nil
	}

	return limit, nil
}
