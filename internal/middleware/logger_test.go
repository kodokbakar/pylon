package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerPassesRequestAndRecordsStatus(t *testing.T) {
	called := false

	handler := Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}

func TestStatusRecorderWriteHeaderStoresStatus(t *testing.T) {
	rec := &statusRecorder{
		ResponseWriter: httptest.NewRecorder(),
		status:         http.StatusOK,
	}

	rec.WriteHeader(http.StatusNoContent)

	if rec.status != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.status)
	}
}

func TestStatusRecorderHijackReturnsErrorWhenUnsupported(t *testing.T) {
	rec := &statusRecorder{
		ResponseWriter: httptest.NewRecorder(),
		status:         http.StatusOK,
	}

	conn, rw, err := rec.Hijack()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if conn != nil {
		t.Fatalf("expected nil conn, got %#v", conn)
	}

	if rw != nil {
		t.Fatalf("expected nil read writer, got %#v", rw)
	}
}

func TestStatusRecorderFlushAndUnwrap(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := &statusRecorder{
		ResponseWriter: underlying,
		status:         http.StatusOK,
	}

	rec.Flush()

	if rec.Unwrap() != underlying {
		t.Fatal("expected unwrap to return underlying response writer")
	}
}
