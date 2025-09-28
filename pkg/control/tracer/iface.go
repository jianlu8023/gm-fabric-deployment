package tracer

import (
	"context"

	traceapi "go.opentelemetry.io/otel/trace"
)

type shutdownTracerProvider interface {
	traceapi.TracerProvider
	Tracer(instrumentationName string, opts ...traceapi.TracerOption) traceapi.Tracer
	Shutdown(ctx context.Context) error
}
