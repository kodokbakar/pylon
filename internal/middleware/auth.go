package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kodokbakar/pylon/internal/response"
)

type contextKey string

const tokenContextKey contextKey = "auth_token"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if header == "" {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "authorization header is required")
			return
		}

		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "bearer token is required")
			return
		}

		ctx := context.WithValue(r.Context(), tokenContextKey, strings.TrimSpace(token))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenContextKey).(string)
	return token, ok
}
