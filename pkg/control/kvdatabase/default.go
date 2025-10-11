package kvdatabase

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取默认配置
func getDefaultConfig() *config.KvDatabaseConfig {
	return &config.KvDatabaseConfig{
		Enabled: false,
	}
}
