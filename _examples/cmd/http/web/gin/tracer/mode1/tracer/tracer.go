package tracer

import (
	"fmt"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
)

const (
	CollectorEndpointHTTP = "http://192.168.58.110:14268/api/traces"
	CollectorEndpointGRPC = "192.168.58.110:14250"
	LocalAgentHostPort2   = "localhost:6831"
)

// NewTracer 使用 opentracing 统一标准
func NewTracer(service string) (opentracing.Tracer, io.Closer) {
	// config := parseConfig()
	// return newTracer(service, CollectorEndpointHTTP)
	return newTracerUber(service, CollectorEndpointHTTP)
}

// // otelCloser 包装 OpenTelemetry TracerProvider 的 Shutdown 方法
// type otelCloser struct {
// 	tp *tracesdk.TracerProvider
// }
//
// // Close 关闭 OpenTelemetry TracerProvider
// func (c *otelCloser) Close() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	if err := c.tp.Shutdown(ctx); err != nil {
// 		log.Printf("Error shutting down tracer provider: %v", err)
// 		return err
// 	}
// 	return nil
// }
//
// // newTracer
// func newTracer(service, collectorEndpoint string) (opentracing.Tracer, io.Closer) {
// 	tp, err := tracerProvider(service, collectorEndpoint)
// 	if err != nil {
// 		log.Fatalf("Failed to create tracer provider: %v", err)
// 		return nil, nil // 返回 nil Tracer 和 Closer 以避免 panic
// 	}
//
// 	// 设置全局的 OpenTelemetry TracerProvider
// 	otel.SetTracerProvider(tp)
//
// 	// 创建一个 Closer，用于在程序退出时关闭 TracerProvider
// 	closer := &otelCloser{tp: tp}
//
// 	// 返回 OpenTelemetry Tracer，但实现了 opentracing.Tracer 接口
// 	// 这样可以保持与现有 opentracing 代码的兼容性
// 	return otel.Tracer(service).(opentracing.Tracer), closer
// }

// newTracerUber
func newTracerUber(service, collectorEndpoint string) (opentracing.Tracer, io.Closer) {
	// 参数详解 https://www.jaegertracing.io/docs/1.20/sampling/
	cfg := jaegerConfig.Configuration{
		ServiceName: service,
		// 采样配置
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           true,
			CollectorEndpoint:  collectorEndpoint, // 将span发往jaeger-collector的服务地址
			LocalAgentHostPort: LocalAgentHostPort2,
		},
	}
	// 不传递 logger 就不会打印日志
	tracer, closer, err := cfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
	// tracer, closer, err := cfg.NewTracer()
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

// // tracerProvider 返回一个 OpenTelemetry TracerProvider
// func tracerProvider(serviceName string, agentEndpoint string) (*tracesdk.TracerProvider, error) {
// 	// 创建 Jaeger exporter (使用 gRPC Agent)
// 	exp, err := jaeger.New(
// 		jaeger.WithCollectorEndpoint(
// 			jaeger.WithEndpoint(agentEndpoint),
// 		),
// 		// jaeger.WithAgentEndpoint(
// 		// 	jaeger.WithAgentPort("6831"),
// 		// 	jaeger.WithAgentHost("192.168.58.110"),
// 		// ),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating Jaeger exporter: %w", err)
// 	}
//
// 	// 创建资源
// 	res, err := resource.New(context.Background(),
// 		resource.WithAttributes(
// 			semconv.ServiceName(serviceName),
// 			// attribute.String("environment", environment),
// 			// attribute.Int64("ID", id),
// 		),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating resource: %w", err)
// 	}
//
// 	// 创建 TracerProvider
// 	tp := tracesdk.NewTracerProvider(
// 		tracesdk.WithBatcher(exp), // 在生产环境中始终使用 Batcher
// 		tracesdk.WithResource(res),
// 	)
//
// 	return tp, nil
// }

//
// type JaegerConfig struct {
// 	CollectorEndpoint string `json:"collectorEndpoint"`
// }
//
// var jc JaegerConfig
//
// // parseConfig 从 viper 中解析配置信息
// func parseConfig() JaegerConfig {
// 	if jc.CollectorEndpoint != "" {
// 		return jc
// 	}
// 	if err := viper.UnmarshalKey("Jaeger", &jc); err != nil {
// 		panic(err)
// 	}
// 	if jc.CollectorEndpoint == "" {
// 		panic("jaeger conf nil")
// 	}
// 	return jc
// }
