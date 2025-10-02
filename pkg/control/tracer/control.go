package tracer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	//	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	// "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"path/filepath"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	// semconv "go.opentelemetry.io/otel/semconv/v1.5.0"
	traceapi "go.opentelemetry.io/otel/trace"

	// semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

var (
	pool = sync.Pool{
		New: func() any {
			control, err := NewTracerControl(nil, logger.NewLoggerControl(nil), nil)
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
	logger        *zap.SugaredLogger
	config        *config.TracerConfig
	once          sync.Once
	ctx           context.Context
	providerMutex sync.RWMutex
	provider      shutdownTracerProvider
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
		logger: tracerLogger,
		config: tracerConfig,
		ctx:    ctx,
	}
	if err := control.setProvider(); err != nil {
		control.logger.Errorf("[control] set provider failed: %v", err)
		return nil, err
	}

	// 自动设置为全局单例实例
	// SetInstance内部使用了instanceOnce.Do，确保只会设置一次
	control.logger.Debugf("[control] setting singleton instance...")
	SetInstance(control)

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
	if err := c.provider.Shutdown(c.ctx); err != nil {
		c.logger.Errorf("[control] shutting down provider failed: %v", err)
		return err
	}
	return nil
}

func (c *Control) GetProvider() shutdownTracerProvider {
	c.logger.Debugf("[control] get tracer provider...")
	c.providerMutex.RLock()
	defer c.providerMutex.RUnlock()
	return c.provider
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
				// exporter, err := otlptracehttp.New(c.ctx)
				// if err != nil {
				// 	return nil, fmt.Errorf("building OTLP HTTP exporter: %w", err)
				// }
				// exporters = append(exporters, exporter)
			case "grpc":
				// exporter, err := otlptracegrpc.New(c.ctx)
				// if err != nil {
				// 	return nil, fmt.Errorf("building OTLP gRPC exporter: %w", err)
				// }
				// exporters = append(exporters, exporter)
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

func (c *Control) GetServiceName() string {
	c.logger.Debugf("[control] get service name...")
	return c.config.ExporterServiceName
}

func (c *Control) newTracerProvider() error {
	c.logger.Debugf("[control] starting generate tracer provider...")
	exporters, err := c.initExporters()
	if err != nil {
		c.logger.Errorf("[control] init exporters failed: %v", err)
		return err
	}
	if len(exporters) == 0 {
		c.logger.Warnf("[control] no exporter found, using noop tracer provider...")
		c.provider = &noopShutdownTracerProvider{TracerProvider: noop.NewTracerProvider()}
		return nil
	}

	var options []sdktrace.TracerProviderOption
	for _, exporter := range exporters {
		options = append(options, sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(time.Second)))
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
	options = append(options, sdktrace.WithSampler(sdktrace.AlwaysSample())) // 或者自定义采样器
	c.providerMutex.Lock()
	c.provider = sdktrace.NewTracerProvider(options...)
	c.providerMutex.Unlock()
	return nil
}

func (c *Control) setProvider() error {
	c.logger.Debugf("[control] setting provider...")
	if err := c.newTracerProvider(); err != nil {
		c.logger.Errorf("[control] new tracer provider failed: %v", err)
		return err
	}
	c.providerMutex.RLock()
	otel.SetTracerProvider(c.provider)
	c.providerMutex.RUnlock()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	// otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())
	return nil
}

func (c *Control) init() {
	c.logger.Debugf("[control] init...")
	ctx, span := c.provider.Tracer(c.config.ExporterServiceName).Start(c.ctx, "initialize")
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
	return c.provider.Tracer(c.config.ExporterServiceName).Start(ctx, fmt.Sprintf("%s.%s", componentName, spanName), opts...)
}

// StartSpan 创建并启动一个 Span (可以根据需要添加 attributes)
func (c *Control) StartSpan(ctx context.Context, componentName string, spanName string, attributes ...attribute.KeyValue) (tCtx context.Context, span traceapi.Span) {
	return c.Span(ctx, componentName, spanName, traceapi.WithAttributes(attributes...))
}

// GetInstance 获取tracer控制器的单例实例
// @description 获取全局唯一的tracer控制器实例，确保全链路追踪的一致性
// @return *Control tracer控制器的单例实例
func GetInstance() *Control {
	instanceMu.RLock()
	result := instance
	instanceMu.RUnlock()
	return result
}

// SetInstance 设置tracer控制器的单例实例
// @description 设置全局唯一的tracer控制器实例，通常在应用初始化时调用一次
// @param control *Control tracer控制器实例
func SetInstance(control *Control) {
	instanceOnce.Do(func() {
		instanceMu.Lock()
		instance = control
		instanceMu.Unlock()
	})
}

func Span(ctx context.Context, componentName string, spanName string, opts ...traceapi.SpanStartOption) (tCtx context.Context, span traceapi.Span) {
	// 使用单例实例，如果单例实例不存在则从池中获取
	control := GetInstance()
	if control == nil {
		fmt.Printf("control is nil\n")
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
