package handler

import (
	"net/http"

	"github.com/kodokbakar/pylon/internal/response"
)

type RoomHandler struct{}

func NewRoomHandler() *RoomHandler {
	return &RoomHandler{}
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "list rooms endpoint is not implemented yet")
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "create room endpoint is not implemented yet")
}

func (h *RoomHandler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	if roomID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "room id is required")
		return
	}

	response.Error(w, http.StatusNotImplemented, "not_implemented", "join room endpoint is not implemented yet")
}

func (h *RoomHandler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")
	if roomID == "" {
		response.Error(w, http.StatusBadRequest, "bad_request", "room id is required")
		return
	}

	response.Error(w, http.StatusNotImplemented, "not_implemented", "leave room endpoint is not implemented yet")
}
