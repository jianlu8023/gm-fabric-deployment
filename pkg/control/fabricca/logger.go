package fabricca

import (
	"path/filepath"
	"time"

	"github.com/hxx258456/fabric-sdk-go-gm/pkg/core/logging/api"
	glog "github.com/jianlu8023/go-logger/v2"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type fabricLogger struct {
	logger *zap.SugaredLogger
}

func (f *fabricLogger) Fatal(args ...interface{}) {
	f.logger.Fatal(args...)
}

func (f *fabricLogger) Fatalf(format string, args ...interface{}) {
	f.logger.Fatalf(format, args...)
}

func (f *fabricLogger) Fatalln(args ...interface{}) {
	f.logger.Fatalln(args...)
}

func (f *fabricLogger) Panic(args ...interface{}) {
	f.logger.Panic(args...)
}

func (f *fabricLogger) Panicf(format string, args ...interface{}) {
	f.logger.Panicf(format, args...)
}

func (f *fabricLogger) Panicln(args ...interface{}) {
	f.logger.Panicln(args...)
}

func (f *fabricLogger) Print(args ...interface{}) {
	f.logger.Info(args...)
}

func (f *fabricLogger) Printf(format string, args ...interface{}) {
	f.logger.Infof(format, args...)
}

func (f *fabricLogger) Println(args ...interface{}) {
	f.logger.Infoln(args...)
}

func (f *fabricLogger) Debug(args ...interface{}) {
	f.logger.Debug(args...)
}

func (f *fabricLogger) Debugf(format string, args ...interface{}) {
	f.logger.Debugf(format, args...)
}

func (f *fabricLogger) Debugln(args ...interface{}) {
	f.logger.Debugln(args...)
}

func (f *fabricLogger) Info(args ...interface{}) {
	f.logger.Info(args...)
}

func (f *fabricLogger) Infof(format string, args ...interface{}) {
	f.logger.Infof(format, args...)
}

func (f *fabricLogger) Infoln(args ...interface{}) {
	f.logger.Infoln(args...)
}

func (f *fabricLogger) Warn(args ...interface{}) {
	f.logger.Warn(args...)
}

func (f *fabricLogger) Warnf(format string, args ...interface{}) {
	f.logger.Warnf(format, args...)
}

func (f *fabricLogger) Warnln(args ...interface{}) {
	f.logger.Warnln(args...)
}

func (f *fabricLogger) Error(args ...interface{}) {
	f.logger.Error(args...)
}

func (f *fabricLogger) Errorf(format string, args ...interface{}) {
	f.logger.Errorf(format, args...)
}

func (f *fabricLogger) Errorln(args ...interface{}) {
	f.logger.Errorln(args...)
}

func (f *fabricLogger) GetLogger(module string) api.Logger {
	moduleLogger := f.logger.Named(module)
	return &fabricLogger{
		logger: moduleLogger,
	}
}

func newFabricSDKLogger(loggerConfig *config.LoggerConfig, logInConsole bool) *fabricLogger {
	fileName := "sdk.log"
	fileDir := filepath.Dir(loggerConfig.FilePath)
	fileName = filepath.Clean(filepath.Join(fileDir, fileName))
	opts := []glog.Option{
		// glog.WithModuleName("Sdk"),
		glog.WithCaller(),
		glog.WithCallerSkip(0),
		glog.WithConsoleConfig(zapcore.EncoderConfig{
			MessageKey:       "msg",
			LevelKey:         "level",
			TimeKey:          "time",
			NameKey:          "logger",
			CallerKey:        "",
			StacktraceKey:    "stacktrace",
			ConsoleSeparator: "  ",
			// FunctionKey:    "func",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    glog.CustomColorCapitalLevelEncoder,
			EncodeTime:     glog.CustomTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}),
		glog.WithFileConfig(zapcore.EncoderConfig{
			MessageKey:       "msg",
			LevelKey:         "level",
			TimeKey:          "time",
			NameKey:          "logger",
			CallerKey:        "caller",
			StacktraceKey:    "stacktrace",
			ConsoleSeparator: "  ",
			// FunctionKey:    "func",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    glog.CustomCapitalLevelEncoder,
			EncodeTime:     glog.CustomTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}),
		glog.WithFileOutPut(),
		glog.WithRotateLog(&glog.RotateLogConfig{
			FileName: fileName,
			// MaxAge:       fmt.Sprintf("%vd", loggerConfig.MaxAge),
			MaxAge:    (time.Duration(loggerConfig.MaxAge) * time.Hour * 24).String(),
			LocalTime: true,
			// RotationTime: fmt.Sprintf("%vh", loggerConfig.RotationTime),
			RotationTime: (time.Duration(loggerConfig.RotationTime) * time.Hour).String(),
		}),
		glog.WithFileLogLevel("debug"),
		glog.WithDefaultLogLevel(loggerConfig.DefaultLogLevel),
	}

	if logInConsole {
		opts = append(opts, glog.WithConsoleOutPut())
	} else {
		opts = append(opts, glog.WithOutConsoleOutPut())
	}

	logger := glog.NewSugaredLogger(opts...)
	// logger.Debug("testing logger...")

	return &fabricLogger{
		logger: logger,
	}
}
