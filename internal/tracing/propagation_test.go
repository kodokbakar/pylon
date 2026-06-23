package tracing

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestKafkaHeaderCarrierSetGetKeys(t *testing.T) {
	carrier := &KafkaHeaderCarrier{}

	carrier.Set("traceparent", "value-1")
	carrier.Set("baggage", "value-2")

	if got := carrier.Get("traceparent"); got != "value-1" {
		t.Fatalf("expected traceparent value-1, got %q", got)
	}

	if got := carrier.Get("TRACEPARENT"); got != "value-1" {
		t.Fatalf("expected case-insensitive traceparent, got %q", got)
	}

	keys := carrier.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %#v", keys)
	}
}

func TestKafkaHeaderCarrierSetReplacesExistingValue(t *testing.T) {
	carrier := &KafkaHeaderCarrier{
		headers: []kafka.Header{
			{Key: "traceparent", Value: []byte("old")},
		},
	}

	carrier.Set("traceparent", "new")

	if got := carrier.Get("traceparent"); got != "new" {
		t.Fatalf("expected replaced traceparent, got %q", got)
	}

	if len(carrier.headers) != 1 {
		t.Fatalf("expected 1 header after replace, got %#v", carrier.headers)
	}
}

func TestInjectAndExtractKafkaHeadersPropagatesTraceContext(t *testing.T) {
	t.Setenv("OTEL_PROPAGATORS", "tracecontext")
	otel.SetTextMapPropagator(propagation.TraceContext{})

	traceID, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	if err != nil {
		t.Fatalf("parse trace id: %v", err)
	}

	spanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	if err != nil {
		t.Fatalf("parse span id: %v", err)
	}

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

	headers := InjectKafkaHeaders(ctx, []kafka.Header{
		{Key: "event_type", Value: []byte("message_created")},
	})

	extracted := ExtractKafkaContext(context.Background(), headers)
	extractedSpanContext := trace.SpanContextFromContext(extracted)

	if !extractedSpanContext.IsValid() {
		t.Fatal("expected valid extracted span context")
	}

	if extractedSpanContext.TraceID() != traceID {
		t.Fatalf("expected trace id %s, got %s", traceID, extractedSpanContext.TraceID())
	}

	if extractedSpanContext.SpanID() != spanID {
		t.Fatalf("expected span id %s, got %s", spanID, extractedSpanContext.SpanID())
	}

	if !extractedSpanContext.IsSampled() {
		t.Fatal("expected sampled trace flag")
	}

	if !extractedSpanContext.IsRemote() {
		t.Fatal("expected extracted span context to be remote")
	}
}
