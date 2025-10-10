package datasource

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	// Postgres PostgreSQL数据库类型标识
	Postgres = "postgres"
)

// newPostgresConn 创建PostgreSQL数据库连接
// @description 根据配置创建PostgreSQL数据库连接
// @param dbControl *Control 数据源控制器实例
// @return *gorm.DB GORM数据库连接实例
// @return error 创建连接过程中可能产生的错误
func newPostgresConn(dbControl *Control) (*gorm.DB, error) {
	return gorm.Open(
		postgres.New(postgres.Config{
			DSN:                  dbControl.config.GenPostgresDSN(),
			PreferSimpleProtocol: false, // 禁用隐式 prepared statement
		}),
		&gorm.Config{
			PrepareStmt:          true,
			DisableAutomaticPing: false,
			CreateBatchSize:      1000,
			Logger:               dbControl.dbLogger,
		},
	)
}
