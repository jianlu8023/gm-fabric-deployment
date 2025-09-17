package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"_examples/cmd/rpc/grpc/mid/demo"
	"_examples/cmd/rpc/grpc/otel/myjaeger"
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
	shutdown, err := myjaeger.InitJaeger("otel-grpc-demo_server")

	if err != nil {
		log.Fatalf("Failed to initialize Jaeger: %v", err)
	}
	defer func() {
		if err = shutdown(context.Background()); err != nil {
			log.Fatalf("Failed to shutdown TracerProvider: %v", err)
		}
	}()

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
