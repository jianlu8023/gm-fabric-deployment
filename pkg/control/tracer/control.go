package tracer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	// semconv "go.opentelemetry.io/otel/semconv/v1.5.0"
	// semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	traceapi "go.opentelemetry.io/otel/trace"

	lognoop "go.opentelemetry.io/otel/log/noop"
	meternoop "go.opentelemetry.io/otel/metric/noop"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

var (
	pool = sync.Pool{
		New: func() any {
			control, err := NewTracerControl(nil, logger.NewLoggerControl(&config.LoggerConfig{
				DefaultLogLevel: "info",
				PrintFormat:     "console",
			}), context.Background())
			if err != nil {
				return nil
			}
			return control
		},
	}

	// singleton instance and mutex for thread-safe initialization
	instance     *Control
	instanceMu   sync.RWMutex
	instanceOnce sync.Once
)

type Control struct {
	logger         *zap.SugaredLogger
	config         *config.TracerConfig
	once           sync.Once
	ctx            context.Context
	traceLogger    logr.Logger
	providerMutex  sync.RWMutex
	tracerProvider shutdownTracerProvider
	meterProvider  shutdownMeterProvider
	loggerProvider shutdownLoggerProvider
	traceApi       traceapi.Tracer
}

// NewTracerControl 创建一个新的OTEL控制器并设置为全局单例
// @description 创建并初始化一个OTEL控制器实例，用于管理OpenTelemetry相关功能，并自动设置为全局单例
// @param tracerConfig *config.OTELConfig OTEL配置信息
// @param loggerControl logger.Control 日志控制器
// @param ctx context.Context 上下文，如果为nil则使用context.Background()
// @return *Control OTEL控制器实例
// @return error 初始化过程中可能出现的错误
func NewTracerControl(tracerConfig *config.TracerConfig, loggerControl *logger.Control, ctx context.Context) (*Control, error) {
	if tracerConfig == nil {
		tracerConfig = getDefaultConfig()
	}
	if !tracerConfig.Enabled {
		return nil, errors.New("tracer is not enabled")
	}

	tracerLogger := loggerControl.GenLogger(logger.ModuleTracer)

	// 如果传入的ctx为nil，则使用context.Background()作为默认值
	if ctx == nil {
		tracerLogger.Warnf("[control] tracer context is nil, using new context...")
		ctx = context.Background()
	}

	control := &Control{
		logger:      tracerLogger,
		config:      tracerConfig,
		ctx:         ctx,
		traceLogger: newTracerCustomLogger(loggerControl.GetConfig(), tracerConfig.LogInConsole),
	}
	otel.SetLogger(control.traceLogger)
	if err := control.setProvider(); err != nil {
		control.logger.Errorf("[control] set provider failed: %v", err)
		return nil, err
	}

	// 自动设置为全局单例实例
	// SetInstance内部使用了instanceOnce.Do，确保只会设置一次
	control.logger.Debugf("[control] setting singleton instance...")
	setInstance(control)

	return control, nil
}
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Debugf("[control] starting tracer server...")
			c.init()
		}
	})
}

func (c *Control) Shutdown() error {
	c.logger.Debug("[control] shutting down tracer server...")
	if err := c.tracerProvider.Shutdown(c.ctx); err != nil {
		c.logger.Errorf("[control] shutting down tracer provider failed: %v", err)
		return err
	}
	if err := c.loggerProvider.Shutdown(c.ctx); err != nil {
		c.logger.Errorf("[control] shutting down logger provider failed: %v", err)
		return err
	}
	if err := c.meterProvider.Shutdown(c.ctx); err != nil {
		c.logger.Errorf("[control] shutting down meter provider failed: %v", err)
		return err
	}
	return nil
}

func (c *Control) TracerProvider() shutdownTracerProvider {
	c.logger.Debugf("[control] get tracer provider...")
	c.providerMutex.RLock()
	defer c.providerMutex.RUnlock()
	return c.tracerProvider
}

