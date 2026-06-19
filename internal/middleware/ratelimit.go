package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kodokbakar/pylon/internal/response"
)

const rateLimitScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local max = tonumber(ARGV[3])
local member = ARGV[4]
local cutoff = now - window

redis.call("ZREMRANGEBYSCORE", key, "-inf", cutoff)

local count = redis.call("ZCARD", key)
if count >= max then
	local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
	local reset = now + window
	if oldest[2] ~= nil then
		reset = tonumber(oldest[2]) + window
	end

	redis.call("PEXPIRE", key, window)

	local remaining = max - count
	if remaining < 0 then
		remaining = 0
	end

	return {0, remaining, reset}
end

redis.call("ZADD", key, now, member)
count = count + 1
redis.call("PEXPIRE", key, window)

local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
local reset = now + window
if oldest[2] ~= nil then
	reset = tonumber(oldest[2]) + window
end

return {1, max - count, reset}
`

var (
	roomJoinPathRegex     = regexp.MustCompile(`^/api/v1/rooms/[^/]+/join$`)
	roomMessagesPathRegex = regexp.MustCompile(`^/api/v1/rooms/[^/]+/messages$`)
)

type RateLimiter struct {
	client   *redis.Client
	limits   map[string]RateLimit
	store    rateLimitStore
	now      func() time.Time
	memberID func() (string, error)
}

type RateLimit struct {
	Max    int
	Window time.Duration
}

type rateLimitStore interface {
	Check(ctx context.Context, key, member string, now time.Time, limit RateLimit) (rateLimitDecision, error)
}

type rateLimitDecision struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

type redisRateLimitStore struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		client:   client,
		limits:   defaultRateLimits(),
		store:    redisRateLimitStore{client: client},
		now:      time.Now,
		memberID: newRateLimitMember,
	}
}

func defaultRateLimits() map[string]RateLimit {
	return map[string]RateLimit{
		"POST /api/v1/auth/register":      {Max: 5, Window: time.Minute},
		"POST /api/v1/auth/login":         {Max: 10, Window: time.Minute},
		"POST /api/v1/auth/refresh":       {Max: 10, Window: time.Minute},
		"POST /api/v1/rooms":              {Max: 10, Window: time.Minute},
		"POST /api/v1/rooms/{id}/join":    {Max: 20, Window: time.Minute},
		"GET /api/v1/rooms/{id}/messages": {Max: 60, Window: time.Minute},
	}
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl == nil || rl.store == nil {
			next.ServeHTTP(w, r)
			return
		}

		routeKey, limit, ok := rl.matchLimit(r)
		if !ok || limit.Max <= 0 || limit.Window <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		now := rl.currentTime()
		member, err := rl.newMember()
		if err != nil {
			log.Printf("rate limiter member generation failed: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		redisKey := rateLimitRedisKey(routeKey, clientIP(r))
		decision, err := rl.store.Check(r.Context(), redisKey, member, now, limit)
		if err != nil {
			log.Printf("rate limiter check failed: route=%q error=%v", routeKey, err)
			next.ServeHTTP(w, r)
			return
		}

		if decision.ResetAt.IsZero() {
			decision.ResetAt = now.Add(limit.Window)
		}

		setRateLimitHeaders(w, limit, decision, now)

		if !decision.Allowed {
			response.Error(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) matchLimit(r *http.Request) (string, RateLimit, bool) {
	path := strings.TrimSpace(r.URL.Path)
	if path == "" {
		path = "/"
	}

	exactKey := r.Method + " " + path
	if limit, ok := rl.limits[exactKey]; ok {
		return exactKey, limit, true
	}

	switch {
	case r.Method == http.MethodPost && roomJoinPathRegex.MatchString(path):
		key := "POST /api/v1/rooms/{id}/join"
		return key, rl.limits[key], true
	case r.Method == http.MethodGet && roomMessagesPathRegex.MatchString(path):
		key := "GET /api/v1/rooms/{id}/messages"
		return key, rl.limits[key], true
	default:
		return "", RateLimit{}, false
	}
}

func (rl *RateLimiter) currentTime() time.Time {
	if rl.now != nil {
		return rl.now()
	}

	return time.Now()
}

func (rl *RateLimiter) newMember() (string, error) {
	if rl.memberID != nil {
		return rl.memberID()
	}

	return newRateLimitMember()
}

func (s redisRateLimitStore) Check(ctx context.Context, key, member string, now time.Time, limit RateLimit) (rateLimitDecision, error) {
	if s.client == nil {
		return rateLimitDecision{
			Allowed:   true,
			Remaining: limit.Max,
			ResetAt:   now.Add(limit.Window),
		}, nil
	}

	result, err := s.client.Eval(
		ctx,
		rateLimitScript,
		[]string{key},
		now.UnixMilli(),
		limit.Window.Milliseconds(),
		limit.Max,
		member,
	).Result()
	if err != nil {
		return rateLimitDecision{}, fmt.Errorf("run redis rate limit script: %w", err)
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 3 {
		return rateLimitDecision{}, fmt.Errorf("unexpected redis rate limit result: %T", result)
	}

	allowed, err := redisInt(values[0])
	if err != nil {
		return rateLimitDecision{}, fmt.Errorf("parse redis allowed result: %w", err)
	}

	remaining, err := redisInt(values[1])
	if err != nil {
		return rateLimitDecision{}, fmt.Errorf("parse redis remaining result: %w", err)
	}

	resetMs, err := redisInt(values[2])
	if err != nil {
		return rateLimitDecision{}, fmt.Errorf("parse redis reset result: %w", err)
	}

	if remaining < 0 {
		remaining = 0
	}
	if remaining > int64(limit.Max) {
		remaining = int64(limit.Max)
	}

	return rateLimitDecision{
		Allowed:   allowed == 1,
		Remaining: int(remaining),
		ResetAt:   time.UnixMilli(resetMs),
	}, nil
}

func redisInt(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse string integer: %w", err)
		}
		return parsed, nil
	case []byte:
		parsed, err := strconv.ParseInt(string(v), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("parse byte integer: %w", err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported integer type %T", value)
	}
}

func setRateLimitHeaders(w http.ResponseWriter, limit RateLimit, decision rateLimitDecision, now time.Time) {
	remaining := decision.Remaining
	if remaining < 0 {
		remaining = 0
	}

	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit.Max))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(decision.ResetAt.Unix(), 10))

	if !decision.Allowed {
		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds(now, decision.ResetAt)))
	}
}

func retryAfterSeconds(now, resetAt time.Time) int {
	duration := resetAt.Sub(now)
	if duration <= 0 {
		return 1
	}

	return int(math.Ceil(duration.Seconds()))
}

func rateLimitRedisKey(routeKey, ip string) string {
	cleanedRoute := strings.NewReplacer(
		"{", "",
		"}", "",
	).Replace(routeKey)

	parts := strings.FieldsFunc(cleanedRoute, func(r rune) bool {
		return r == ' ' || r == '/'
	})

	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			segments = append(segments, part)
		}
	}

	route := strings.Join(segments, ":")
	if route == "" {
		route = "unknown"
	}

	ip = strings.TrimSpace(ip)
	if ip == "" {
		ip = "unknown"
	}

	return fmt.Sprintf("ratelimit:%s:%s", route, ip)
}

func newRateLimitMember() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return hex.EncodeToString(bytes[:]), nil
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
