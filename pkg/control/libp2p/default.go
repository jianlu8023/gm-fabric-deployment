package libp2p

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.Libp2pConfig {
	return &config.Libp2pConfig{
		Enabled: false,
	}
}
