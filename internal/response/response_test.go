package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessWritesJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	Success(rec, http.StatusOK, map[string]string{"id": "123"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected content type application/json, got %q", rec.Header().Get("Content-Type"))
	}

	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if !body.Success {
		t.Fatal("expected success true")
	}

	if body.Error != nil {
		t.Fatalf("expected nil error, got %#v", body.Error)
	}
}

func TestErrorWritesJSONResponse(t *testing.T) {
	rec := httptest.NewRecorder()

	Error(rec, http.StatusNotFound, "not_found", "room not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var body Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body.Success {
		t.Fatal("expected success false")
	}

	if body.Error == nil {
		t.Fatal("expected error detail")
	}

	if body.Error.Code != "not_found" {
		t.Fatalf("expected error code not_found, got %q", body.Error.Code)
	}
}
