package otel

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	traceapi "go.opentelemetry.io/otel/trace"

	// semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

var (
	once sync.Once
	pp   atomic.Pointer[Control]
)

type Control struct {
	logger   *zap.SugaredLogger
	config   *config.OTELConfig
	ctx      context.Context
	provider shutdownTracerProvider
}

func NewOTELControl(otelConfig *config.OTELConfig, loggerControl logger.Control, ctx context.Context) (*Control, error) {
	if pp.Load() == nil {
		if otelConfig == nil {
			otelConfig = getDefaultConfig()
		}
		if !otelConfig.Enabled {
			return nil, errors.New("otel is not enabled")
		}

		otelLogger := loggerControl.GenLogger("otel")

		control := &Control{
			logger: otelLogger,
			config: otelConfig,
			ctx:    ctx,
		}
		control.setProvider()
		pp.Store(control)
		// return control, nil
	}
	return pp.Load(), nil
}

func (c *Control) initExporters() ([]sdktrace.SpanExporter, error) {
	var exporters []sdktrace.SpanExporter
	for _, exporterStr := range strings.Split(c.config.TracesExporter, ",") {
		exporterStr = strings.TrimSpace(exporterStr)
		switch exporterStr {
		case "otlp":
			protocol := "http/protobuf"
			if !stringer.IsBlank(c.config.ExporterOTELProtocol) {
				protocol = c.config.ExporterOTELProtocol
			}
			switch protocol {
			case "http/protobuf":
				exporter, err := otlptracehttp.New(c.ctx)
				if err != nil {
					return nil, fmt.Errorf("building OTLP HTTP exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			case "grpc":
				exporter, err := otlptracegrpc.New(c.ctx)
				if err != nil {
					return nil, fmt.Errorf("building OTLP gRPC exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			default:
				return nil, fmt.Errorf("unknown or unsupported OTLP exporter '%s'", exporterStr)
			}
		case "zipkin":
			exporter, err := zipkin.New(c.config.ExporterZipkinEndpoint)
			if err != nil {
				return nil, fmt.Errorf("building Zipkin exporter: %w", err)
			}
			exporters = append(exporters, exporter)
		case "file":
			if stringer.IsBlank(c.config.ExporterFilePath) {
				wd, err := path.GetWorkDir()
				if err != nil {
					return nil, fmt.Errorf("finding working directory for the OpenTelemetry file exporter: %w", err)
				}
				c.config.ExporterFilePath = filepath.Join(wd, "traces.json")
			}
			exporter, err := newFileExporter(c.config.ExporterFilePath)
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		case "none":
			continue
		case "":
			continue
		case "stdout":
			exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		default:
			return nil, fmt.Errorf("unknown or unsupported exporter '%s'", exporterStr)
		}
	}
	return exporters, nil
}

func (c *Control) newTracerProvider() error {
	c.logger.Debugf("[control] starting generate tracer provider...")
	exporters, err := c.initExporters()
	if err != nil {
		c.logger.Errorf("[control] init exporters failed: %v", err)
		return nil
	}
	if len(exporters) == 0 {
		c.logger.Warnf("[control] no exporter found, using noop tracer provider...")
		c.provider = &noopShutdownTracerProvider{TracerProvider: noop.NewTracerProvider()}
		return nil
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
		c.logger.Errorf("[control] create resource failed: %v", err)
		return err
	}

	r, err := resource.Merge(
		res,
		// resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String(c.config.ExporterServiceName),
		),
	)
	if err != nil {
		c.logger.Errorf("[control] merge resource failed: %v", err)
		return err
	}
	options = append(options, sdktrace.WithResource(r))
	// options = append(options, sdktrace.WithSampler(sdktrace.AlwaysSample())) // 或者自定义采样器
	c.provider = sdktrace.NewTracerProvider(options...)
	return nil
}

func (c *Control) setProvider() {
	c.logger.Debugf("[control] setting provider...")
	otel.SetTracerProvider(c.provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	// otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	tracer := c.provider.Tracer(c.config.ExporterServiceName)
	ctx, span := tracer.Start(c.ctx, "init")
	defer span.End()
	c.ctx = ctx
}

func (c *Control) Span(componentName string, spanName string, opts ...traceapi.SpanStartOption) (context.Context, traceapi.Span) {
	return c.provider.Tracer(c.config.ExporterServiceName).Start(c.ctx, fmt.Sprintf("%s.%s", componentName, spanName), opts...)
}
