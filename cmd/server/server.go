package main

import (
	"errors"
	"fmt"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/internal/middleware/grpc"
	myhttp "github.com/jianlu8023/gm-fabric-deployment/internal/middleware/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/pidfile"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	version string
)

func main() {
	fmt.Printf("start server version %s\n", version)
	if err := pidfile.CreateOrUpdatePIDFile("server.pid"); err != nil {
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

	loggerControl := mylogger.NewLoggerControl(configControl.Config.LoggerConfig)
	mainLogger := loggerControl.GenLogger("main")

	mainLogger.Infof("starting grpc server...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	control, err := grpc.NewGrpcControl(configControl.Config.GrpcConfig, loggerControl)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer control.Shutdown()

	mainLogger.Infof("starting http server...")

	httpControl := myhttp.NewServerControl(configControl.Config.HttpConfig, loggerControl)

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
