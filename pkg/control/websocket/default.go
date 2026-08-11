package websocket

import (
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取默认的WebSocket配置
//
// @description 返回一个默认禁用的WebSocket配置，所有超时/缓冲参数使用项目约定值
// @return *config.WebSocketConfig 默认WebSocket配置
func getDefaultConfig() *config.WebSocketConfig {
	return &config.WebSocketConfig{
		Enabled:          false,
		PingInterval:     54 * time.Second,
		PongWait:         60 * time.Second,
		WriteWait:        10 * time.Second,
		MaxMessageSize:   4096,
		SendBufferSize:   256,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 10 * time.Second,
		CheckOrigins:     nil,
	}
}
