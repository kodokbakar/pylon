package middleware

import (
	"log"
	"net/http"

	"github.com/kodokbakar/pylon/internal/response"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic recovered: %v", recovered)
				response.Error(w, http.StatusInternalServerError, "internal_server_error", "internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
