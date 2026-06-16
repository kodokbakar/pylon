package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool         `json:"success"`
	Data    any          `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, Response{
		Success: true,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, Response{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
