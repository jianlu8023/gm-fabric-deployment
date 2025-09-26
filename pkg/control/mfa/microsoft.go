package mfa

import (
	"encoding/base64"
	"fmt"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
)

// MicrosoftProvider Microsoft认证提供商实现
// @struct MicrosoftProvider
// @implements Provider
type MicrosoftProvider struct {
	logger   *zap.SugaredLogger
	tenantID string // 租户ID
	clientID string // 客户端ID
}

func (m *MicrosoftProvider) GenRecoverySecret(userId string, recoveryNum int) ([]string, error) {
	if stringer.IsBlank(userId) {
		return nil, ErrInvalidUserID
	}
	recoverySecret := make([]string, 0, recoveryNum)

	for i := 0; i < recoveryNum; i++ {
		secret, _, err := m.GenerateSecret(userId)
		if err != nil {
			return nil, err
		}
		recoverySecret = append(recoverySecret, secret)
	}
	return recoverySecret, nil
}

// GenerateSecret 生成Microsoft认证器的MFA密钥
// @param userID string 用户ID
// @return string 密钥
// @return string 二维码URL
// @return error 生成过程中的错误
func (m *MicrosoftProvider) GenerateSecret(userID string) (string, string, error) {
	if userID == "" {
		return "", "", ErrInvalidUserID
	}

	// 对于Microsoft，我们也使用TOTP，但格式化用户名以符合Microsoft的规范
	formattedUserID := fmt.Sprintf("%s@%s", userID, m.tenantID)

	// 生成TOTP密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Microsoft",
		AccountName: formattedUserID,
		Period:      30,
		SecretSize:  20,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		m.logger.Errorf("[control] failed to generate Microsoft TOTP key: %v", err)
		return "", "", ErrGenerateSecretFailed
	}

	// 存储密钥
	m.logger.Debugf("[control] Microsoft MFA secret generated for user: %s", userID)

	// 返回密钥和二维码URL
	return key.Secret(), key.URL(), nil
}

// VerifyCode 验证Microsoft认证器的MFA代码
// @param userID string 用户ID
// @param code string MFA代码
// @return bool 验证结果
func (m *MicrosoftProvider) VerifyCode(secret string, code string) bool {

	// 验证代码
	valid := totp.Validate(code, secret)

	m.logger.Debugf("[control] Microsoft MFA code validation for valid: %v", valid)
	return valid
}

// GenerateQrCode 生成MFA二维码图片
// @param otpauthURL string OTP认证URL
// @return string 二维码图片的Base64编码
func (m *MicrosoftProvider) GenerateQrCode(otpauthURL string) string {
	// 创建二维码
	qr, err := qrcode.New(otpauthURL, qrcode.Medium)
	if err != nil {
		m.logger.Errorf("[control] failed to create QR code: %v", err)
		return ""
	}

	// 调整二维码尺寸
	qr.DisableBorder = false

	// 生成PNG并转换为Base64
	qrBytes, err := qr.PNG(240)
	if err != nil {
		m.logger.Errorf("[control] failed to generate QR code PNG: %v", err)
		return ""
	}

	// 转换为Base64编码的图片数据
	base64Str := base64.StdEncoding.EncodeToString(qrBytes)
	return "data:image/png;base64," + base64Str
}
