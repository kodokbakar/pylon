package tracing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kodokbakar/pylon/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	defaultCollectorEndpoint = "localhost:4317"
	defaultServiceVersion    = "1.0.0"
)

func InitTracer(ctx context.Context, serviceName string, cfg config.TracingConfig) (func(context.Context) error, error) {
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	if !cfg.Enabled {
		return func(context.Context) error {
			return nil
		}, nil
	}

	collectorEndpoint := strings.TrimSpace(cfg.CollectorEndpoint)
	if collectorEndpoint == "" {
		collectorEndpoint = defaultCollectorEndpoint
	}

	serviceVersion := strings.TrimSpace(cfg.ServiceVersion)
	if serviceVersion == "" {
		serviceVersion = defaultServiceVersion
	}

	sampleRatio := cfg.SampleRatio
	if sampleRatio <= 0 || sampleRatio > 1 {
		sampleRatio = 1
	}

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(collectorEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP trace exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create trace resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(sampleRatio),
			),
		),
	)

	otel.SetTracerProvider(provider)

	return provider.Shutdown, nil
}
