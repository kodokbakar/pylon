package handler

import (
	"net/http"

	"github.com/kodokbakar/pylon/internal/response"
)

type MessageHandler struct{}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	if roomID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "room id is required")
		return
	}

	response.Error(w, http.StatusNotImplemented, "not_implemented", "message history endpoint is not implemented yet")
}
