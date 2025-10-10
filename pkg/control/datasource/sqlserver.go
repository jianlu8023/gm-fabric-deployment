package datasource

import (
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

const (
	// SqlServer SQL Server数据库类型标识
	SqlServer = "sqlserver"
)

// newSqlServerConn 创建SQL Server数据库连接
// @description 根据配置创建SQL Server数据库连接
// @param dbControl *Control 数据源控制器实例
// @return *gorm.DB GORM数据库连接实例
// @return error 创建连接过程中可能产生的错误
func newSqlServerConn(dbControl *Control) (*gorm.DB, error) {
	return gorm.Open(sqlserver.Open(dbControl.config.GenSqlServerDSN()),
		&gorm.Config{
			PrepareStmt:          true,
			DisableAutomaticPing: false,
			CreateBatchSize:      1000,
			Logger:               dbControl.dbLogger,
		},
	)
}
