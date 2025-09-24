package datasource

// import (
// 	"gorm.io/driver/sqlserver"
// 	"gorm.io/gorm"
// )

const (
	SqlServer = "sqlserver"
)

// func newSqlServerConn(dbControl *Control) (*gorm.DB, error) {
// 	return gorm.Open(sqlserver.Open(dbControl.config.GenSqlServerDSN()),
// 		&gorm.Config{
// 			PrepareStmt:          true,
// 			DisableAutomaticPing: false,
// 			CreateBatchSize:      1000,
// 			Logger:               dbControl.dbLogger,
// 		},
// 	)
// }
