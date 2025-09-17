package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"_examples/cmd/rpc/grpc/mid/demo"
	"_examples/cmd/rpc/grpc/uber/interceptor"
	"_examples/cmd/rpc/grpc/uber/tracer"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
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

	listen, err := net.Listen("tcp", ":9000")
	if err != nil {
		fmt.Println(fmt.Sprintf("err:%s", err))
	}
	newTracer, closer := tracer.NewTracer("gRPC-server")
	defer closer.Close()

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(1024*1024*1024),
		grpc.MaxSendMsgSize(1024*1024*1024),
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				interceptor.ServerInterceptor(newTracer),
			),
		),
	)

	// server := grpc.NewServer(
	// 	grpc.MaxRecvMsgSize(1024*1024*1024),
	// 	grpc.MaxSendMsgSize(1024*1024*1024),
	// 	grpc.UnaryInterceptor(
	// 		grpcMiddleware.ChainUnaryServer(
	// 			interceptor.ServerInterceptor(),
	// 		),
	// 	),
	// )

	demo.RegisterDemoServiceServer(server, NewDemoService())

	log.Println("start grpc server on 9000")
	reflection.Register(server)
	err = server.Serve(listen)
	if err != nil {
		fmt.Println(fmt.Sprintf("err:%s", err))
	}
}
