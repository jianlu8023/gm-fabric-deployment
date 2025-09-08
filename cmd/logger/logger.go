package main

import (
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/logger"
)

func main() {
	loggerConfig := &config.LoggerConfig{
		DefaultLogLevel: "debug",
		StackLogLevel:   "error",
		PrintFormat:     "console",
		FilePath:        "./logs/app.log",
		MaxAge:          7,
		RotationTime:    1,
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
