package datasource

// import (
// 	"gorm.io/driver/gaussdb"
// 	"gorm.io/gorm"
// )

const (
	// GaussDB GaussDB数据库类型标识
	GaussDB = "gaussDB"
)

// func newGaussDBConn(dbControl *Control) (*gorm.DB, error) {
// 	return gorm.Open(gaussdb.New(gaussdb.Config{
// 		DSN:                  dbControl.config.GenGaussDBDSN(),
// 		PreferSimpleProtocol: false, // 禁用隐式 prepared statement
// 	}),
// 		&gorm.Config{
// 			PrepareStmt:          true,
// 			DisableAutomaticPing: false,
// 			CreateBatchSize:      1000,
// 			Logger:               dbControl.dbLogger,
// 		})
// }
