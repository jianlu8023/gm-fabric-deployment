package certificate

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.CertificateConfig {
	return &config.CertificateConfig{
		Enabled: false,
	}
}
