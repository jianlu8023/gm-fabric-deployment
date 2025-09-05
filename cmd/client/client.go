package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mylogger "github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/internal/middleware/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/internal/proto/message"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/pidfile"
)

var (
	version string
)

func main() {

	fmt.Printf("start client version %s\n", version)

	if err := pidfile.CreateOrUpdatePIDFile("client.pid"); err != nil {
		fmt.Printf("generate pid file failed: %v\n", err)
		return
	}
	defer func() {
		pidfile.ReleasePID()
	}()
	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	loggerControl := mylogger.NewLoggerControl(configControl.GetConfig().LoggerConfig)
	mainLogger := loggerControl.GenLogger("main")

	mainLogger.Infof("starting grpc server...")
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
	control, err := grpc.NewGrpcControl(grpcConfig, loggerControl)
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
				mainLogger.Errorf("send ping message err: %v", err)
			} else {
				mainLogger.Debugf("send ping message success %v", response)
			}
		}
	}()

	<-quit
	mainLogger.Infof("received shutdown signal...")

}
