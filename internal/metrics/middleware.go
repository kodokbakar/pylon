package metrics

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}

	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	return r.ResponseWriter.Write(body)
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not implement http.Hijacker")
	}

	return hijacker.Hijack()
}

func (r *statusRecorder) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}

	flusher.Flush()
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		rec := newStatusRecorder(w)

		IncHTTPRequestsInFlight()
		defer DecHTTPRequestsInFlight()

		next.ServeHTTP(rec, r)

		ObserveHTTPRequest(
			r.Method,
			httpPathLabel(r),
			rec.status,
			time.Since(startedAt),
		)
	})
}

func GRPCMiddleware(serviceName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		rec := newStatusRecorder(w)

		next.ServeHTTP(rec, r)

		ObserveGRPCRequest(
			serviceName,
			grpcMethodLabel(r.URL.Path),
			rec.status,
			time.Since(startedAt),
		)
	})
}

func httpPathLabel(r *http.Request) string {
	if r == nil {
		return "unknown"
	}

	if pattern := strings.TrimSpace(r.Pattern); pattern != "" {
		return pattern
	}

	if r.URL == nil {
		return "unknown"
	}

	path := strings.TrimSpace(r.URL.Path)
	if path == "" {
		return "/"
	}

	return path
}

func grpcMethodLabel(path string) string {
	path = strings.Trim(strings.TrimSpace(path), "/")
	if path == "" {
		return "unknown"
	}

	parts := strings.Split(path, "/")
	method := strings.TrimSpace(parts[len(parts)-1])
	if method == "" {
		return "unknown"
	}

	return method
}
