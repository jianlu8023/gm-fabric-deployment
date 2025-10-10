package datasource

import (
	"gorm.io/gorm"
	// _ "modernc.org/SQLite" // 网上找到的
	// _"github.com/glebarez/go-SQLite" // 网上找到的
	// _ "github.com/glebarez/SQLite" // 网上找到的
	// "gorm.io/driver/sqlite" // cgo 实现
	"github.com/glebarez/sqlite" // 纯go 实现
)

const (
	// Sqlite3 SQLite3数据库类型标识
	Sqlite3 = "sqlite3"
)

// newSqlite3Conn 创建SQLite3数据库连接
// @description 根据配置创建SQLite3数据库连接
// @param dbControl *Control 数据源控制器实例
// @return *gorm.DB GORM数据库连接实例
// @return error 创建连接过程中可能产生的错误
func newSqlite3Conn(dbControl *Control) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbControl.config.GenSqlite3DSN()),
		&gorm.Config{
			PrepareStmt:          true,
			DisableAutomaticPing: false,
			CreateBatchSize:      1000,
			Logger:               dbControl.dbLogger,
		},
	)
}
