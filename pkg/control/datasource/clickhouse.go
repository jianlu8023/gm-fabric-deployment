package datasource

// import (
// 	"gorm.io/driver/clickhouse"
// 	"gorm.io/gorm"
// )

const (
	Clickhouse = "clickhouse"
)

// func newClickhouseConn(dbControl *Control) (*gorm.DB, error) {
// 	// 设置表选项
// 	// db.Set("gorm:table_options", "ENGINE=Distributed(cluster, default, hits)").AutoMigrate(&User{})
// 	return gorm.Open(clickhouse.Open(dbControl.config.GenClickhouseDSN()),
// 		&gorm.Config{
// 			PrepareStmt:          true,
// 			DisableAutomaticPing: false,
// 			CreateBatchSize:      1000,
// 			Logger:               dbControl.dbLogger,
// 		})
// }
