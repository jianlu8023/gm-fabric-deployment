package tracer

import (
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	glog "github.com/jianlu8023/go-logger/v2"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"go.uber.org/zap/zapcore"
)

func newTracerCustomLogger(loggerConfig *config.LoggerConfig, logInConsole bool) logr.Logger {
	fileName := "tracer.log"
	fileDir := filepath.Dir(loggerConfig.FilePath)
	fileName = filepath.Clean(filepath.Join(fileDir, fileName))
	opts := []glog.Option{
		glog.WithModuleName("Tracer"),
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
			CallerKey:        "",
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
		glog.WithDefaultLogLevel("info"),
		glog.WithConsoleLogLevel("info"),
	}

	if logInConsole {
		opts = append(opts, glog.WithConsoleOutPut())
	} else {
		opts = append(opts, glog.WithOutConsoleOutPut())
	}
	logrLogger := zapr.NewLoggerWithOptions(glog.NewLogger(opts...))
	// logrLogger.Info("info logger")
	// logger.Debug("testing logger...")
	return logrLogger
}
