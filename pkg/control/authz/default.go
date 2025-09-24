package authz

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.AuthzConfig {
	return &config.AuthzConfig{
		Enabled:        false,
		ModelFile:      "configs/authz/model.conf",
		PolicyFile:     "configs/authz/policy.csv",
		AutoCreateFile: true,
		DefaultAllow:   true,
		LogEnabled:     true,
	}
}
