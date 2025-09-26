package mfa

import (
	// "github.com/google/uuid"
	// "github.com/pquerna/otp"
	// "github.com/pquerna/otp/totp"

	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// Provider 认证提供商接口
// @interface Provider
// @method GenerateSecret(userID string) (string, string, error) 生成MFA密钥
// @method VerifyCode(userID string, code string) bool 验证MFA代码
type Provider interface {
	// GenerateSecret 生成MFA密钥
	// @param userID string 用户ID
	// @return string 密钥
	// @return string 二维码URL
	// @return error 生成过程中的错误
	GenerateSecret(userID string) (string, string, error)

	// VerifyCode 验证MFA代码
	// @param secret string 用户ID
	// @param code string MFA代码
	// @return bool 验证结果
	VerifyCode(secret string, code string) bool

	GenerateQrCode(otpauthURL string) string
}

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
		Enabled: false,
	}
}
