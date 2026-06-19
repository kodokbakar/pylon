package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type fakeRateLimitStore struct {
	decisions []rateLimitDecision
	err       error
	calls     []fakeRateLimitCall
}

type fakeRateLimitCall struct {
	key    string
	member string
	now    time.Time
	limit  RateLimit
}

func (s *fakeRateLimitStore) Check(_ context.Context, key, member string, now time.Time, limit RateLimit) (rateLimitDecision, error) {
	s.calls = append(s.calls, fakeRateLimitCall{
		key:    key,
		member: member,
		now:    now,
		limit:  limit,
	})

	if s.err != nil {
		return rateLimitDecision{}, s.err
	}

	if len(s.decisions) == 0 {
		return rateLimitDecision{
			Allowed:   true,
			Remaining: limit.Max - 1,
			ResetAt:   now.Add(limit.Window),
		}, nil
	}

	decision := s.decisions[0]
	s.decisions = s.decisions[1:]

	return decision, nil
}

type windowRateLimitStore struct {
	hits map[string][]time.Time
}

func newWindowRateLimitStore() *windowRateLimitStore {
	return &windowRateLimitStore{
		hits: make(map[string][]time.Time),
	}
}

func (s *windowRateLimitStore) Check(_ context.Context, key, _ string, now time.Time, limit RateLimit) (rateLimitDecision, error) {
	cutoff := now.Add(-limit.Window)
	active := make([]time.Time, 0, len(s.hits[key]))

	for _, hit := range s.hits[key] {
		if hit.After(cutoff) {
			active = append(active, hit)
		}
	}

	if len(active) >= limit.Max {
		resetAt := active[0].Add(limit.Window)
		s.hits[key] = active

		return rateLimitDecision{
			Allowed:   false,
			Remaining: 0,
			ResetAt:   resetAt,
		}, nil
	}

	active = append(active, now)
	s.hits[key] = active

	resetAt := now.Add(limit.Window)
	if len(active) > 0 {
		resetAt = active[0].Add(limit.Window)
	}

	return rateLimitDecision{
		Allowed:   true,
		Remaining: limit.Max - len(active),
		ResetAt:   resetAt,
	}, nil
}

func TestRateLimiterAllowsRequestsWithinLimit(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	resetAt := now.Add(time.Minute)

	store := &fakeRateLimitStore{
		decisions: []rateLimitDecision{
			{
				Allowed:   true,
				Remaining: 9,
				ResetAt:   resetAt,
			},
		},
	}

	nextCalled := false
	limiter := newTestRateLimiter(store, now)
	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	assertRateLimitHeader(t, rec, "X-RateLimit-Limit", "10")
	assertRateLimitHeader(t, rec, "X-RateLimit-Remaining", "9")
	assertRateLimitHeader(t, rec, "X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

	if len(store.calls) != 1 {
		t.Fatalf("expected 1 rate limit call, got %d", len(store.calls))
	}

	if store.calls[0].key != "ratelimit:POST:api:v1:auth:login:203.0.113.10" {
		t.Fatalf("unexpected redis key %q", store.calls[0].key)
	}
}

func TestRateLimiterRejectsRequestsOverLimit(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	resetAt := now.Add(30 * time.Second)

	store := &fakeRateLimitStore{
		decisions: []rateLimitDecision{
			{
				Allowed:   false,
				Remaining: 0,
				ResetAt:   resetAt,
			},
		},
	}

	limiter := newTestRateLimiter(store, now)
	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec.Code)
	}

	assertRateLimitHeader(t, rec, "X-RateLimit-Limit", "10")
	assertRateLimitHeader(t, rec, "X-RateLimit-Remaining", "0")
	assertRateLimitHeader(t, rec, "X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))
	assertRateLimitHeader(t, rec, "Retry-After", "30")

	if !strings.Contains(rec.Body.String(), `"code":"rate_limited"`) {
		t.Fatalf("expected rate_limited error body, got %q", rec.Body.String())
	}
}

