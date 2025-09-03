package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	_logMap = make(map[string]*zap.SugaredLogger)

	loggerMutex sync.RWMutex
)

type Logger struct {
}

func GenLogger(moduleName string) *zap.SugaredLogger {
	loggerMutex.Lock()
	existLogger, ok := _logMap[moduleName]
	if ok {
		return existLogger
	}
	return nil
}
