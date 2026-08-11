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
		FilePath:        "./logs/logger.log",
		MaxAge:          7,
		RotationTime:    1,
		LoggerLevel: map[string]string{
			"main": "debug",
			"grpc": "debug",
		},
	}

	loggerControl := logger.NewLoggerControl(loggerConfig)

	sugaredLogger := loggerControl.GenLogger("main")

	sugaredLogger.Infof("info sugaredLogger")
	sugaredLogger.Debugf("debug sugaredLogger")
	sugaredLogger.Warnf("warn sugaredLogger")
	sugaredLogger.Errorf("error sugaredLogger")
	sugaredLogger.Fatalf("fatal sugaredLogger")

}
