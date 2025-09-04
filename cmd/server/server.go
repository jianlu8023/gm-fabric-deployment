package main

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/internal/middleware/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("server")
	grpcConfig := &config.GrpcConfig{
		Server: &config.GrpcServerConfig{
			Host:           "127.0.0.1:65534",
			MaxRecvMsgSize: 5242880,
			MaxSendMsgSize: 5242880,
			ChunkSize:      4194304,
			TlsEnabled:     true,
			TlsCertFile:    "./certs/gserver.crt",
			TlsKeyFile:     "./certs/gserver.key",
			TlsRCACertFile: "./certs/root-ca.crt",
		},
		Client: &config.GrpcClientConfig{
			Host:               "127.0.0.1:65533",
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

	<-quit
	fmt.Println("server stop")

}
