package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics holds common application metrics
type Metrics struct {
	RequestCounter  metric.Int64Counter
	RequestDuration metric.Float64Histogram
	ErrorCounter    metric.Int64Counter
}

// InitMetrics initializes OpenTelemetry metrics
func InitMetrics(serviceName string) (*Metrics, error) {
	meter := otel.Meter(serviceName)

	requestCounter, err := meter.Int64Counter(
		"requests_total",
		metric.WithDescription("Total number of requests"),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(
		"request_duration_seconds",
		metric.WithDescription("Request duration in seconds"),
	)
	if err != nil {
		return nil, err
	}

	errorCounter, err := meter.Int64Counter(
		"errors_total",
		metric.WithDescription("Total number of errors"),
	)
	if err != nil {
		return nil, err
	}

	return &Metrics{
		RequestCounter:  requestCounter,
		RequestDuration: requestDuration,
		ErrorCounter:    errorCounter,
	}, nil
}

// RecordRequest records a request metric
func (m *Metrics) RecordRequest(ctx context.Context, labels ...metric.AddOption) {
	m.RequestCounter.Add(ctx, 1, labels...)
}

// RecordDuration records a request duration metric
func (m *Metrics) RecordDuration(ctx context.Context, duration float64, labels ...metric.RecordOption) {
	m.RequestDuration.Record(ctx, duration, labels...)
}

// RecordError records an error metric
func (m *Metrics) RecordError(ctx context.Context, labels ...metric.AddOption) {
	m.ErrorCounter.Add(ctx, 1, labels...)
}
