package datasource

import (
	"path/filepath"
	"time"

	dblogger "github.com/jianlu8023/go-logger/db-logger/v2"
	glog "github.com/jianlu8023/go-logger/v2"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"go.uber.org/zap/zapcore"
)

// newDbLogger 创建数据库日志记录器
// @description 创建用于记录数据库操作的日志记录器
// @param loggerConfig *config.LoggerConfig 日志配置
// @param logInConsole bool 是否在控制台输出日志
// @return *dblogger.Logger 数据库日志记录器实例
func newDbLogger(loggerConfig *config.LoggerConfig, logInConsole bool) *dblogger.Logger {
	fileName := "sql.log"
	fileDir := filepath.Dir(loggerConfig.FilePath)
	fileName = filepath.Clean(filepath.Join(fileDir, fileName))
	opts := []glog.Option{
		glog.WithModuleName("Sql"),
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

	logger := glog.NewLogger(opts...)
	// logger.Debug("testing logger...")
	dbLogger := dblogger.NewDBLogger(dblogger.Config{
		LogLevel:                  dblogger.INFO,
		SlowThreshold:             100 * time.Millisecond,
		Colorful:                  false,
		IgnoreRecordNotFoundError: false,
		ParameterizedQueries:      false,
		ShowSql:                   true,
	},
		dblogger.WithCustomLogger(logger),
	)
	return dbLogger
}
