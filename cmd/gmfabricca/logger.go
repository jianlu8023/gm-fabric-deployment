package main

import (
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/core/logging/api"
	"go.uber.org/zap"
)

type fabricLogger struct {
	logger *zap.SugaredLogger
}

func (f fabricLogger) Fatal(v ...interface{}) {
	f.logger.Fatal(v...)
}

func (f fabricLogger) Fatalf(format string, v ...interface{}) {
	f.logger.Fatalf(format, v...)
}

func (f fabricLogger) Fatalln(v ...interface{}) {
	f.logger.Fatalln(v...)
}

func (f fabricLogger) Panic(v ...interface{}) {
	f.logger.Panic(v...)
}

func (f fabricLogger) Panicf(format string, v ...interface{}) {
	f.logger.Panicf(format, v...)
}

func (f fabricLogger) Panicln(v ...interface{}) {
	f.logger.Panicln(v...)
}

func (f fabricLogger) Print(v ...interface{}) {
	f.logger.Info(v...)
}

func (f fabricLogger) Printf(format string, v ...interface{}) {
	f.logger.Infof(format, v...)
}

func (f fabricLogger) Println(v ...interface{}) {
	f.logger.Infoln(v...)
}

func (f fabricLogger) Debug(args ...interface{}) {
	f.logger.Debug(args...)
}

func (f fabricLogger) Debugf(format string, args ...interface{}) {
	f.logger.Debugf(format, args...)
}

func (f fabricLogger) Debugln(args ...interface{}) {
	f.logger.Debugln(args...)
}

func (f fabricLogger) Info(args ...interface{}) {
	f.logger.Info(args...)
}

func (f fabricLogger) Infof(format string, args ...interface{}) {
	f.logger.Infof(format, args...)
}

func (f fabricLogger) Infoln(args ...interface{}) {
	f.logger.Infoln(args...)
}

func (f fabricLogger) Warn(args ...interface{}) {
	f.logger.Warn(args...)
}

func (f fabricLogger) Warnf(format string, args ...interface{}) {
	f.logger.Warnf(format, args...)
}

func (f fabricLogger) Warnln(args ...interface{}) {
	f.logger.Warnln(args...)
}

func (f fabricLogger) Error(args ...interface{}) {
	f.logger.Error(args...)
}

func (f fabricLogger) Errorf(format string, args ...interface{}) {
	f.logger.Errorf(format, args...)
}

func (f fabricLogger) Errorln(args ...interface{}) {
	f.logger.Errorln(args...)
}

func (f fabricLogger) GetLogger(module string) api.Logger {
	named := f.logger.Named(module)
	return newFabricLogger(named)
}

func newFabricLogger(logger *zap.SugaredLogger) *fabricLogger {
	return &fabricLogger{logger: logger}
}
