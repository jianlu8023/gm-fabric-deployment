package logger

import (
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/str"
	glog "github.com/jianlu8023/go-logger/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strconv"
	"strings"
	"sync"
)

var (
	_logMap = make(map[string]*zap.SugaredLogger)
)

type Control struct {
	loggerConfig *config.LoggerConfig
	_logMap      map[string]*zap.SugaredLogger
	_loggerLevel map[string]string
	sync.RWMutex
}

func NewLoggerControl(loggerConfig *config.LoggerConfig) *Control {
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
	c.Lock()
	defer c.Unlock()
	existLogger, ok := _logMap[moduleName]
	if ok {
		return existLogger
	}

	opts := []glog.Option{
		glog.WithModuleName(moduleName),
		glog.WithStackLogLevel("error"),
		glog.WithCaller(),
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
			FileName:     c.loggerConfig.FilePath,
			MaxAge:       strconv.Itoa(c.loggerConfig.MaxAge),
			LocalTime:    true,
			RotationTime: strconv.Itoa(c.loggerConfig.RotationTime),
		}),
	}

	// 日志日志级别
	if level, ok := c._loggerLevel[strings.ToLower(moduleName)]; ok {
		// 配置文件中有日志级别
		opts = append(opts, glog.WithLogLevel(level))
	} else {
		// 配置文件中没有日志级别
		opts = append(opts, glog.WithLogLevel(c.loggerConfig.DefaultLogLevel))
		c._loggerLevel[strings.ToLower(moduleName)] = c.loggerConfig.DefaultLogLevel
	}

	// 日志输出格式
	if str.CompareIgnoreCase("console", c.loggerConfig.PrintFormat) {
		opts = append(opts, glog.WithConsoleFormat())
	} else if str.CompareIgnoreCase("json", c.loggerConfig.PrintFormat) {
		opts = append(opts, glog.WithJSONFormat())
	}

	logger := glog.NewSugaredLogger(opts...)

	_logMap[moduleName] = logger

	return logger
}
