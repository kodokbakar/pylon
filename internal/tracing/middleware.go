package tracing

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func HTTPMiddleware(operationName string, next http.Handler) http.Handler {
	operationName = strings.TrimSpace(operationName)
	if operationName == "" {
		operationName = "http"
	}

	return otelhttp.NewHandler(next, operationName)
}

func HTTPTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return otelhttp.NewTransport(base)
}
