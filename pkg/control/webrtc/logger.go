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

// webrtcCustomLogger 适配 pion logging.LoggerFactory 接口的 zap 日志包装器
//
// @description 适配 pion logging.LoggerFactory 接口的 zap 日志包装器，
// 将 pion 内部日志统一导向项目的 zap 日志体系，便于集中管理与检索
// @struct
type webrtcCustomLogger struct {
	logger *zap.SugaredLogger
}

// Trace 记录 trace 级别日志
//
// @description 记录 trace 级别日志，由于 zap 原生无 trace 级别，统一降级为 debug 级别输出
// @param msg string 日志消息
func (w *webrtcCustomLogger) Trace(msg string) {
	w.logger.Debug(msg)
}

// Tracef 记录格式化 trace 级别日志
//
// @description 记录格式化 trace 级别日志，由于 zap 原生无 trace 级别，统一降级为 debug 级别输出
// @param format string 格式化字符串
// @param args ...interface{} 格式化参数
func (w *webrtcCustomLogger) Tracef(format string, args ...interface{}) {
	w.logger.Debugf(format, args...)
}

// Debug 记录 debug 级别日志
//
// @description 记录 debug 级别日志
// @param msg string 日志消息
func (w *webrtcCustomLogger) Debug(msg string) {
	w.logger.Debug(msg)
}

// Debugf 记录格式化 debug 级别日志
//
// @description 记录格式化 debug 级别日志
// @param format string 格式化字符串
// @param args ...interface{} 格式化参数
func (w *webrtcCustomLogger) Debugf(format string, args ...interface{}) {
	w.logger.Debugf(format, args...)
}

// Info 记录 info 级别日志
//
// @description 记录 info 级别日志
// @param msg string 日志消息
func (w *webrtcCustomLogger) Info(msg string) {
	w.logger.Info(msg)
}

// Infof 记录格式化 info 级别日志
//
// @description 记录格式化 info 级别日志
// @param format string 格式化字符串
// @param args ...interface{} 格式化参数
func (w *webrtcCustomLogger) Infof(format string, args ...interface{}) {
	w.logger.Infof(format, args...)
}

// Warn 记录 warn 级别日志
//
// @description 记录 warn 级别日志
// @param msg string 日志消息
func (w *webrtcCustomLogger) Warn(msg string) {
	w.logger.Warn(msg)
}

// Warnf 记录格式化 warn 级别日志
//
// @description 记录格式化 warn 级别日志
// @param format string 格式化字符串
// @param args ...interface{} 格式化参数
func (w *webrtcCustomLogger) Warnf(format string, args ...interface{}) {
	w.logger.Warnf(format, args...)
}

// Error 记录 error 级别日志
//
// @description 记录 error 级别日志
// @param msg string 日志消息
func (w *webrtcCustomLogger) Error(msg string) {
	w.logger.Error(msg)
}

// Errorf 记录格式化 error 级别日志
//
// @description 记录格式化 error 级别日志
// @param format string 格式化字符串
// @param args ...interface{} 格式化参数
func (w *webrtcCustomLogger) Errorf(format string, args ...interface{}) {
	w.logger.Errorf(format, args...)
}

// NewLogger 为指定 scope 创建子日志记录器
//
// @description 为指定 scope 创建子日志记录器，便于按模块区分日志来源
// @param scope string 日志作用域
// @return logging.LeveledLogger 创建的子日志记录器
func (w *webrtcCustomLogger) NewLogger(scope string) logging.LeveledLogger {
	named := w.logger.Named(scope)
	return &webrtcCustomLogger{
		logger: named,
	}
}

// newWebRTCLogger 创建 WebRTC SDK 专用日志记录器
//
// @description 创建 WebRTC SDK 专用日志记录器，基于 go-logger 配置独立的文件输出与日志轮转。
// 注意：RotateLogConfig 的 MaxAge/RotationTime 字段为 string 类型，
// 因此这里将 time.Duration 转换为字符串形式（如 "24h0m0s"）传入，
// 由 go-logger 内部解析为 time.Duration
// @param loggerConfig *config.LoggerConfig 日志配置信息
// @param logInConsole bool 是否在控制台输出日志
// @return *webrtcCustomLogger 创建的 WebRTC 日志记录器
func newWebRTCLogger(loggerConfig *config.LoggerConfig, logInConsole bool) *webrtcCustomLogger {
	fileName := "webrtc.log"
	fileDir := filepath.Dir(loggerConfig.FilePath)
	fileName = filepath.Clean(filepath.Join(fileDir, fileName))
	opts := []glog.Option{
		glog.WithModuleName("WebrtcSDK"),
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
		glog.WithConsoleLogLevel("debug"),
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
