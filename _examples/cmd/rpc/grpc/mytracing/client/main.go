package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"_examples/cmd/rpc/grpc/mid/demo"
	"_examples/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	os.Setenv("OTEL_TRACES_EXPORTER", "otlp,file")
	// os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://192.168.58.110:4317")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://192.168.58.110:4318")
	os.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
	os.Setenv("OTEL_EXPORTER_FILE_PATH", "./traces.log")
	// os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	// os.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "grpc")
	// 默认是http模式 所以设置端口应该为4318 而不是4317
	// 指定 grpc 则使用4317

	// 初始化 TracerProvider
	tp, err := tracing.NewTracerProvider(context.Background(), "grpc-client")
	if err != nil {
		log.Fatalf("Failed to initialize TracerProvider: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown TracerProvider: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"127.0.0.1:9000",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1048576)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1048576)),
		// grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		// grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)

	if err != nil {
		fmt.Println(fmt.Sprintf("err:%s", err))
	}
	defer conn.Close()
	client := demo.NewDemoServiceClient(conn)
	ctx := context.WithValue(context.Background(), "key", "client")

	response, err := client.Login(ctx, &demo.DemoRequest{
		Username: "admin",
		Password: "adminpw",
	})
	if err != nil {
		fmt.Println(fmt.Sprintf("err:%s", err))
		return
	}

	fmt.Println(response)

}
