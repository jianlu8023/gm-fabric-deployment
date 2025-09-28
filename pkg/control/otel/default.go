package otel

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.OTELConfig {
	return &config.OTELConfig{
		Enabled: false,
	}
}
