package ai

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.AIConfig {
	return &config.AIConfig{
		Enabled: false,
	}
}
