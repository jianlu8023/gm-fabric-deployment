package logger

import (
	"fmt"
	"strings"
	"sync"
	"time"

	glog "github.com/jianlu8023/go-logger/v2"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Control struct {
	loggerConfig *config.LoggerConfig
	_logMap      map[string]*zap.SugaredLogger
	_loggerLevel map[string]string
	loggerMutex  sync.RWMutex
	once         sync.Once
}

func (c *Control) GetConfig() *config.LoggerConfig {
	return c.loggerConfig
}

func NewLoggerControl(loggerConfig *config.LoggerConfig) *Control {
	if loggerConfig == nil {
		loggerConfig = getDefaultConfig()
	}
	loggerLevel := make(map[string]string)
	for logger, level := range loggerConfig.LoggerLevel {
		loggerLevel[strings.ToLower(logger)] = level
	}

	return &Control{
		loggerConfig: loggerConfig,
		_logMap:      make(map[string]*zap.SugaredLogger),
		_loggerLevel: loggerLevel,
	}
}

func (c *Control) GenLogger(moduleName string) *zap.SugaredLogger {
	c.loggerMutex.Lock()
	defer c.loggerMutex.Unlock()

	if c.loggerConfig == nil {
		return nil
	}

	existLogger, ok := c._logMap[moduleName]
	if ok {
		return existLogger
	}

	opts := []glog.Option{
		glog.WithModuleName(moduleName),
		glog.WithCaller(),
		glog.WithCallerSkip(0),
		glog.WithConsoleConfig(zapcore.EncoderConfig{
			MessageKey:       "msg",
			LevelKey:         "level",
			TimeKey:          "time",
			NameKey:          "logger",
			CallerKey:        "caller",
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
			FileName: c.loggerConfig.FilePath,
			// MaxAge:       fmt.Sprintf("%vd", c.loggerConfig.MaxAge),
			MaxAge:    (time.Duration(c.loggerConfig.MaxAge) * time.Hour * 24).String(),
			LocalTime: true,
			// RotationTime: fmt.Sprintf("%vh", c.loggerConfig.RotationTime),
			RotationTime: (time.Duration(c.loggerConfig.RotationTime) * time.Hour).String(),
		}),
		glog.WithFileLogLevel("debug"),
	}

	// 设置stack 日志界别 当日志级别高于stack日志级别时，才会打印stack信息
	if !stringer.IsBlank(c.loggerConfig.StackLogLevel) {
		opts = append(opts, glog.WithStackLogLevel(c.loggerConfig.StackLogLevel))
	}

	// 日志日志级别
	if level, ok := c._loggerLevel[strings.ToLower(moduleName)]; ok {
		// 配置文件中有日志级别
		opts = append(opts, glog.WithDefaultLogLevel(level))
	} else {
		// 配置文件中没有日志级别
		opts = append(opts, glog.WithDefaultLogLevel(c.loggerConfig.DefaultLogLevel))
		c._loggerLevel[strings.ToLower(moduleName)] = c.loggerConfig.DefaultLogLevel
	}

	// 日志输出格式
	if stringer.CompareIgnoreCase("console", c.loggerConfig.PrintFormat) {
		opts = append(opts, glog.WithConsoleFormat())
	} else if stringer.CompareIgnoreCase("json", c.loggerConfig.PrintFormat) {
		opts = append(opts, glog.WithJSONFormat())
	}

	logger := glog.NewSugaredLogger(opts...)

	c._logMap[moduleName] = logger

	return logger
}

func (c *Control) StartUp(failedFunc func(err error)) {
	// no-op
	c.once.Do(func() {
		fmt.Printf("starting up logger server...\n")
	})
}

func (c *Control) Shutdown() error {
	fmt.Printf("shutting down logger server...\n")
	// no-op
	return nil
}
