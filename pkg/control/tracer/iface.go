package tracer

import (
	"context"

	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	traceapi "go.opentelemetry.io/otel/trace"
)

type shutdownTracerProvider interface {
	traceapi.TracerProvider
	Tracer(instrumentationName string, opts ...traceapi.TracerOption) traceapi.Tracer
	Shutdown(ctx context.Context) error
}

type shutdownLoggerProvider interface {
	log.LoggerProvider
	Logger(name string, options ...log.LoggerOption) log.Logger
	Shutdown(ctx context.Context) error
}

type shutdownMeterProvider interface {
	metric.MeterProvider
	Meter(name string, opts ...metric.MeterOption) metric.Meter
	Shutdown(ctx context.Context) error
}
