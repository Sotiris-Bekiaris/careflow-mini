package observability

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// InitTracer initializes the OpenTelemetry tracer with OTLP exporter
func InitTracer(serviceName, otlpEndpoint string) (func(context.Context) error, error) {
	// Create OTLP HTTP exporter
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(), // For local development
	)
	if err != nil {
		return nil, err
	}

	// Create tracer provider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// Register as global tracer provider
	otel.SetTracerProvider(tp)

	log.Printf("OpenTelemetry tracer initialized for service: %s with OTLP endpoint: %s", serviceName, otlpEndpoint)

	return tp.Shutdown, nil
}

// GetTracer returns a tracer for the given name
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
