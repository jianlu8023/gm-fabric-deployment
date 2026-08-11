package kvdatabase

import (
	"context"
	"fmt"
	"sync"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.uber.org/zap"
)

// Control KvDatabase控制结构体
type Control struct {
	config        *config.KvDatabaseConfig
	logger        *zap.SugaredLogger
	ctx           context.Context
	cancel        context.CancelFunc
	once          sync.Once
	tracerControl *tracer.Control
	db            KvDatabase
}

// NewKvDatabaseControl 创建一个新的KvDatabase控制器
// @param kvDatabaseConfig *config.KvDatabaseConfig kvdatabase配置
// @param loggerControl *logger.Control 日志控制器
// @param opts ...Option 可选的配置选项
// @return *Control kvdatabase控制器
func NewKvDatabaseControl(kvDatabaseConfig *config.KvDatabaseConfig, loggerControl *logger.Control, opts ...Option) *Control {
	kvDatabaseConfig = check.IF[*config.KvDatabaseConfig](kvDatabaseConfig == nil,
		getDefaultConfig(),
		kvDatabaseConfig,
	)
	// 如果KvDatabase未启用，返回nil
	if !kvDatabaseConfig.Enabled {
		return nil
	}
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	// kvDatabaseLogger := loggerControl.GenLogger("kvdatabase")
	kvDatabaseLogger := loggerControl.GenLogger("")
	kvDatabaseLogger.Infof("[kvcache/control] starting new kvdatabase control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化KvDatabaseControl
	kc := &Control{
		logger: kvDatabaseLogger,
		ctx:    ctx,
		cancel: cancel,
		config: kvDatabaseConfig,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(kc)
	}

	kvDatabaseLogger.Infof("[kvcache/control] kvdatabase control started successfully")
	return kc
}

// StartUp 启动KvDatabase服务
// @description 启动KvDatabase服务，初始化数据库连接
// @param failedFunc func(err error) 启动失败回调函数
func (kc *Control) StartUp(failedFunc func(err error)) {
	kc.once.Do(func() {
		if kc.config.Enabled {
			kc.logger.Debugf("[kvcache/control] kvdatabase service is starting...")

			// 确保数据库目录存在
			if err := kc.ensureDbPath(); err != nil {
				kc.logger.Errorf("[kvcache/control] failed to ensure db path: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}

			// 初始化数据库
			if err := kc.initDatabase(); err != nil {
				kc.logger.Errorf("[kvcache/control] kvdatabase initialization database failed: %v", err)
				kc.cancel()
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}

			kc.logger.Infof("[kvcache/control] kvdatabase service started successfully")
		}
	})
}

// Shutdown 关闭KvDatabase数据库连接
// @description 关闭KvDatabase数据库连接并释放资源
// @return error 关闭过程中的错误
func (kc *Control) Shutdown() error {
	kc.logger.Infof("[kvcache/control] shutting down kvdatabase control...")
	kc.cancel()

	if kc.db != nil {
		if err := kc.db.Close(); err != nil {
			kc.logger.Errorf("[kvcache/control] failed to close kvdatabase database: %v", err)
			return err
		}
	}

	kc.logger.Infof("[kvcache/control] kvdatabase control shutdown successfully")
	return nil
}

// Get 获取键对应的值
// @description 获取键对应的值
// @param key string 键
// @return []byte 键对应的值
// @return error 获取过程中的错误
func (kc *Control) Get(key string) ([]byte, error) {
	if kc.db == nil {
		return nil, fmt.Errorf("kvdatabase database is nil")
	}

	value, err := kc.db.Get(key)
	if err != nil {
		kc.logger.Errorf("[kvcache/control] failed to get key %s: %v", key, err)
		return nil, err
	}

	kc.logger.Debugf("[kvcache/control] get key %s successfully", key)
	return value, nil
}

// Set 设置键值对
// @description 设置键值对
// @param key string 键
// @param value []byte 值
// @return error 设置过程中的错误
func (kc *Control) Set(key string, value []byte) error {
	if kc.db == nil {
		return fmt.Errorf("kvdatabase database is nil")
	}

	err := kc.db.Set(key, value)
	if err != nil {
		kc.logger.Errorf("[kvcache/control] failed to set key %s: %v", key, err)
		return err
	}

	kc.logger.Debugf("[kvcache/control] set key %s successfully", key)
	return nil
}

// Delete 删除键
// @description 删除指定的键
// @param key string 要删除的键
// @return error 删除过程中的错误
func (kc *Control) Delete(key string) error {
	if kc.db == nil {
		return fmt.Errorf("kvdatabase database is nil")
	}

	err := kc.db.Delete(key)
	if err != nil {
		kc.logger.Errorf("[kvcache/control] failed to delete key %s: %v", key, err)
		return err
	}

	kc.logger.Debugf("[kvcache/control] deleted key %s successfully", key)
	return nil
}

// ensureDbPath 确保数据库路径存在
// @description 确保数据库路径存在，如果不存在则创建
// @return error 创建过程中的错误
func (kc *Control) ensureDbPath() error {
	if kc.config.DbPath == "" {
		return fmt.Errorf("db path is empty")
	}

	_, _ = path.CreateDir(kc.config.DbPath)

	return nil
}

// initDatabase 初始化数据库
// @description 根据配置初始化对应的数据库
// @return error 初始化过程中的错误
func (kc *Control) initDatabase() error {
	kc.logger.Debugf("[kvcache/control] initializing %s database...", kc.config.DbType)

	// 根据数据库类型初始化对应的数据库实现
	switch kc.config.DbType {
	case levelDBDatabase:
		db, err := NewLevelDB(kc.config.DbPath, kc.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize leveldb: %v", err)
		}
		kc.db = db
		kc.logger.Infof("[kvcache/control] leveldb initialized successfully at %s", kc.config.DbPath)
	case pebbleDatabase:
		db, err := NewPebble(kc.config.DbPath, kc.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize pebble: %v", err)
		}
		kc.db = db
		kc.logger.Infof("[kvcache/control] pebble initialized successfully at %s", kc.config.DbPath)
	case badger4Database:
		db, err := NewBadger4(kc.config.DbPath, kc.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize badger: %v", err)
		}
		kc.db = db
		kc.logger.Infof("[kvcache/control] badger initialized successfully at %s", kc.config.DbPath)
	default:
		return fmt.Errorf("unsupported database type: %s", kc.config.DbType)
	}

	return nil
}
