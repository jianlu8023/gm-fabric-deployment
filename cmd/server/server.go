package main

import (
	"errors"
	"fmt"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/internal/middleware/grpc"
	myhttp "github.com/jianlu8023/gm-fabric-deployment/internal/middleware/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("server")

	loggerConfig := &config.LoggerConfig{
		DefaultLogLevel: "debug",
		PrintFormat:     "console",
		FilePath:        "./logs/app.log",
		MaxAge:          7,
		RotationTime:    3,
		LoggerLevel: map[string]string{
			"main": "info",
			"grpc": "debug",
		},
	}

	loggerControl := mylogger.NewLoggerControl(loggerConfig)
	mainLogger := loggerControl.GenLogger("main")

	mainLogger.Infof("starting grpc server...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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
	control, err := grpc.NewGrpcControl(grpcConfig, loggerControl)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer control.Shutdown()

	mainLogger.Infof("starting http server...")
	httpConfig := &config.HttpServerConfig{
		Address:     ":8080",
		RunMode:     "release",
		ContextPath: "example",
		TlsEnabled:  true,
		TlsCertFile: "./certs/hserver.crt",
		TlsKeyFile:  "./certs/hserver.key",
	}

	httpControl := myhttp.NewServerControl(httpConfig, loggerControl)

	httpControl.StartUp(func(err error) {
		if !errors.Is(err, http.ErrServerClosed) {
			mainLogger.Errorf("start http server err: %v", err)
		}
		quit <- os.Interrupt
	})
	defer httpControl.Shutdown()

	<-quit
	mainLogger.Infof("received shutdown signal...")
}
