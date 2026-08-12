package webrtc

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// 默认 WebRTC 配置常量
const (
	// DefaultMinPort 默认 UDP 最小端口
	DefaultMinPort = 30000
	// DefaultMaxPort 默认 UDP 最大端口
	DefaultMaxPort = 40000
)

// getDefaultConfig 获取默认的WebRTC配置
//
// @description 获取默认的WebRTC配置，返回一个禁用状态的配置对象，并预置合理的默认端口范围与公开 STUN 服务器，
// 供上层启用时直接使用
// @return *config.WebRTCConfig 默认的WebRTC配置对象
func getDefaultConfig() *config.WebRTCConfig {
	return &config.WebRTCConfig{
		Enabled: false,
		ICEServers: []*config.ICEServerConfig{
			{
				URL: "stun:stun.l.google.com:19302",
			},
		},
		MinPort:             DefaultMinPort,
		MaxPort:             DefaultMaxPort,
		DisableMulticastDNS: true, // 服务端默认禁用mDNS，避免Windows下pion/mdns库的已知警告
	}
}