func TestRateLimiterBypassesUnconfiguredEndpoint(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	store := &fakeRateLimitStore{}

	nextCalled := false
	limiter := newTestRateLimiter(store, now)
	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if len(store.calls) != 0 {
		t.Fatalf("expected no rate limit calls, got %d", len(store.calls))
	}
}

func TestRateLimiterUsesFirstForwardedForValueInKey(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	store := &fakeRateLimitStore{}

	limiter := newTestRateLimiter(store, now)
	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.10, 203.0.113.20")
	req.RemoteAddr = "127.0.0.1:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(store.calls) != 1 {
		t.Fatalf("expected 1 rate limit call, got %d", len(store.calls))
	}

	if !strings.HasSuffix(store.calls[0].key, ":198.51.100.10") {
		t.Fatalf("expected forwarded ip in redis key, got %q", store.calls[0].key)
	}
}

func TestRateLimiterMatchesDynamicRoomJoinRoute(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	store := &fakeRateLimitStore{}

	limiter := newTestRateLimiter(store, now)
	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms/room-1/join", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(store.calls) != 1 {
		t.Fatalf("expected 1 rate limit call, got %d", len(store.calls))
	}

	if store.calls[0].limit.Max != 20 {
		t.Fatalf("expected room join max 20, got %d", store.calls[0].limit.Max)
	}

	if !strings.Contains(store.calls[0].key, "POST:api:v1:rooms:id:join") {
		t.Fatalf("expected room join route key, got %q", store.calls[0].key)
	}
}

func TestRateLimiterMatchesDynamicRoomMessagesRoute(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	store := &fakeRateLimitStore{}

	limiter := newTestRateLimiter(store, now)
	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/room-1/messages", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if len(store.calls) != 1 {
		t.Fatalf("expected 1 rate limit call, got %d", len(store.calls))
	}

	if store.calls[0].limit.Max != 60 {
		t.Fatalf("expected room messages max 60, got %d", store.calls[0].limit.Max)
	}
}

func TestRateLimiterAllowsRequestAfterWindowReset(t *testing.T) {
	currentTime := time.Unix(1_700_000_000, 0)

	limiter := &RateLimiter{
		limits: defaultRateLimits(),
		store:  newWindowRateLimitStore(),
		now: func() time.Time {
			return currentTime
		},
		memberID: func() (string, error) {
			return strconv.FormatInt(currentTime.UnixNano(), 10), nil
		},
	}
	limiter.limits["POST /api/v1/auth/login"] = RateLimit{
		Max:    1,
		Window: time.Minute,
	}

	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	firstReq.RemoteAddr = "203.0.113.10:1234"
	firstRec := httptest.NewRecorder()
	handler.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first status 200, got %d", firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	secondReq.RemoteAddr = "203.0.113.10:1234"
	secondRec := httptest.NewRecorder()
	handler.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second status 429, got %d", secondRec.Code)
	}

	currentTime = currentTime.Add(time.Minute + time.Second)

	thirdReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	thirdReq.RemoteAddr = "203.0.113.10:1234"
	thirdRec := httptest.NewRecorder()
	handler.ServeHTTP(thirdRec, thirdReq)

	if thirdRec.Code != http.StatusOK {
		t.Fatalf("expected third status 200 after reset, got %d", thirdRec.Code)
	}
}

func TestRateLimiterHandlesStoreErrorsGracefully(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	store := &fakeRateLimitStore{
		err: errors.New("redis unavailable"),
	}

	nextCalled := false
	limiter := newTestRateLimiter(store, now)
	handler := limiter.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
}

func newTestRateLimiter(store *fakeRateLimitStore, now time.Time) *RateLimiter {
	return &RateLimiter{
		limits: defaultRateLimits(),
		store:  store,
		now: func() time.Time {
			return now
		},
		memberID: func() (string, error) {
			return "member-1", nil
		},
	}
}

func assertRateLimitHeader(t *testing.T, rec *httptest.ResponseRecorder, key, want string) {
	t.Helper()

	if got := rec.Header().Get(key); got != want {
		t.Fatalf("expected %s %q, got %q", key, want, got)
	}
}
