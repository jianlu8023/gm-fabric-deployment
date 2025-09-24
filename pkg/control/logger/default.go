package logger

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.LoggerConfig {
	return &config.LoggerConfig{
		DefaultLogLevel: "debug",
		StackLogLevel:   "error",
		PrintFormat:     "console",
		FilePath:        "logs/app.log",
		MaxAge:          7,
		RotationTime:    1,
		LoggerLevel:     map[string]string{},
	}
}
