package mfa

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取默认MFA配置
// @return *config.MFAConfig 默认MFA配置
func getDefaultConfig() *config.MFAConfig {
	return &config.MFAConfig{
		Enabled: false,
	}
}
