package tracer

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type traceExporter struct {
	file           *os.File
	writerExporter *stdouttrace.Exporter
}

func newFileTraceExporter(filePath string) (*traceExporter, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0o666))
	if err != nil {
		return nil, fmt.Errorf("opening '%s' for OpenTelemetry file exporter: %w", filePath, err)
	}
	stdoutExporter, err := stdouttrace.New(stdouttrace.WithWriter(f))
	if err != nil {
		return nil, err
	}
	return &traceExporter{
		writerExporter: stdoutExporter,
		file:           f,
	}, nil
}

func (e *traceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return e.writerExporter.ExportSpans(ctx, spans)
}

func (e *traceExporter) Shutdown(ctx context.Context) error {
	if err := e.writerExporter.Shutdown(ctx); err != nil {
		return err
	}
	if err := e.file.Close(); err != nil {
		return fmt.Errorf("closing trace file: %w", err)
	}
	return nil
}

type loggerExporter struct {
	file           *os.File
	writerExporter *stdoutlog.Exporter
}

func (l *loggerExporter) Export(ctx context.Context, records []sdklog.Record) error {
	return l.writerExporter.Export(ctx, records)
}

func (l *loggerExporter) Shutdown(ctx context.Context) error {
	if err := l.writerExporter.Shutdown(ctx); err != nil {
		return err
	}
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("closing log file: %w", err)
	}
	return nil
}

func (l *loggerExporter) ForceFlush(ctx context.Context) error {
	return l.writerExporter.ForceFlush(ctx)
}

func newFileLoggerExporter(filePath string) (*loggerExporter, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0o666))
	if err != nil {
		return nil, fmt.Errorf("opening '%s' for OpenTelemetry file exporter: %w", filePath, err)
	}
	// stdoutExporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint(), stdoutlog.WithWriter(f))
	stdoutExporter, err := stdoutlog.New(stdoutlog.WithWriter(f))
	if err != nil {
		return nil, err
	}
	return &loggerExporter{
		writerExporter: stdoutExporter,
		file:           f,
	}, nil
}

type meterExporter struct {
	file           *os.File
	writerExporter sdkmetric.Exporter
}

func (m *meterExporter) Temporality(kind sdkmetric.InstrumentKind) metricdata.Temporality {
	return m.writerExporter.Temporality(kind)
}

func (m *meterExporter) Aggregation(kind sdkmetric.InstrumentKind) sdkmetric.Aggregation {
	return m.writerExporter.Aggregation(kind)
}

func (m *meterExporter) Export(ctx context.Context, resourceMetrics *metricdata.ResourceMetrics) error {
	return m.writerExporter.Export(ctx, resourceMetrics)
}

func (m *meterExporter) ForceFlush(ctx context.Context) error {
	return m.writerExporter.ForceFlush(ctx)
}

func (m *meterExporter) Shutdown(ctx context.Context) error {
	if err := m.writerExporter.Shutdown(ctx); err != nil {
		return err
	}
	if err := m.file.Close(); err != nil {
		return fmt.Errorf("closing log file: %w", err)
	}
	return nil
}

func newFileMeterExporter(filePath string) (*meterExporter, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0o666))
	if err != nil {
		return nil, fmt.Errorf("opening '%s' for OpenTelemetry file exporter: %w", filePath, err)
	}
	// stdoutExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint(), stdoutmetric.WithWriter(f))
	stdoutExporter, err := stdoutmetric.New(stdoutmetric.WithWriter(f))
	if err != nil {
		return nil, err
	}
	return &meterExporter{
		writerExporter: stdoutExporter,
		file:           f,
	}, nil
}