func (c *Control) LoggerProvider() shutdownLoggerProvider {
	c.logger.Debugf("[control] get logger provider...")
	c.providerMutex.RLock()
	defer c.providerMutex.RUnlock()
	return c.loggerProvider
}

func (c *Control) MeterProvider() shutdownMeterProvider {
	c.logger.Debugf("[control] get meter provider...")
	c.providerMutex.RLock()
	defer c.providerMutex.RUnlock()
	return c.meterProvider
}

func (c *Control) initTracerExporters() ([]sdktrace.SpanExporter, error) {
	var exporters []sdktrace.SpanExporter
	for _, exporterStr := range strings.Split(c.config.Tracer.Exporters, ",") {
		exporterStr = strings.TrimSpace(exporterStr)
		switch exporterStr {
		case ExporterOtlp:
			protocol := ExporterOtlpProtocolHttp
			if !stringer.IsBlank(c.config.Tracer.OTELProtocol) {
				protocol = c.config.Tracer.OTELProtocol
			}
			switch protocol {
			case ExporterOtlpProtocolHttp:
				opts := []otlptracehttp.Option{
					otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
					otlptracehttp.WithTimeout(10 * time.Second),
					otlptracehttp.WithRetry(otlptracehttp.RetryConfig{
						Enabled:         true,
						InitialInterval: 10 * time.Second,
						MaxInterval:     10 * time.Second,
						MaxElapsedTime:  5 * time.Minute,
					}),
				}
				if c.config.Tracer.OTELInsecure {
					opts = append(opts, otlptracehttp.WithInsecure())
				}
				if !stringer.IsBlank(c.config.Tracer.OTELEndpoint) {
					endpoint := c.config.Tracer.OTELEndpoint

					parsedURL, err := url.Parse(endpoint)
					if err == nil && (stringer.CompareIgnoreCase(parsedURL.Scheme, "http") ||
						stringer.CompareIgnoreCase(parsedURL.Scheme, "https")) {
						// 有schema
						c.logger.Debugf("[control] http/protobuf protocol endpoint (with scheme): %v", endpoint)
						// 检查是否有path，如果没有则添加/v1/traces
						if stringer.IsBlank(parsedURL.Path) || parsedURL.Path == "/" {
							if parsedURL.Path == "/" {
								endpoint = endpoint + "v1/traces"
							} else {
								endpoint = endpoint + "/v1/traces"
							}
						}
						opts = append(opts, otlptracehttp.WithEndpointURL(endpoint))
					} else {
						// 没有schema
						c.logger.Debugf("[control] http/protobuf protocol endpoint (without scheme): %v", endpoint)
						opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
					}
				}
				exporter, err := otlptracehttp.New(c.ctx, opts...)
				if err != nil {
					return nil, fmt.Errorf("building OTLP HTTP exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			case ExporterOtlpProtocolGrpc:
				opts := []otlptracegrpc.Option{
					otlptracegrpc.WithCompressor("gzip"),
					otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
						Enabled:         true,
						InitialInterval: 10 * time.Second,
						MaxInterval:     10 * time.Second,
						MaxElapsedTime:  5 * time.Minute,
					}),
				}
				if c.config.Tracer.OTELInsecure {
					opts = append(opts, otlptracegrpc.WithInsecure())
				}
				if !stringer.IsBlank(c.config.Tracer.OTELEndpoint) {
					endpoint := c.config.Tracer.OTELEndpoint
					// 直接使用url.Parse解析
					parsedURL, err := url.Parse(endpoint)
					if err == nil && (stringer.CompareIgnoreCase(parsedURL.Scheme, "http") ||
						stringer.CompareIgnoreCase(parsedURL.Scheme, "https")) {
						// 有schema
						c.logger.Debugf("[control] grpc protocol endpoint (with scheme): %v", endpoint)
						// 检查是否有path，如果没有则添加/v1/traces
						if stringer.IsBlank(parsedURL.Path) || parsedURL.Path == "/" {
							if parsedURL.Path == "/" {
								endpoint = endpoint + "v1/traces"
							} else {
								endpoint = endpoint + "/v1/traces"
							}
						}
						opts = append(opts, otlptracegrpc.WithEndpointURL(endpoint))
					} else {
						// 没有schema
						c.logger.Debugf("[control] grpc protocol endpoint (without scheme): %v", endpoint)
						opts = append(opts, otlptracegrpc.WithEndpoint(endpoint))
					}
				}
				exporter, err := otlptracegrpc.New(c.ctx, opts...)
				if err != nil {
					return nil, fmt.Errorf("building OTLP gRPC exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			default:
				return nil, fmt.Errorf("unknown or unsupported OTLP exporter '%s'", exporterStr)
			}
		case ExporterZipkin:
			var opts []zipkin.Option
			opts = append(opts, zipkin.WithLogr(c.traceLogger))
			exporter, err := zipkin.New(c.config.Tracer.ZipkinEndpoint, opts...)
			if err != nil {
				return nil, fmt.Errorf("building Zipkin exporter: %w", err)
			}
			exporters = append(exporters, exporter)
		case ExporterFile:
			if stringer.IsBlank(c.config.Tracer.FilePath) {
				wd, err := path.GetWorkDir()
				if err != nil {
					return nil, fmt.Errorf("finding working directory for the OpenTelemetry file exporter: %w", err)
				}
				c.config.Tracer.FilePath = filepath.Join(wd, "traces.json")
			}
			exporter, err := newFileTraceExporter(c.config.Tracer.FilePath)
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		case ExporterNone:
			continue
		case ExporterBlack:
			continue
		case ExporterStdout:
			exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		default:
			c.logger.Warnf("[control] unknown or unsupported exporter '%s'", exporterStr)
			// return nil, fmt.Errorf("unknown or unsupported exporter '%s'", exporterStr)
		}
	}
	return exporters, nil
}

func (c *Control) initLoggerExporters() ([]sdklog.Exporter, error) {
	var exporters []sdklog.Exporter
	for _, exporterStr := range strings.Split(c.config.Logger.Exporters, ",") {
		exporterStr = strings.TrimSpace(exporterStr)
		switch exporterStr {
		case ExporterOtlp:
			protocol := ExporterOtlpProtocolHttp
			if !stringer.IsBlank(c.config.Logger.OTELProtocol) {
				protocol = c.config.Logger.OTELProtocol
			}
			switch protocol {
			case ExporterOtlpProtocolHttp:
				opts := []otlploghttp.Option{
					otlploghttp.WithCompression(otlploghttp.GzipCompression),
					otlploghttp.WithTimeout(10 * time.Second),
					otlploghttp.WithRetry(otlploghttp.RetryConfig{
						Enabled:         true,
						InitialInterval: 10 * time.Second,
						MaxInterval:     10 * time.Second,
						MaxElapsedTime:  5 * time.Minute,
					}),
				}
				if c.config.Logger.OTELInsecure {
					opts = append(opts, otlploghttp.WithInsecure())
				}
				if !stringer.IsBlank(c.config.Logger.OTELEndpoint) {
					endpoint := c.config.Logger.OTELEndpoint

					parsedURL, err := url.Parse(endpoint)
					if err == nil && (stringer.CompareIgnoreCase(parsedURL.Scheme, "http") ||
						stringer.CompareIgnoreCase(parsedURL.Scheme, "https")) {
						// 有schema
						c.logger.Debugf("[control] http/protobuf protocol endpoint (with scheme): %v", endpoint)
						// 检查是否有path，如果没有则添加/v1/logs
						if stringer.IsBlank(parsedURL.Path) || parsedURL.Path == "/" {
							if parsedURL.Path == "/" {
								endpoint = endpoint + "v1/logs"
							} else {
								endpoint = endpoint + "/v1/logs"
							}
						}
						opts = append(opts, otlploghttp.WithEndpointURL(endpoint))
					} else {
						// 没有schema
						c.logger.Debugf("[control] http/protobuf protocol endpoint (without scheme): %v", endpoint)
						opts = append(opts, otlploghttp.WithEndpoint(endpoint))
					}
				}
				exporter, err := otlploghttp.New(c.ctx, opts...)
				if err != nil {
					return nil, fmt.Errorf("building OTLP HTTP exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			case ExporterOtlpProtocolGrpc:
				opts := []otlploggrpc.Option{
					otlploggrpc.WithCompressor("gzip"),
					otlploggrpc.WithRetry(otlploggrpc.RetryConfig{
						Enabled:         true,
						InitialInterval: 10 * time.Second,
						MaxInterval:     10 * time.Second,
						MaxElapsedTime:  5 * time.Minute,
					}),
				}
				if c.config.Logger.OTELInsecure {
					opts = append(opts, otlploggrpc.WithInsecure())
				}
				if !stringer.IsBlank(c.config.Logger.OTELEndpoint) {
					endpoint := c.config.Logger.OTELEndpoint
					// 直接使用url.Parse解析
					parsedURL, err := url.Parse(endpoint)
					if err == nil && (stringer.CompareIgnoreCase(parsedURL.Scheme, "http") ||
						stringer.CompareIgnoreCase(parsedURL.Scheme, "https")) {
						// 有schema
						c.logger.Debugf("[control] grpc protocol endpoint (with scheme): %v", endpoint)
						// 检查是否有path，如果没有则添加/v1/logs
						if stringer.IsBlank(parsedURL.Path) || parsedURL.Path == "/" {
							if parsedURL.Path == "/" {
								endpoint = endpoint + "v1/logs"
							} else {
								endpoint = endpoint + "/v1/logs"
							}
						}
						opts = append(opts, otlploggrpc.WithEndpointURL(endpoint))
					} else {
						// 没有schema
						c.logger.Debugf("[control] grpc protocol endpoint (without scheme): %v", endpoint)
						opts = append(opts, otlploggrpc.WithEndpoint(endpoint))
					}
				}
				exporter, err := otlploggrpc.New(c.ctx, opts...)
				if err != nil {
					return nil, fmt.Errorf("building OTLP gRPC exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			default:
				return nil, fmt.Errorf("unknown or unsupported OTLP exporter '%s'", exporterStr)
			}
		case ExporterNone:
			continue
		case ExporterBlack:
			continue
		case ExporterStdout:
			exporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		case ExporterFile:
			if stringer.IsBlank(c.config.Logger.FilePath) {
				wd, err := path.GetWorkDir()
				if err != nil {
					return nil, fmt.Errorf("finding working directory for the OpenTelemetry file exporter: %w", err)
				}
				c.config.Logger.FilePath = filepath.Join(wd, "logs.json")
			}
			exporter, err := newFileLoggerExporter(c.config.Logger.FilePath)
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		default:
			c.logger.Warnf("[control] unknown or unsupported exporter '%s'", exporterStr)
			// return nil, fmt.Errorf("unknown or unsupported exporter '%s'", exporterStr)
		}
	}
	return exporters, nil
}

func (c *Control) initMeterExporters() ([]sdkmetric.Exporter, error) {
	var exporters []sdkmetric.Exporter
	for _, exporterStr := range strings.Split(c.config.Meter.Exporters, ",") {
		exporterStr = strings.TrimSpace(exporterStr)
		switch exporterStr {
		case ExporterOtlp:
			protocol := ExporterOtlpProtocolHttp
			if !stringer.IsBlank(c.config.Meter.OTELProtocol) {
				protocol = c.config.Meter.OTELProtocol
			}
			switch protocol {
			case ExporterOtlpProtocolHttp:
				opts := []otlpmetrichttp.Option{
					otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
					otlpmetrichttp.WithTimeout(10 * time.Second),
					otlpmetrichttp.WithRetry(otlpmetrichttp.RetryConfig{
						Enabled:         true,
						InitialInterval: 10 * time.Second,
						MaxInterval:     10 * time.Second,
						MaxElapsedTime:  5 * time.Minute,
					}),
				}
				if c.config.Meter.OTELInsecure {
					opts = append(opts, otlpmetrichttp.WithInsecure())
				}
				if !stringer.IsBlank(c.config.Meter.OTELEndpoint) {
					endpoint := c.config.Meter.OTELEndpoint

					parsedURL, err := url.Parse(endpoint)
					if err == nil && (stringer.CompareIgnoreCase(parsedURL.Scheme, "http") ||
						stringer.CompareIgnoreCase(parsedURL.Scheme, "https")) {
						// 有schema
						c.logger.Debugf("[control] http/protobuf protocol endpoint (with scheme): %v", endpoint)
						// 检查是否有path，如果没有则添加/v1/metrics
						if stringer.IsBlank(parsedURL.Path) || parsedURL.Path == "/" {
							if parsedURL.Path == "/" {
								endpoint = endpoint + "v1/metrics"
							} else {
								endpoint = endpoint + "/v1/metrics"
							}
						}
						opts = append(opts, otlpmetrichttp.WithEndpointURL(endpoint))
					} else {
						// 没有schema
						c.logger.Debugf("[control] http/protobuf protocol endpoint (without scheme): %v", endpoint)
						opts = append(opts, otlpmetrichttp.WithEndpoint(endpoint))
					}
				}
				exporter, err := otlpmetrichttp.New(c.ctx, opts...)
				if err != nil {
					return nil, fmt.Errorf("building OTLP HTTP exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			case ExporterOtlpProtocolGrpc:
				opts := []otlpmetricgrpc.Option{
					otlpmetricgrpc.WithCompressor("gzip"),
					otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
						Enabled:         true,
						InitialInterval: 10 * time.Second,
						MaxInterval:     10 * time.Second,
						MaxElapsedTime:  5 * time.Minute,
					}),
				}
				if c.config.Meter.OTELInsecure {
					opts = append(opts, otlpmetricgrpc.WithInsecure())
				}
				if !stringer.IsBlank(c.config.Meter.OTELEndpoint) {
					endpoint := c.config.Meter.OTELEndpoint
					// 直接使用url.Parse解析
					parsedURL, err := url.Parse(endpoint)
					if err == nil && (stringer.CompareIgnoreCase(parsedURL.Scheme, "http") ||
						stringer.CompareIgnoreCase(parsedURL.Scheme, "https")) {
						// 有schema
						c.logger.Debugf("[control] grpc protocol endpoint (with scheme): %v", endpoint)
						// 检查是否有path，如果没有则添加/v1/metrics
						if stringer.IsBlank(parsedURL.Path) || parsedURL.Path == "/" {
							if parsedURL.Path == "/" {
								endpoint = endpoint + "v1/metrics"
							} else {
								endpoint = endpoint + "/v1/metrics"
							}
						}
						opts = append(opts, otlpmetricgrpc.WithEndpointURL(endpoint))
					} else {
						// 没有schema
						c.logger.Debugf("[control] grpc protocol endpoint (without scheme): %v", endpoint)
						opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))
					}
				}
				exporter, err := otlpmetricgrpc.New(c.ctx, opts...)
				if err != nil {
					return nil, fmt.Errorf("building OTLP gRPC exporter: %w", err)
				}
				exporters = append(exporters, exporter)
			default:
				return nil, fmt.Errorf("unknown or unsupported OTLP exporter '%s'", exporterStr)
			}
		case ExporterNone:
			continue
		case ExporterBlack:
			continue
		case ExporterStdout:
			exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		case ExporterFile:
			if stringer.IsBlank(c.config.Meter.FilePath) {
				wd, err := path.GetWorkDir()
				if err != nil {
					return nil, fmt.Errorf("finding working directory for the OpenTelemetry file exporter: %w", err)
				}
				c.config.Meter.FilePath = filepath.Join(wd, "metrics.json")
			}
			exporter, err := newFileMeterExporter(c.config.Meter.FilePath)
			if err != nil {
				return nil, err
			}
			exporters = append(exporters, exporter)
		default:
			c.logger.Warnf("[control] unknown or unsupported exporter '%s'", exporterStr)
			// 	return nil, fmt.Errorf("unknown or unsupported exporter '%s'", exporterStr)
		}
	}
	return exporters, nil
}

func (c *Control) GetServiceName() string {
	c.logger.Debugf("[control] get service name...")
	return c.config.ServiceName
}

func (c *Control) newProvider() error {
	c.logger.Debugf("[control] starting generate provider...")

	// 创建 resourceEnd
	resMid, err := resource.New(
		context.Background(),
		// resourceEnd.WithAttributes(
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

	resourceEnd, err := resource.Merge(
		resMid,
		// resourceEnd.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String(c.config.ServiceName),
			semconv.ServiceInstanceIDKey.String(uuid.GetUUID()),
			semconv.ServiceVersionKey.String(version.Version),
		),
	)
	if err != nil {
		c.logger.Errorf("[control] merge resource failed: %v", err)
		return err
	}
	if err = c.newTracerProvider(resourceEnd); err != nil {
		c.logger.Errorf("[control] new tracer provider failed: %v", err)
		return err
	}
	if err = c.newLoggerProvider(resourceEnd); err != nil {
		c.logger.Errorf("[control] new logger provider failed: %v", err)
		return err
	}
	if err = c.newMeterProvider(resourceEnd); err != nil {
		c.logger.Errorf("[control] new meter provider failed: %v", err)
		return err
	}
	return nil
}

func (c *Control) newTracerProvider(resource *resource.Resource) error {
	exporters, err := c.initTracerExporters()
	if err != nil {
		c.logger.Errorf("[control] init tracer exporters failed: %v", err)
		return err
	}
	if len(exporters) == 0 {
		c.logger.Warnf("[control] no exporter found, using noop tracerProvider...")
		c.tracerProvider = &noopShutdownTracerProvider{TracerProvider: noop.NewTracerProvider()}
		return nil
	}
	var options []sdktrace.TracerProviderOption
	for _, exporter := range exporters {
		options = append(options, sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)))
	}
	options = append(options, sdktrace.WithResource(resource))
	options = append(options, sdktrace.WithSampler(sdktrace.AlwaysSample())) // 或者自定义采样器
	c.providerMutex.Lock()
	c.tracerProvider = sdktrace.NewTracerProvider(options...)
	c.providerMutex.Unlock()
	return nil
}

func (c *Control) newLoggerProvider(resource *resource.Resource) error {
	exporters, err := c.initLoggerExporters()
	if err != nil {
		c.logger.Errorf("[control] init logger exporters failed: %v", err)
		return err
	}
	if len(exporters) == 0 {
		c.logger.Warnf("[control] no exporter found, using noop loggerProvider...")
		c.loggerProvider = &noopShutdownLoggerProvider{LoggerProvider: lognoop.NewLoggerProvider()}
		return nil
	}

	var opts []sdklog.LoggerProviderOption
	for _, exporter := range exporters {
		opts = append(opts, sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)))
	}
	opts = append(opts, sdklog.WithResource(resource))
	c.providerMutex.Lock()
	c.loggerProvider = sdklog.NewLoggerProvider(opts...)
	c.providerMutex.Unlock()
	return nil
}

