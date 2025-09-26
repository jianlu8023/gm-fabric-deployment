package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// MFASecretResponse MFA密钥响应数据
// @description MFA密钥生成响应的数据结构
// @property secret string MFA密钥
// @property qrCodeURL string 二维码URL
// @property qrCodeImage string 二维码图片的Base64编码
// @property provider string MFA提供商
// @property status string 状态信息
// @property message string 消息说明
type MFASecretResponse struct {
	Secret      string `json:"secret,omitempty" yaml:"secret,omitempty"`
	QrCodeURL   string `json:"qr_code_url,omitempty" yaml:"qr_code_url,omitempty"`
	QrCodeImage string `json:"qr_code_image,omitempty" yaml:"qr_code_image,omitempty"`
	Provider    string `json:"provider,omitempty" yaml:"provider,omitempty"`
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
// @param provider MFA提供商
// @return *MFASecretResponse MFA密钥响应对象
func NewMFASecretResponse(secret, qrCodeURL, qrCodeImage, provider string) *MFASecretResponse {
	return &MFASecretResponse{
		Secret:      secret,
		QrCodeURL:   qrCodeURL,
		QrCodeImage: qrCodeImage,
		Provider:    provider,
	}
}

// MFAVerifyResponse MFA验证响应数据
// @description MFA验证响应的数据结构
// @property valid bool 验证结果
// @property provider string MFA提供商
// @property status string 状态信息
// @property message string 消息说明
type MFAVerifyResponse struct {
	Valid    bool   `json:"valid,omitempty" yaml:"valid,omitempty"`
	Provider string `json:"provider,omitempty" yaml:"provider,omitempty"`
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
func NewMFAVerifyResponse(valid bool, provider string) *MFAVerifyResponse {
	return &MFAVerifyResponse{
		Valid:    valid,
		Provider: provider,
	}
}
