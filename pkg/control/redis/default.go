package redis

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取默认的Redis配置
// @return *config.RedisConfig 默认的Redis配置
func getDefaultConfig() *config.RedisConfig {
	return &config.RedisConfig{
		Enabled: false,
	}
}
