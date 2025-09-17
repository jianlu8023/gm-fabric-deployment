package myjaeger

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	// go.opentelemetry.io/otel/exporters/jaeger
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Deprecated module go. opentelemetry. io/ otel/ exporters/ jaeger.
// Deprecation notice: This module is no longer supported.
// OpenTelemetry dropped support for Jaeger exporter in July 2023.
// Jaeger officially accepts and recommends using OTLP.
// Use [go. opentelemetry. io/ otel/ exporters/ otlp/ otlptrace/ otlptracehttp] or
// [go. opentelemetry. io/ otel/ exporters/ otlp/ otlptrace/ otlptracegrpc] instead.

// func InitJaeger(serviceName string) (func(ctx context.Context) error, error) {
//
// 	// Jaeger endpoint
// 	jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
// 	if jaegerEndpoint == "" {
// 		jaegerEndpoint = "http://localhost:14268/api/traces" // 默认 Jaeger 地址
// 		log.Println("JAEGER_ENDPOINT not set, using default:", jaegerEndpoint)
// 	}
//
// 	// 创建 Jaeger exporter
// 	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
// 	if err != nil {
// 		return nil, fmt.Errorf("creating Jaeger exporter: %w", err)
// 	}
//
// 	// 创建 resource
// 	res, err := resource.New(
// 		context.Background(),
// 		resource.WithAttributes(
// 			semconv.ServiceName(serviceName),
// 		),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("creating resource: %w", err)
// 	}
//
// 	// 创建 TracerProvider
// 	tp := sdktrace.NewTracerProvider(
// 		sdktrace.WithBatcher(exp),
// 		sdktrace.WithResource(res),
// 	)
//
// 	otel.SetTracerProvider(tp)
// 	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
//
// 	// 返回一个关闭函数
// 	return func(ctx context.Context) error {
// 		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
// 		defer cancel()
// 		if err := tp.Shutdown(ctx); err != nil {
// 			return fmt.Errorf("shutting down tracer provider: %w", err)
// 		}
// 		return nil
// 	}, nil
// }

func InitJaeger(serviceName string) (func(ctx context.Context) error, error) {
	// OTLP endpoint
	otlpEndpoint := os.Getenv("OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "localhost:4318" // 默认 OTLP 地址
		log.Println("OTLP_ENDPOINT not set, using default:", otlpEndpoint)
	}

	// 创建 OTLP HTTP Exporter
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(), // 生产环境移除，使用 TLS
	)

	exp, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP HTTP exporter: %w", err)
	}

	// 创建 resource
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName), // 使用 semconv.ServiceNameKey
		),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// 返回一个关闭函数
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutting down tracer provider: %w", err)
		}
		return nil
	}, nil
}
