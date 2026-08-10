package libp2p

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取libp2p默认配置
//
// @description 返回libp2p的默认配置，默认不启用，需在配置文件中显式开启
// @return *config.Libp2pConfig 默认配置
func getDefaultConfig() *config.Libp2pConfig {
	return &config.Libp2pConfig{
		Enabled:             false,
		ProtocolID:          "/golang-example/msg/1.0.0",
		ServiceTag:          "golang-example_tags",
		HealthCheckInterval: 60,
		MessageQueueSize:    1024,
		WorkerCount:         5,
		BroadcastInterval:   5,
		DHTMode:             "server",
		Persistence: &config.PersistenceConfig{
			Enabled:      false,
			FilePath:     "data/peerstore.json",
			SaveInterval: 300,
		},
	}
}
