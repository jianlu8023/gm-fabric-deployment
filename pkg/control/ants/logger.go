package ants

import (
	"go.uber.org/zap"
)

type antsCustomLogger struct {
	logger *zap.SugaredLogger
}

func (a *antsCustomLogger) Printf(format string, args ...any) {
	a.logger.Infof(format, args...)
}

func newAntsLogger(logger *zap.SugaredLogger) *antsCustomLogger {
	return &antsCustomLogger{logger: logger}
}
