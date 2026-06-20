package tracing

import (
	"context"
	"strings"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type KafkaHeaderCarrier struct {
	headers []kafka.Header
}

func InjectKafkaHeaders(ctx context.Context, headers []kafka.Header) []kafka.Header {
	carrier := &KafkaHeaderCarrier{
		headers: append([]kafka.Header(nil), headers...),
	}

	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier.headers
}

func ExtractKafkaContext(ctx context.Context, headers []kafka.Header) context.Context {
	carrier := &KafkaHeaderCarrier{
		headers: headers,
	}

	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

func (c *KafkaHeaderCarrier) Get(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}

	for _, header := range c.headers {
		if strings.EqualFold(header.Key, key) {
			return string(header.Value)
		}
	}

	return ""
}

func (c *KafkaHeaderCarrier) Set(key, value string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}

	filtered := c.headers[:0]
	for _, header := range c.headers {
		if strings.EqualFold(header.Key, key) {
			continue
		}

		filtered = append(filtered, header)
	}

	c.headers = append(filtered, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

func (c *KafkaHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for _, header := range c.headers {
		keys = append(keys, header.Key)
	}

	return keys
}
