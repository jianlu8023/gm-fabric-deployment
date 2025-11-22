package mfa

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
	// @param secret string 密钥
	// @param code string MFA代码
	// @return bool 验证结果
	VerifyCode(secret string, code string) bool

	GenerateQrCode(otpauthURL string) string

	GenRecoverySecret(userId string, recoveryNum int) ([]string, error)
}

// 常量定义
const (
	// ProviderGoogle Google认证提供商
	ProviderGoogle = "google"
	// ProviderMicrosoft Microsoft认证提供商
	ProviderMicrosoft = "microsoft"
)
