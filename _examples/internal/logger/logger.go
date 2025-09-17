package logger

import (
	"sync"
	"time"

	glog "github.com/jianlu8023/go-logger"
	"github.com/jianlu8023/go-logger/dblogger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var once sync.Once

var (
	appLogger *zap.SugaredLogger
	sdkLogger *zap.SugaredLogger
	dbLogger  *dblogger.Logger
)

func init() {
	once.Do(func() {
		appLogger = glog.NewSugaredLogger(
			&glog.Config{
				LogLevel:    "DEBUG",
				DevelopMode: true,
				StackLevel:  "ERROR",
				ModuleName:  "APP",
				Caller:      true,
			},
			glog.WithFileOutPut(),
			glog.WithConsoleFormat(),
			glog.WithLumberjack(
				&glog.LumberjackConfig{
					FileName:   "./logs/lumberjack-app.log",
					MaxAge:     30,
					MaxBackups: 7,
					MaxSize:    5,
					Compress:   true,
					Localtime:  true,
				},
			),
			glog.WithLumberjack(glog.LumberjackDefaultConfig()),
			glog.WithRotateLog(
				&glog.RotateLogConfig{
					FileName:     "./logs/rotatelogs-app.log",
					LocalTime:    true,
					RotationTime: "1h",
					MaxAge:       "365",
				},
			),
			glog.WithRotateLog(glog.RotateLogDefaultConfig()),
		)
		sdkLogger = glog.NewSugaredLogger(
			&glog.Config{
				LogLevel:    "DEBUG",
				DevelopMode: true,
				StackLevel:  "ERROR",
				ModuleName:  "SDK",
				Caller:      true,
			},
			glog.WithFileOutPut(),
			glog.WithConsoleFormat(),
			glog.WithLumberjack(
				&glog.LumberjackConfig{
					FileName:   "./logs/lumberjack-sdk.log",
					MaxAge:     30,
					MaxBackups: 7,
					MaxSize:    5,
					Compress:   true,
					Localtime:  true,
				},
			),
			glog.WithLumberjack(glog.LumberjackDefaultConfig()),
			glog.WithRotateLog(
				&glog.RotateLogConfig{
					FileName:     "./logs/rotatelogs-sdk.log",
					LocalTime:    true,
					RotationTime: "1h",
					MaxAge:       "365",
				},
			),
			glog.WithRotateLog(glog.RotateLogDefaultConfig()),
		)
		dbLogger = dblogger.NewDBLogger(
			dblogger.Config{
				LogLevel:                  dblogger.INFO,
				ShowSql:                   true,
				SlowThreshold:             2 * time.Second,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      true,
				Colorful:                  true,
			},
			dblogger.WithCustomLogger(glog.NewLogger(
				&glog.Config{
					LogLevel:    "DEBUG",
					DevelopMode: true,
					StackLevel:  "ERROR",
					ModuleName:  "DB",
					Caller:      true,
				},
				glog.WithConsoleFormat(),
				glog.WithLumberjack(
					&glog.LumberjackConfig{
						FileName:   "./logs/lumberjack-db.log",
						MaxAge:     30,
						MaxBackups: 7,
						MaxSize:    5,
						Compress:   true,
						Localtime:  true,
					},
				),
				glog.WithConsoleConfig(zapcore.EncoderConfig{
					MessageKey:     "msg",
					LevelKey:       "",
					TimeKey:        "",
					NameKey:        "",
					CallerKey:      "",
					FunctionKey:    "",
					StacktraceKey:  "",
					SkipLineEnding: false,
					LineEnding:     zapcore.DefaultLineEnding,
					EncodeLevel:    glog.CustomColorCapitalLevelEncoder,
					EncodeTime:     glog.CustomTimeEncoder,
					EncodeDuration: zapcore.SecondsDurationEncoder,
					EncodeCaller:   zapcore.ShortCallerEncoder,
					EncodeName:     zapcore.FullNameEncoder,
				}),

				glog.WithFileOutPut(),
				glog.WithLumberjack(glog.LumberjackDefaultConfig()),
				glog.WithRotateLog(
					&glog.RotateLogConfig{
						FileName:     "./logs/rotatelogs-db.log",
						LocalTime:    true,
						RotationTime: "1h",
						MaxAge:       "365",
					},
				),
				glog.WithRotateLog(glog.RotateLogDefaultConfig()),
			)),
		)
	})
}

func GetAppLogger() *zap.SugaredLogger {
	return appLogger
}

func GetSDKLogger() *zap.SugaredLogger {
	return sdkLogger
}

func GetDBLogger() *dblogger.Logger {
	return dbLogger
}
