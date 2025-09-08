package datasource

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const Postgres = "postgres"

func newPostgresConn(dbControl *Control) (*gorm.DB, error) {
	return gorm.Open(
		postgres.New(postgres.Config{
			DSN:                  dbControl.dbConfig.GenPostgresDSN(),
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
