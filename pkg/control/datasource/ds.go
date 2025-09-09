package datasource

import (
	"time"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/go-logger/v2/dblogger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Control struct {
	dbConfig *config.DataSourceConfig
	dbConn   *gorm.DB
	logger   *zap.SugaredLogger
	dbLogger *dblogger.Logger
}

func (c *Control) Close() error {
	return nil
}

func (c *Control) GetConn() *gorm.DB {
	return c.dbConn
}

// AutoMigrateTable 自动迁移表
// @param tables interface{} 表结构
// @return error 错误信息
func (c *Control) AutoMigrateTable(tables ...interface{}) error {
	c.logger.Infof("[control] start auto migrate table...")
	if err := c.dbConn.AutoMigrate(tables...); err != nil {
		c.logger.Errorf("[control] auto migrate table failed: %v", err)
		return err
	}
	return nil
}

func (c *Control) setConnPool() error {
	sqlDB, err := c.dbConn.DB()
	if err != nil {
		c.logger.Errorf("[control] faild from gorm.DB get sql.DB to set pool: %v", err)
		return err
	}
	sqlDB.SetMaxIdleConns(c.dbConfig.MaxIdleConn)
	sqlDB.SetMaxOpenConns(c.dbConfig.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(time.Minute)
	c.logger.Debugf("[control] success to set conn pool...")
	return nil
}

func NewDataSourceControl(dbConfig *config.DataSourceConfig, loggerControl *logger.Control) (*Control, error) {
	dsLogger := loggerControl.GenLogger(logger.ModuleDataSource)
	dsLogger.Infof("[control] starting new datasource control...")

	ctl := &Control{
		dbConfig: dbConfig,
		logger:   dsLogger,
		dbLogger: newDbLogger(loggerControl.GetConfig(), dbConfig.LogInConsole),
	}

	switch dbConfig.DataSourceType {
	case Mysql:
		dsLogger.Debugf("[control] using mysql data source...")
		conn, err := newMysqlConn(ctl)
		if err != nil {
			dsLogger.Error("[control] failed to create mysql connection: %v", err)
			return nil, err
		}
		dsLogger.Debugf("[control] mysql connection create success...")
		ctl.dbConn = conn
	case Postgres:
		dsLogger.Debugf("[control] using postgres data source...")
		conn, err := newPostgresConn(ctl)
		if err != nil {
			dsLogger.Error("[control] failed to create postgres connection: %v", err)
			return nil, err
		}
		dsLogger.Debugf("[control] postgres connection create success...")
		ctl.dbConn = conn
	case Sqlite3:
		dsLogger.Debugf("[control] using sqlite3 data source...")
		conn, err := newSqlite3Conn(ctl)
		if err != nil {
			dsLogger.Error("[control] failed to create sqlite3 connection: %v", err)
			return nil, err
		}
		dsLogger.Debugf("[control] sqlite3 connection create success...")
		ctl.dbConn = conn
	default:
		dsLogger.Warnf("[control] unknown data source type: %s", dbConfig.DataSourceType)
		return nil, ErrUnknownDataSourceType
	}

	dsLogger.Debugf("[control] set db conn pool...")
	if err := ctl.setConnPool(); err != nil {
		dsLogger.Errorf("[control] failed to set db conn pool: %v", err)
		return nil, err
	}
	return ctl, nil
}
