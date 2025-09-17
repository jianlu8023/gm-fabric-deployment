package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"_examples/cmd/rpc/grpc/mid/demo"
	"_examples/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type DemoService struct {
	demo.UnimplementedDemoServiceServer
}

func (d *DemoService) Login(ctx context.Context, de *demo.DemoRequest) (*demo.DemoResponse, error) {
	fmt.Println(ctx)
	if de.GetUsername() == "admin" && de.GetPassword() == "adminpw" {
		return &demo.DemoResponse{
			Code:    200,
			Message: "success",
		}, nil
	}
	return &demo.DemoResponse{
		Code:    403,
		Message: "failure",
	}, nil
}

func NewDemoService() *DemoService {
	return &DemoService{}
}

func main() {
	os.Setenv("OTEL_TRACES_EXPORTER", "otlp,file")
	// os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://192.168.58.110:4317")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://192.168.58.110:4318")
	os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
	os.Setenv("OTEL_EXPORTER_FILE_PATH", "./traces.log")
	// os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	// os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "grpc")

	ctx := context.Background()

	// 初始化 TracerProvider
	tp, err := tracing.NewTracerProvider(ctx, "grpc-server")
	if err != nil {
		log.Fatalf("Failed to initialize TracerProvider: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatalf("Failed to shutdown TracerProvider: %v", err)
		}
	}()

	// ctx, span := tracing.StartSpan(ctx, "grpc-client", "main") // 使用 StartSpan 创建 span
	// defer span.End()

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println(fmt.Sprintf("err:%s", err))
	}
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.MaxSendMsgSize(1024*1024*1024),
		// grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		// grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	demo.RegisterDemoServiceServer(server, NewDemoService())

	log.Println("start grpc server on 9000")
	reflection.Register(server)
	err = server.Serve(listen)
	if err != nil {
		fmt.Println(fmt.Sprintf("err:%s", err))
	}
}
