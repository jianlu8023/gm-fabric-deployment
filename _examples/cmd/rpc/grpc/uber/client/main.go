package main

import (
	"context"
	"fmt"

	"_examples/cmd/rpc/grpc/mid/demo"
	"_examples/cmd/rpc/grpc/uber/interceptor"
	"_examples/cmd/rpc/grpc/uber/tracer"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	newTracer, closer := tracer.NewTracer("gRPC-client")
	defer closer.Close()

	conn, err := grpc.NewClient(
		"127.0.0.1:9000",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1048576)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1048576)),
		grpc.WithUnaryInterceptor(
			grpcMiddleware.ChainUnaryClient(
				interceptor.ClientInterceptor(newTracer),
			),
		),
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
