package tracing

import (
	"context"
	"testing"

	"github.com/kodokbakar/pylon/internal/config"
)

func TestInitTracerEnabledDoesNotFailOnResourceSchemaConflict(t *testing.T) {
	shutdown, err := InitTracer(context.Background(), "test-service", config.TracingConfig{
		Enabled:           true,
		CollectorEndpoint: "127.0.0.1:4317",
		ServiceVersion:    "test",
		SampleRatio:       1,
	})
	if err != nil {
		t.Fatalf("expected tracer initialization to succeed, got %v", err)
	}

	if shutdown == nil {
		t.Fatal("expected shutdown function")
	}

	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown tracer: %v", err)
	}
}
