package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/kodokbakar/pylon/internal/response"
)

type contextKey string

const (
	userIDContextKey   contextKey = "user_id"
	usernameContextKey contextKey = "username"
	emailContextKey    contextKey = "email"
)

type AuthMiddleware struct {
	secret []byte
}

func NewAuthMiddleware(secret string) (*AuthMiddleware, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}

	return &AuthMiddleware{
		secret: []byte(secret),
	}, nil
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := bearerToken(r)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "bearer token is required")
			return
		}

		userID, username, email, err := m.validateToken(tokenString)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		if username != "" {
			ctx = context.WithValue(ctx, usernameContextKey, username)
		}
		if email != "" {
			ctx = context.WithValue(ctx, emailContextKey, email)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailContextKey).(string)
	return email, ok
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameContextKey).(string)
	return username, ok
}

func bearerToken(r *http.Request) (string, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header != "" {
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok {
			return "", false
		}

		token = strings.TrimSpace(token)
		if token == "" {
			return "", false
		}

		return token, true
	}

	for _, key := range []string{"access_token", "token"} {
		token := strings.TrimSpace(r.URL.Query().Get(key))
		if token != "" {
			return token, true
		}
	}

	return "", false
}

func (m *AuthMiddleware) validateToken(tokenString string) (string, string, string, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}

		return m.secret, nil
	})
	if err != nil {
		return "", "", "", fmt.Errorf("parse jwt token: %w", err)
	}

	if !token.Valid {
		return "", "", "", fmt.Errorf("jwt token is invalid")
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return "", "", "", fmt.Errorf("get jwt subject: %w", err)
	}

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "", "", "", fmt.Errorf("jwt subject is required")
	}

	tokenType := stringClaim(claims, "token_type")
	if tokenType != "" && tokenType != "access" {
		return "", "", "", fmt.Errorf("jwt token is not an access token")
	}

	return subject, stringClaim(claims, "username", "preferred_username", "name"), stringClaim(claims, "email"), nil
}

func stringClaim(claims jwt.MapClaims, keys ...string) string {
	for _, key := range keys {
		value, ok := claims[key].(string)
		if !ok {
			continue
		}

		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}
