package main

import (
	"context"
	"fmt"
	"log"

	"_examples/cmd/rpc/grpc/mid/demo"
	"_examples/cmd/rpc/grpc/otel/myjaeger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	// 初始化 Jaeger
	shutdown, err := myjaeger.InitJaeger("otel-grpc-demo_client") // 替换为你的服务名称
	if err != nil {
		log.Fatalf("Failed to initialize Jaeger: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
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
