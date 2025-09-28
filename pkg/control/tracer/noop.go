package tracer

import (
	"context"

	traceapi "go.opentelemetry.io/otel/trace"
)

type noopShutdownTracerProvider struct {
	traceapi.TracerProvider
}

func (n *noopShutdownTracerProvider) Shutdown(ctx context.Context) error { return nil }
