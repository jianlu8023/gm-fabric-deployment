package webrtc

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.WebRTCConfig {
	return &config.WebRTCConfig{
		Enabled: false,
	}
}
