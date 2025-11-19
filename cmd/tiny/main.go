package main

import (
	"log"
	"os"

	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"github.com/jianlu8023/golang-example/cmd/tiny/command"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/container"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/container/verify"
	"github.com/jianlu8023/golang-example/cmd/tiny/server"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

func main() {
	var err error

	loggerControl := logger.NewLoggerControl(&config.LoggerConfig{
		DefaultLogLevel: "info",
		PrintFormat:     "console",
	})
	loggerControl.StartUp(func(err error) {
		if err != nil {
			os.Exit(1)
		}
	})
	defer func() {
		_ = loggerControl.Shutdown()

		if err != nil {
			log.Printf(colour.Red(err.Error()))
		}
	}()

	mainLogger := loggerControl.GenLogger("")

	if err = conf.LoadConfig(mainLogger); err != nil {
		mainLogger.Errorf("加载配置文件失败: %v", err)
		return
	}

	if err = command.CLI().Run(os.Args); err != nil {
		mainLogger.Errorf("解析启用命令失败: %v", err)
		return
	}

	if err = container.NewHandleChain().
		AddToHead(verify.NewPortVerifier(conf.Config.Port)).
		AddToHead(verify.NewPathVerifier(conf.Config.RootPath)).
		Iterator(); err != nil {
		mainLogger.Errorf("链式检查必要配置失败: %v", err)
		return
	}
	server.RunCore(mainLogger)
}
