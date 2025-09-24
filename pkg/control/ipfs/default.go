package ipfs

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.IpfsConfig {
	return &config.IpfsConfig{
		Enabled: false,
	}
}
