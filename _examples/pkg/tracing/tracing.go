package tracing

import (
	"context"
	"fmt"
	"log"

	"_examples/pkg/str"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace/noop"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	// semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	traceapi "go.opentelemetry.io/otel/trace"
)

var service string

// shutdownTracerProvider adds a shutdown method for tracer providers.
type shutdownTracerProvider interface {
	traceapi.TracerProvider
	Tracer(instrumentationName string, opts ...traceapi.TracerOption) traceapi.Tracer
	Shutdown(ctx context.Context) error
}

// noopShutdownTracerProvider adds a no-op Shutdown method to a TracerProvider.
type noopShutdownTracerProvider struct{ traceapi.TracerProvider }

func (n *noopShutdownTracerProvider) Shutdown(ctx context.Context) error { return nil }

// NewTracerProvider creates and configures a TracerProvider.
func NewTracerProvider(ctx context.Context, serviceName string) (shutdownTracerProvider, error) {
	provider, err := myNewTracerProvider(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	// otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	service = serviceName
	tracer := provider.Tracer(serviceName)
	ctx, span := tracer.Start(ctx, "init")
	defer span.End()
	return provider, nil
}

func myNewTracerProvider(ctx context.Context, serviceName string) (shutdownTracerProvider, error) {
	exporters, err := newSpanExporters(ctx)
	if err != nil {
		return nil, err
	}
	if len(exporters) == 0 {
		return &noopShutdownTracerProvider{TracerProvider: noop.NewTracerProvider()}, nil
	}
	var options []sdktrace.TracerProviderOption

	for _, exporter := range exporters {
		options = append(options, sdktrace.WithBatcher(exporter))
	}

	// 创建 resource
	res, err := resource.New(
		context.Background(),
		// resource.WithAttributes(
		// 	semconv.ServiceNameKey.String(serviceName), // 使用 semconv.ServiceNameKey
		// ),
		resource.WithOS(),
		resource.WithProcessPID(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}
	r, err := resource.Merge(
		res,
		// resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}
	options = append(options, sdktrace.WithResource(r))
	// options = append(options, sdktrace.WithSampler(sdktrace.AlwaysSample())) // 或者自定义采样器

	return sdktrace.NewTracerProvider(options...), nil
}

// Span starts a new span using the standard conventions.
func Span(ctx context.Context, componentName string, spanName string, opts ...traceapi.SpanStartOption) (context.Context, traceapi.Span) {
	if str.CompareIgnoreCase("", service) {
		log.Fatalf("provider is not initialized")
	}
	return otel.Tracer(service).Start(ctx, fmt.Sprintf("%s.%s", componentName, spanName), opts...)
}

// StartSpan 创建并启动一个 Span (可以根据需要添加 attributes)
func StartSpan(ctx context.Context, componentName string, spanName string, attributes ...attribute.KeyValue) (context.Context, traceapi.Span) {
	return Span(ctx, componentName, spanName, traceapi.WithAttributes(attributes...))
}
