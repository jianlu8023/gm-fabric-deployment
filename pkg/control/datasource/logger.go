package datasource

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	glog "github.com/jianlu8023/go-logger/v2"
	"github.com/jianlu8023/go-logger/v2/dblogger"
	"go.uber.org/zap/zapcore"
	"path/filepath"
	"time"
)

func newDbLogger(loggerConfig *config.LoggerConfig) *dblogger.Logger {
	var fileName = "sql.log"
	fileDir := filepath.Dir(loggerConfig.FilePath)
	fileName = filepath.Clean(filepath.Join(fileDir, fileName))
	fmt.Printf("fileName: %s\n", fileName)
	opts := []glog.Option{
		glog.WithModuleName("Sql"),
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
