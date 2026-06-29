package middleware

import (
	"net/http"
	"strings"
)

const (
	corsAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowedHeaders = "Content-Type, Authorization, Connect-Protocol-Version, Connect-Timeout-Ms, X-Requested-With"
)

func CORS(next http.Handler) http.Handler {
	return CORSWithOrigins(next, []string{"*"})
}

func CORSWithOrigins(next http.Handler, allowedOrigins []string) http.Handler {
	origins := normalizeOrigins(allowedOrigins)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))

		allowedOrigin, allowed := matchOrigin(origin, origins)
		if origin != "" && !allowed {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		headers := w.Header()
		if allowedOrigin != "" {
			headers.Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		headers.Set("Access-Control-Allow-Methods", corsAllowedMethods)
		headers.Set("Access-Control-Allow-Headers", corsAllowedHeaders)
		headers.Set("Access-Control-Max-Age", "86400")

		if allowedOrigin != "*" {
			headers.Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func normalizeOrigins(origins []string) []string {
	normalized := make([]string, 0, len(origins))

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			normalized = append(normalized, origin)
		}
	}

	if len(normalized) == 0 {
		return []string{"*"}
	}

	return normalized
}

func matchOrigin(origin string, allowedOrigins []string) (string, bool) {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return "*", true
		}

		if origin != "" && origin == allowedOrigin {
			return allowedOrigin, true
		}
	}

	if origin == "" {
		return "", true
	}

	return "", false
}
