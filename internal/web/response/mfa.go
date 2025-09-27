package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// MFASecretResponse MFA密钥响应数据
// @description MFA密钥生成响应的数据结构
// @struct
// @property RecoverySecret []string MFA恢复密钥
type MFASecretResponse struct {
	RecoverySecret []string `json:"recovery_secret" yaml:"recovery_secret"` // MFA恢复密钥
}

// MarshalJSON 自定义JSON序列化方法
// @description 将MFASecretResponse结构体转换为JSON格式
// @return []byte JSON字节数组
// @return error 错误信息
func (resp MFASecretResponse) MarshalJSON() ([]byte, error) {
	type Alias MFASecretResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// String 将MFASecretResponse转换为字符串
// @description 将MFASecretResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp MFASecretResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewMFASecretResponse 创建MFA密钥响应对象
// @description 创建一个新的MFA密钥响应对象
// @param recoverySecret []string MFA恢复密钥
// @return *MFASecretResponse MFA密钥响应对象
func NewMFASecretResponse(recoverySecret []string) MFASecretResponse {
	return MFASecretResponse{
		RecoverySecret: recoverySecret,
	}
}

// MFAVerifyResponse MFA验证响应数据
// @description MFA验证响应的数据结构
// @struct
// @property Valid bool 验证结果
type MFAVerifyResponse struct {
	Valid bool `json:"valid" yaml:"valid"` // 验证结果
}

// MarshalJSON 自定义JSON序列化方法
// @description 将MFAVerifyResponse结构体转换为JSON格式
// @return []byte JSON字节数组
// @return error 错误信息
func (resp MFAVerifyResponse) MarshalJSON() ([]byte, error) {
	type Alias MFAVerifyResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// String 将MFAVerifyResponse转换为字符串
// @description 将MFAVerifyResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp MFAVerifyResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewMFAVerifyResponse 创建MFA验证响应对象
// @description 创建一个新的MFA验证响应对象
// @param valid bool 验证结果
// @return MFAVerifyResponse MFA验证响应对象
func NewMFAVerifyResponse(valid bool) MFAVerifyResponse {
	return MFAVerifyResponse{
		Valid: valid,
	}
}

// MFAQrCodeResponse MFA二维码响应数据
// @description MFA二维码生成响应的数据结构
// @struct
// @property QrCodeImage string 二维码图片的Base64编码
type MFAQrCodeResponse struct {
	QrCodeImage string `json:"qr_code_image" yaml:"qr_code_image"` // 二维码图片的Base64编码
}

// MarshalJSON 自定义JSON序列化方法
// @description 将MFAQrCodeResponse结构体转换为JSON格式
// @return []byte JSON字节数组
// @return error 错误信息
func (resp MFAQrCodeResponse) MarshalJSON() ([]byte, error) {
	type Alias MFAQrCodeResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// String 将MFAQrCodeResponse转换为字符串
// @description 将MFAQrCodeResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp MFAQrCodeResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewMFAQrCodeResponse 创建MFA二维码响应对象
// @description 创建一个新的MFA二维码响应对象
// @param qrCodeImage string 二维码图片的Base64编码
// @return MFAQrCodeResponse MFA二维码响应对象
func NewMFAQrCodeResponse(qrCodeImage string) MFAQrCodeResponse {
	return MFAQrCodeResponse{
		QrCodeImage: qrCodeImage,
	}
}
