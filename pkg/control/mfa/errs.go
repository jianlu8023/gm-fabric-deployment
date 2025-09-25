package mfa

import (
	"errors"
)



// 错误定义
var (
	// ErrMissingGoogleIssuer 缺少Google认证器颁发者名称
	ErrMissingGoogleIssuer = errors.New("google issuer is required")
	// ErrMissingMicrosoftTenantID 缺少Microsoft租户ID
	ErrMissingMicrosoftTenantID = errors.New("microsoft tenant id is required")
	// ErrMissingMicrosoftClientID 缺少Microsoft客户端ID
	ErrMissingMicrosoftClientID = errors.New("microsoft client id is required")
	// ErrGenerateSecretFailed 生成密钥失败
	ErrGenerateSecretFailed = errors.New("failed to generate MFA secret")
	// ErrInvalidUserID 用户ID无效
	ErrInvalidUserID = errors.New("invalid user id")
	// ErrInvalidCode 验证码无效
	ErrInvalidCode = errors.New("invalid MFA code")
	// ErrSecretNotFound 密钥未找到
	ErrSecretNotFound = errors.New("MFA secret not found")
)
