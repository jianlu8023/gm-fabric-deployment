package webrtc

import (
	"path/filepath"
	"time"

	glog "github.com/jianlu8023/go-logger/v2"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/pion/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type webrtcCustomLogger struct {
	logger *zap.SugaredLogger
}

func (w *webrtcCustomLogger) Trace(msg string) {
	w.logger.Debug(msg)
}

func (w *webrtcCustomLogger) Tracef(format string, args ...interface{}) {
	w.logger.Debugf(format, args...)
}

func (w *webrtcCustomLogger) Debug(msg string) {
	w.logger.Debug(msg)
}

func (w *webrtcCustomLogger) Debugf(format string, args ...interface{}) {
	w.logger.Debugf(format, args...)
}

func (w *webrtcCustomLogger) Info(msg string) {
	w.logger.Info(msg)
}

func (w *webrtcCustomLogger) Infof(format string, args ...interface{}) {
	w.logger.Infof(format, args...)
}

func (w *webrtcCustomLogger) Warn(msg string) {
	w.logger.Warn(msg)
}

func (w *webrtcCustomLogger) Warnf(format string, args ...interface{}) {
	w.logger.Warnf(format, args...)
}

func (w *webrtcCustomLogger) Error(msg string) {
	w.logger.Error(msg)
}

func (w *webrtcCustomLogger) Errorf(format string, args ...interface{}) {
	w.logger.Errorf(format, args...)
}

func (w *webrtcCustomLogger) NewLogger(scope string) logging.LeveledLogger {
	named := w.logger.Named(scope)
	return &webrtcCustomLogger{
		logger: named,
	}
}

func newWebRTCLogger(loggerConfig *config.LoggerConfig, logInConsole bool) *webrtcCustomLogger {
	fileName := "webrtc.log"
	fileDir := filepath.Dir(loggerConfig.FilePath)
	fileName = filepath.Clean(filepath.Join(fileDir, fileName))
	opts := []glog.Option{
		glog.WithModuleName("webrtc.sdk"),
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
		glog.WithStackLogLevel(loggerConfig.StackLogLevel),
	}

	if logInConsole {
		opts = append(opts, glog.WithConsoleOutPut())
	} else {
		opts = append(opts, glog.WithOutConsoleOutPut())
	}

	logger := glog.NewSugaredLogger(opts...)
	// logger.Debug("testing logger...")

	return &webrtcCustomLogger{
		logger: logger,
	}
}
