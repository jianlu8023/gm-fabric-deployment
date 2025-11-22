package email

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.EmailConfig {
	return &config.EmailConfig{
		Enabled: false,
	}
}
