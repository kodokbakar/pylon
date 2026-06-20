package tracing

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
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

func TestInjectAndExtractKafkaHeaders(t *testing.T) {
	propagator := propagation.TraceContext{}
	ctx := context.Background()

	carrier := &KafkaHeaderCarrier{}
	propagator.Inject(ctx, carrier)

	headers := InjectKafkaHeaders(ctx, carrier.headers)
	extracted := ExtractKafkaContext(ctx, headers)

	if extracted == nil {
		t.Fatal("expected extracted context")
	}
}
