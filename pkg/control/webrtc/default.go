package webrtc

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取默认的WebRTC配置
//
// @description 获取默认的WebRTC配置，返回一个禁用状态的配置对象
// @return *config.WebRTCConfig 默认的WebRTC配置对象
func getDefaultConfig() *config.WebRTCConfig {
	return &config.WebRTCConfig{
		Enabled: false,
	}
}
