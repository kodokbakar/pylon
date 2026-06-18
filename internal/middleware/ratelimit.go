package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kodokbakar/pylon/internal/response"
)

type rateLimitClient struct {
	count    int
	resetAt  time.Time
	lastSeen time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	rps     int
	clients map[string]*rateLimitClient
}

func RateLimit(next http.Handler, rps int) http.Handler {
	limiter := &rateLimiter{
		rps:     rps,
		clients: make(map[string]*rateLimitClient),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rps <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		if !limiter.allow(clientIP(r)) {
			response.Error(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *rateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	client, ok := l.clients[ip]
	if !ok || now.After(client.resetAt) {
		l.clients[ip] = &rateLimitClient{
			count:    1,
			resetAt:  now.Add(time.Second),
			lastSeen: now,
		}
		l.cleanup(now)
		return true
	}

	client.lastSeen = now
	if client.count >= l.rps {
		return false
	}

	client.count++
	return true
}

func (l *rateLimiter) cleanup(now time.Time) {
	for ip, client := range l.clients {
		if now.Sub(client.lastSeen) > time.Minute {
			delete(l.clients, ip)
		}
	}
}

func clientIP(r *http.Request) string {
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		firstIP := strings.TrimSpace(parts[0])
		if firstIP != "" {
			return firstIP
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	if r.RemoteAddr == "" {
		return "unknown"
	}

	return r.RemoteAddr
}
