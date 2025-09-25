package mfa

import (
	// "github.com/google/uuid"
	// "github.com/pquerna/otp"
	// "github.com/pquerna/otp/totp"
	
	"github.com/jianlu8023/golang-example/pkg/control/config"
)


// 常量定义
const (
	// ProviderGoogle Google认证提供商
	ProviderGoogle = "google"
	// ProviderMicrosoft Microsoft认证提供商
	ProviderMicrosoft = "microsoft"
)

// getDefaultConfig 获取默认MFA配置
// @return *config.MFAConfig 默认MFA配置
func getDefaultConfig() *config.MFAConfig {
	return &config.MFAConfig{
		Enabled:         false,
		
	}
}


