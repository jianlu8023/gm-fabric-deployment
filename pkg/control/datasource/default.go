package datasource

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取默认的数据源配置
// @description 返回一个默认的数据源配置实例
// @return *config.DataSourceConfig 默认的数据源配置
func getDefaultConfig() *config.DataSourceConfig {
	return &config.DataSourceConfig{
		Enabled:        false,
		DataSourceType: "sqlite3",
		UserName:       "myself",
		Password:       "myself_pw",
		Host:           "localhost",
		Port:           3306,
		DataBaseName:   "myself_example",
		DataBasePath:   "db/golang-example.db",
		MaxIdleConn:    10,
		MaxOpenConn:    50,
		LogInConsole:   true,
	}
}
