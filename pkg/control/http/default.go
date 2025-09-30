package http

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.HttpServerConfig {
	return &config.HttpServerConfig{
		Enabled:      false,
		Http2Enabled: true, // 默认启用HTTP/2支持
	}
}
