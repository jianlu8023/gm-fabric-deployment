package main

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/internal/middleware/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/internal/proto/message"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	fmt.Println("client")

	grpcConfig := &config.GrpcConfig{
		Server: &config.GrpcServerConfig{
			Host:           "127.0.0.1:65533",
			MaxRecvMsgSize: 5242880,
			MaxSendMsgSize: 5242880,
			ChunkSize:      4194304,
			TlsEnabled:     true,
			TlsCertFile:    "./certs/gserver.crt",
			TlsKeyFile:     "./certs/gserver.key",
			TlsRCACertFile: "./certs/root-ca.crt",
		},
		Client: &config.GrpcClientConfig{
			Host:               "127.0.0.1:65534",
			MaxCallRecvMsgSize: 5242880,
			MaxCallSendMsgSize: 5242880,
			ChunkSize:          4194304,
			CallTimeout:        99999,
			TlsEnabled:         true,
			TlsCertFile:        "./certs/gclient.crt",
			TlsKeyFile:         "./certs/gclient.key",
			TlsRCACertFile:     "./certs/root-ca.crt",
		},
	}
	control, err := grpc.NewGrpcControl(grpcConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer control.Shutdown()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			response, err := control.Call(&message.BaseRequest{
				MessageType: "base/ping",
			})
			if err != nil {
				fmt.Printf("ping error %v\n", err)
			} else {

				fmt.Printf("ping result %v\n", response)
			}
		}
	}()

	<-quit
	fmt.Println("server stop")

}
