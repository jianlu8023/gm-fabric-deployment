package datasource

import (
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

const (
	// Clickhouse ClickHouse数据库类型标识
	Clickhouse = "clickhouse"
)

// newClickhouseConn 创建ClickHouse数据库连接
// @description 根据配置创建ClickHouse数据库连接
// @param dbControl *Control 数据源控制器实例
// @return *gorm.DB GORM数据库连接实例
// @return error 创建连接过程中可能产生的错误
func newClickhouseConn(dbControl *Control) (*gorm.DB, error) {
	// 设置表选项
	// db.Set("gorm:table_options", "ENGINE=Distributed(cluster, default, hits)").AutoMigrate(&User{})
	return gorm.Open(clickhouse.Open(dbControl.config.GenClickhouseDSN()),
		&gorm.Config{
			PrepareStmt:          true,
			DisableAutomaticPing: false,
			CreateBatchSize:      1000,
			Logger:               dbControl.dbLogger,
		})
}
