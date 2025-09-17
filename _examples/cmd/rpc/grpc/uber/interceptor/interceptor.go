package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// // MetadataTextMap adapts metadata.MD to OpenTelemetry TextMapCarrier interface.
// type MetadataTextMap metadata.MD
//
// // Get retrieves a single value for a given key.
// func (m MetadataTextMap) Get(key string) string {
// 	key = strings.ToLower(key)
// 	values := metadata.MD(m).Get(key)
// 	if len(values) > 0 {
// 		return values[0]
// 	}
// 	return ""
// }
//
// // Set stores a key-value pair.
// func (m MetadataTextMap) Set(key string, value string) {
// 	key = strings.ToLower(key)
// 	metadata.MD(m).Set(key, value)
// }
//
// // Keys lists the keys stored in this carrier.
// func (m MetadataTextMap) Keys() []string {
// 	keys := make([]string, 0, len(metadata.MD(m)))
// 	for k := range metadata.MD(m) {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }
//
// // ServerInterceptor returns a new unary server interceptor for OpenTelemetry tracing.
// func ServerInterceptor() grpc.UnaryServerInterceptor {
// 	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 		tracer := otel.Tracer("grpc-server") // 获取 OpenTelemetry Tracer
//
// 		// 从 gRPC metadata 中提取 context
// 		md, ok := metadata.FromIncomingContext(ctx)
// 		if !ok {
// 			md = metadata.New(nil)
// 		}
//
// 		// 使用 OpenTelemetry propagator 从 metadata 中提取 span context
// 		carrier := MetadataTextMap(md) // 使用自定义的 MetadataTextMap
// 		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
//
// 		// 创建 span
// 		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
// 		defer span.End()
//
// 		// 将一些有用的属性添加到 span
// 		span.SetAttributes(attribute.String("grpc.method", info.FullMethod))
//
// 		// 调用 handler
// 		resp, err := handler(ctx, req)
//
// 		// 处理错误 (可选)
// 		if err != nil {
// 			span.RecordError(err)
// 			span.SetAttributes(attribute.String("error", err.Error())) // 设置 attribute
// 			fmt.Printf("error happen : %v\n", err)
// 		}
//
// 		return resp, err
// 	}
// }
//
// // ClientInterceptor grpc client interceptor using OpenTelemetry
// func ClientInterceptor() grpc.UnaryClientInterceptor {
// 	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
// 		tracer := otel.Tracer("grpc-client") // 获取 OpenTelemetry Tracer
//
// 		// 创建 span
// 		ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
// 		defer span.End()
//
// 		// 将一些有用的属性添加到 span
// 		span.SetAttributes(attribute.String("grpc.method", method))
//
// 		// 从 context 中提取 metadata
// 		md, ok := metadata.FromOutgoingContext(ctx)
// 		if !ok {
// 			md = metadata.New(nil)
// 		} else {
// 			md = md.Copy()
// 		}
//
// 		// 使用 OpenTelemetry propagator 将 span context 注入到 metadata 中
// 		carrier := MetadataTextMap(md) // 使用自定义的 MetadataTextMap
// 		otel.GetTextMapPropagator().Inject(ctx, carrier)
//
// 		newCtx := metadata.NewOutgoingContext(ctx, md)
// 		err := invoker(newCtx, method, req, reply, cc, opts...)
// 		if err != nil {
// 			span.RecordError(err)
// 			span.SetAttributes(attribute.String("error", err.Error())) // 设置 attribute
// 		}
// 		return err
// 	}
// }

// MDReaderWriter 自定义的metadata
// 为了做载体(carrier),必须要实现 `opentracing.TextMapWriter` `opentracing.TextMapReader` 这两个接口。
type MDReaderWriter struct {
	metadata.MD
}

// ForeachKey 实现opentracing.TextMapReader
func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// Set 实现 opentracing.TextMapWriter 接口
func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

func ServerInterceptor(tracer opentracing.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		// 服务端拦截器则是在MD中把 span提取出来
		spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
		if err != nil && !errors.Is(err, opentracing.ErrSpanContextNotFound) {
			fmt.Print("extract from metadata error: ", err)
		} else {
			span := tracer.StartSpan(
				info.FullMethod,
				ext.RPCServerOption(spanContext),
				opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
				ext.SpanKindRPCServer,
			)
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)
		}
		return handler(ctx, req)
	}
}

// ClientInterceptor grpc client
func ClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span, _ := opentracing.StartSpanFromContext(ctx,
			"call gRPC",
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient)
		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		// 在客户端拦截器中把 span 注入进去
		err := tracer.Inject(span.Context(), opentracing.TextMap, MDReaderWriter{md})
		if err != nil {
			span.LogFields(log.String("inject-error", err.Error()))
		}

		newCtx := metadata.NewOutgoingContext(ctx, md)
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(log.String("call-error", err.Error()))
		}
		return err
	}
}
