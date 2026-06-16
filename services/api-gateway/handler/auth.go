package handler

import (
	"net/http"

	"github.com/kodokbakar/pylon/internal/response"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "register endpoint is not implemented yet")
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "login endpoint is not implemented yet")
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "refresh endpoint is not implemented yet")
}
