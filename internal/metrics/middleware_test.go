package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPPathLabel(t *testing.T) {
	tests := []struct {
		name string
		req  *http.Request
		want string
	}{
		{
			name: "nil request",
			req:  nil,
			want: "unknown",
		},
		{
			name: "uses request pattern",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/room-1/messages", nil)
				req.Pattern = "GET /api/v1/rooms/{id}/messages"
				return req
			}(),
			want: "GET /api/v1/rooms/{id}/messages",
		},
		{
			name: "uses url path when pattern is empty",
			req:  httptest.NewRequest(http.MethodGet, "/health", nil),
			want: "/health",
		},
		{
			name: "empty url path becomes slash",
			req: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
				req.URL.Path = ""
				return req
			}(),
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := httpPathLabel(tt.req)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestGRPCMethodLabel(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "empty path",
			path: "",
			want: "unknown",
		},
		{
			name: "single segment",
			path: "/SendMessage",
			want: "SendMessage",
		},
		{
			name: "multi segment",
			path: "/pylon.chat.v1.ChatService/SendMessage",
			want: "SendMessage",
		},
		{
			name: "trims whitespace and slashes",
			path: "  /pylon.room.v1.RoomService/CreateRoom/  ",
			want: "CreateRoom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grpcMethodLabel(tt.path)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
