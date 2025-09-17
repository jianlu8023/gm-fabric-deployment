package main

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
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

	loggerControl := logger.NewLoggerControl(loggerConfig)

	logger := loggerControl.GenLogger("main")

	logger.Infof("info logger")
	logger.Debugf("debug logger")
	logger.Warnf("warn logger")
	logger.Errorf("error logger")
	logger.Fatalf("fatal logger")

}