func (c *Control) newMeterProvider(resource *resource.Resource) error {
	exporters, err := c.initMeterExporters()
	if err != nil {
		c.logger.Errorf("[control] init meter exporters failed: %v", err)
		return err
	}
	if len(exporters) == 0 {
		c.logger.Warnf("[control] no exporter found, using noop meterProvider...")
		c.meterProvider = &noopShutdownMeterProvider{MeterProvider: meternoop.NewMeterProvider()}
		return nil
	}
	var opts []sdkmetric.Option
	for _, exporter := range exporters {
		opts = append(opts, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
	}
	opts = append(opts, sdkmetric.WithResource(resource))
	opts = append(opts, sdkmetric.WithExemplarFilter(exemplar.AlwaysOnFilter))
	c.providerMutex.Lock()
	c.meterProvider = sdkmetric.NewMeterProvider(opts...)
	c.providerMutex.Unlock()
	return nil
}

func (c *Control) setProvider() error {
	c.logger.Debugf("[control] setting provider...")
	if err := c.newProvider(); err != nil {
		c.logger.Errorf("[control] new provider failed: %v", err)
		return err
	}
	otel.SetLogger(c.traceLogger)
	c.providerMutex.RLock()
	otel.SetTracerProvider(c.tracerProvider)
	c.traceApi = c.tracerProvider.Tracer(
		c.config.ServiceName,
		traceapi.WithInstrumentationVersion(version.Version),
		traceapi.WithInstrumentationAttributes(
			attribute.String("", ""),
		),
		traceapi.WithSchemaURL(""),
	)
	c.providerMutex.RUnlock()
	otel.SetMeterProvider(c.meterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	// otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	return nil
}

func (c *Control) init() {
	c.logger.Debugf("[control] init...")
	ctx, span := c.traceApi.Start(c.ctx, "initialize")
	defer span.End()
	c.ctx = ctx
}

// Span 创建一个新的span
// @description 创建一个新的span，使用传入的context作为父上下文，实现全链路追踪
// @param ctx context.Context 父上下文，用于链路追踪
// @param componentName string 组件名称
// @param spanName string span名称
// @param opts ...traceapi.SpanStartOption span选项
// @return context.Context 包含新span的上下文
// @return traceapi.Span 创建的span
func (c *Control) Span(ctx context.Context, componentName string, spanName string, opts ...traceapi.SpanStartOption) (tCtx context.Context, span traceapi.Span) {
	// 如果传入的ctx为nil，则使用控制器的ctx作为后备
	if ctx == nil {
		ctx = c.ctx
	}
	if stringer.IsBlank(componentName) {
		componentName = "defaultComponent"
	}
	if stringer.IsBlank(spanName) {
		spanName = "defaultSpanName"
	}
	return c.traceApi.Start(ctx, fmt.Sprintf("%s.%s", componentName, spanName), opts...)
}

// StartSpan 创建并启动一个 Span (可以根据需要添加 attributes)
func (c *Control) StartSpan(ctx context.Context, componentName string, spanName string, attributes ...attribute.KeyValue) (tCtx context.Context, span traceapi.Span) {
	return c.Span(ctx, componentName, spanName, traceapi.WithAttributes(attributes...))
}

// getInstance 获取tracer控制器的单例实例
// @description 获取全局唯一的tracer控制器实例，确保全链路追踪的一致性
// @return *Control tracer控制器的单例实例
func getInstance() *Control {
	instanceMu.RLock()
	result := instance
	instanceMu.RUnlock()
	return result
}

// setInstance 设置tracer控制器的单例实例
// @description 设置全局唯一的tracer控制器实例，通常在应用初始化时调用一次
// @param control *Control tracer控制器实例
func setInstance(control *Control) {
	instanceOnce.Do(func() {
		instanceMu.Lock()
		instance = control
		instanceMu.Unlock()
	})
}

func Span(ctx context.Context, componentName string, spanName string, opts ...traceapi.SpanStartOption) (tCtx context.Context, span traceapi.Span) {
	// 使用单例实例，如果单例实例不存在则从池中获取
	control := getInstance()
	if control == nil {
		// 如果单例还未初始化，使用池中的实例作为后备
		control = pool.Get().(*Control)
		defer pool.Put(control) // 使用完后放回池中
	}
	return control.Span(ctx, componentName, spanName, opts...)
}

// StartSpan 创建并启动一个 Span (可以根据需要添加 attributes)
// @description 创建并启动一个带有属性的Span，优先使用单例控制器实例
// @param ctx context.Context 父上下文，用于链路追踪
// @param componentName string 组件名称
// @param spanName string span名称
// @param attributes ...attribute.KeyValue span属性
// @return context.Context 包含新span的上下文
// @return traceapi.Span 创建的span
func StartSpan(ctx context.Context, componentName string, spanName string, attributes ...attribute.KeyValue) (tCtx context.Context, span traceapi.Span) {
	return Span(ctx, componentName, spanName, traceapi.WithAttributes(attributes...))
}
