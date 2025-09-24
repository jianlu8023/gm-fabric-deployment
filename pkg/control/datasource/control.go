package datasource

import (
	"sync"
	"time"

	"github.com/jianlu8023/go-logger/v2/dblogger"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Control struct {
	config           *config.DataSourceConfig
	dbConn           *gorm.DB
	logger           *zap.SugaredLogger
	dbLogger         *dblogger.Logger
	once             sync.Once
	autoMigrateTable []interface{}
	autoMigrateMutex sync.RWMutex
}

// RegisterAutoMigrateTable 注册自动迁移的表
// @param tables ...interface{} 要自动迁移的表结构
func (c *Control) RegisterAutoMigrateTable(tables ...interface{}) {
	c.logger.Debugf("[control] register auto migrate table...")

	c.autoMigrateMutex.Lock()
	defer c.autoMigrateMutex.Unlock()
	c.autoMigrateTable = append(c.autoMigrateTable, tables...)
	c.logger.Debugf("[control] register auto migrate table successfully...")
}

// autoMigrate 自动迁移表（内部方法）
// @return error 迁移过程中可能产生的错误
func (c *Control) autoMigrate() error {
	c.logger.Debugf("[control] auto migrate table...")
	c.autoMigrateMutex.RLock()
	defer c.autoMigrateMutex.RUnlock()
	if err := c.dbConn.AutoMigrate(c.autoMigrateTable...); err != nil {
		c.logger.Errorf("[control] auto migrate table failed: %s", err)
		return err
	}
	return nil
}

// ReAutoMigrate 手动触发自动迁移
// @return error 迁移过程中可能产生的错误
func (c *Control) ReAutoMigrate() error {
	c.logger.Debugf("[control] call auto migrate table by hand...")
	c.autoMigrateMutex.RLock()
	defer c.autoMigrateMutex.RUnlock()
	if err := c.dbConn.AutoMigrate(c.autoMigrateTable...); err != nil {
		c.logger.Errorf("[control] auto migrate table failed: %s", err)
		return err
	}
	return nil
}

// Close 关闭数据库连接
// @return error 关闭过程中可能产生的错误
// func (c *Control) Close() error {
// 	return nil
// }

// GetConn 获取数据库连接
// @return *gorm.DB 数据库连接实例
func (c *Control) GetConn() *gorm.DB {
	return c.dbConn
}

// // AutoMigrateTable 自动迁移表
// // @param tables interface{} 表结构
// // @return error 错误信息
// func (c *Control) AutoMigrateTable(tables ...interface{}) error {
// 	c.logger.Infof("[control] start auto migrate table...")
// 	if err := c.dbConn.AutoMigrate(tables...); err != nil {
// 		c.logger.Errorf("[control] auto migrate table failed: %v", err)
// 		return err
// 	}
// 	return nil
// }

// setConnPool 设置数据库连接池参数（内部方法）
// @return error 设置过程中可能产生的错误
func (c *Control) setConnPool() error {
	sqlDB, err := c.dbConn.DB()
	if err != nil {
		c.logger.Errorf("[control] faild from gorm.DB get sql.DB to set pool: %v", err)
		return err
	}
	sqlDB.SetMaxIdleConns(c.config.MaxIdleConn)
	sqlDB.SetMaxOpenConns(c.config.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Minute)
	c.logger.Debugf("[control] success to set conn pool...")
	return nil
}

// StartUp 启动数据源服务
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			if err := c.initDBConn(); err != nil {
				c.logger.Errorf("[control] init db conn failed: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}

			c.logger.Debugf("[control] call db ping instead startup...")
			sqlDB, err := c.dbConn.DB()
			if err != nil {
				c.logger.Errorf("[control] failed from gorm.DB get sql.DB: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
			if err = sqlDB.Ping(); err != nil {
				c.logger.Errorf("[control] failed from sqlDB.Ping: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}

			c.logger.Debugf("[control] starting to auto migrate tables...")
			if err := c.autoMigrate(); err != nil {
				c.logger.Errorf("[control] auto migrate table failed: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
		}
	})
}

// Shutdown 关闭数据源服务
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutdown datasource...")
	sqlDB, err := c.dbConn.DB()
	if err != nil {
		c.logger.Errorf("[control] failed from gorm.DB get sql.DB: %v", err)
		return err
	}
	if err = sqlDB.Close(); err != nil {
		c.logger.Errorf("[control] failed from sqlDB.Close: %v", err)
		return err
	}
	return nil
}

func (c *Control) initDBConn() error {
	switch c.config.DataSourceType {
	case Mysql:
		c.logger.Debugf("[control] using mysql data source...")
		conn, err := newMysqlConn(c)
		if err != nil {
			c.logger.Error("[control] failed to create mysql connection: %v", err)
			return err
		}
		c.logger.Debugf("[control] mysql connection create success...")
		c.dbConn = conn
	case Postgres:
		c.logger.Debugf("[control] using postgres data source...")
		conn, err := newPostgresConn(c)
		if err != nil {
			c.logger.Error("[control] failed to create postgres connection: %v", err)
			return err
		}
		c.logger.Debugf("[control] postgres connection create success...")
		c.dbConn = conn
	case Sqlite3:
		c.logger.Debugf("[control] using sqlite3 data source...")
		conn, err := newSqlite3Conn(c)
		if err != nil {
			c.logger.Error("[control] failed to create sqlite3 connection: %v", err)
			return err
		}
		c.logger.Debugf("[control] sqlite3 connection create success...")
		c.dbConn = conn
	default:
		c.logger.Warnf("[control] unknown data source type: %s", c.config.DataSourceType)
		return ErrUnknownDataSourceType
	}

	c.logger.Debugf("[control] set db conn pool...")
	if err := c.setConnPool(); err != nil {
		c.logger.Errorf("[control] failed to set db conn pool: %v", err)
		return err
	}
	return nil
}

// NewDataSourceControl 创建数据源控制器
// @param config *config.DataSourceConfig 数据源配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control 数据源控制器实例

func NewDataSourceControl(dbConfig *config.DataSourceConfig, loggerControl *logger.Control) *Control {
	if dbConfig == nil {
		dbConfig = getDefaultConfig()
	}
	if !dbConfig.Enabled {
		return nil
	}

	dsLogger := loggerControl.GenLogger(logger.ModuleDataSource)
	dsLogger.Infof("[control] starting new datasource control...")

	ctl := &Control{
		config:           dbConfig,
		logger:           dsLogger,
		dbLogger:         newDbLogger(loggerControl.GetConfig(), dbConfig.LogInConsole),
		autoMigrateTable: make([]interface{}, 0, 8),
	}

	return ctl
}
