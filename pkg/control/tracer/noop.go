package tracer

import (
	"context"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	traceapi "go.opentelemetry.io/otel/trace"
)

type noopShutdownTracerProvider struct {
	traceapi.TracerProvider
}

func (n *noopShutdownTracerProvider) Shutdown(ctx context.Context) error { return nil }

type noopShutdownLoggerProvider struct {
	log.LoggerProvider
}

func (n *noopShutdownLoggerProvider) Shutdown(ctx context.Context) error {
	return nil
}

type noopShutdownMeterProvider struct {
	metric.MeterProvider
}

func (n *noopShutdownMeterProvider) Shutdown(ctx context.Context) error {
	return nil
}
