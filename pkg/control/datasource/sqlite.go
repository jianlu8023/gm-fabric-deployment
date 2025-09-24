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
	Sqlite3 = "sqlite3"
)

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
