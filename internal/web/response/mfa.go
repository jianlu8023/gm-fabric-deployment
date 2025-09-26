package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// MFASecretResponse MFA密钥响应数据
// @description MFA密钥生成响应的数据结构
// @property RecoverySecret string MFA恢复密钥
type MFASecretResponse struct {
	RecoverySecret []string `json:"recovery_secret,omitempty" yaml:"recovery_secret,omitempty"`
}

func (resp MFASecretResponse) MarshalJSON() ([]byte, error) {
	type Alias MFASecretResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

func (resp MFASecretResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewMFASecretResponse 创建MFA密钥响应对象
// @param secret MFA密钥
// @param qrCodeURL 二维码URL
// @param qrCodeImage 二维码图片的Base64编码
// @return *MFASecretResponse MFA密钥响应对象
func NewMFASecretResponse(recoverySecret []string) *MFASecretResponse {
	return &MFASecretResponse{
		RecoverySecret: recoverySecret,
	}
}

// MFAVerifyResponse MFA验证响应数据
// @description MFA验证响应的数据结构
// @property valid bool 验证结果
// @property provider string MFA提供商
// @property status string 状态信息
// @property message string 消息说明
type MFAVerifyResponse struct {
	Valid bool `json:"valid,omitempty" yaml:"valid,omitempty"`
}

func (resp MFAVerifyResponse) MarshalJSON() ([]byte, error) {
	type Alias MFAVerifyResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

func (resp MFAVerifyResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewMFAVerifyResponse 创建MFA验证响应对象
// @param valid 验证结果
// @param provider MFA提供商
// @return *MFAVerifyResponse MFA验证响应对象
func NewMFAVerifyResponse(valid bool) MFAVerifyResponse {
	return MFAVerifyResponse{
		Valid: valid,
	}
}

type MFAQrCodeResponse struct {
	QrCodeImage string `json:"qr_code_image,omitempty" yaml:"qr_code_image,omitempty"`
}

func (resp MFAQrCodeResponse) MarshalJSON() ([]byte, error) {
	type Alias MFAQrCodeResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

func (resp MFAQrCodeResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func NewMFAQrCodeResponse(qrCodeImage string) MFAQrCodeResponse {
	return MFAQrCodeResponse{
		QrCodeImage: qrCodeImage,
	}
}
