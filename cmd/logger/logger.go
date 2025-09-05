package main

import (
	mylogger "github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
)

func main() {
	loggerConfig := &config.LoggerConfig{
		DefaultLogLevel: "debug",
		PrintFormat:     "console",
		FilePath:        "./logs/app.log",
		MaxAge:          7,
		RotationTime:    3,
		LoggerLevel: map[string]string{
			"main": "debug",
			"grpc": "debug",
		},
	}

	loggerControl := mylogger.NewLoggerControl(loggerConfig)

	logger := loggerControl.GenLogger("main")

	logger.Infof("info logger")
	logger.Debugf("debug logger")
	logger.Warnf("warn logger")
	logger.Errorf("error logger")
	logger.Fatalf("fatal logger")

}
