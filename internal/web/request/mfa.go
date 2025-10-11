package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// MFARecoverySecretRequest 生成MFA恢复密钥请求
//
// @description 用于生成MFA恢复密钥的请求参数
// @struct
type MFARecoverySecretRequest struct {
	UserID string `json:"user_id,omitempty" yaml:"user_id,omitempty" form:"userId" binding:"required,min=1,max=100"` // UserID 用户ID（必填）
}

// String 将生成MFA恢复密钥请求参数转换为字符串表示
//
// @description 将MFARecoverySecretRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req MFARecoverySecretRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 检查生成MFA恢复密钥请求参数是否合法
//
// @description 验证MFARecoverySecretRequest结构体中的必要参数是否满足基本要求
// @return bool 参数是否合法
func (req MFARecoverySecretRequest) IsLegal() bool {
	if stringer.IsBlank(req.UserID) {
		return false
	}
	return true
}

// MFAVerifyCodeRequest 验证MFA代码请求
//
// @description 用于验证用户输入的MFA验证码的请求参数
// @struct
type MFAVerifyCodeRequest struct {
	UserID string `json:"user_id,omitempty" yaml:"user_id,omitempty" form:"userId" binding:"required,min=1,max=100"` // UserID 用户ID（必填）
	Code   string `json:"code,omitempty" yaml:"code,omitempty" form:"code" binding:"required,len=6,numeric"`         // Code MFA验证码（必填，6位数字）
}

// String 将验证MFA代码请求参数转换为字符串表示
//
// @description 将MFAVerifyCodeRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req MFAVerifyCodeRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 检查MFA代码请求参数是否合法
//
// @description 验证MFAVerifyCodeRequest结构体中的必要参数是否满足基本要求
// @return bool 参数是否合法
func (req MFAVerifyCodeRequest) IsLegal() bool {
	if stringer.IsBlank(req.Code) ||
		stringer.IsBlank(req.UserID) {
		return false
	}
	return true
}

// MFAQrcodeRequest 生成MFA二维码请求
//
// @description 用于生成MFA二维码的请求参数
// @struct
type MFAQrcodeRequest struct {
	UserID string `json:"user_id,omitempty" yaml:"user_id,omitempty" form:"userId" binding:"required,min=1,max=100"` // UserID 用户ID（必填）
}

// IsLegal 检查生成MFA二维码请求参数是否合法
//
// @description 验证MFAQrcodeRequest结构体中的必要参数是否满足基本要求
// @return bool 参数是否合法
func (req MFAQrcodeRequest) IsLegal() bool {
	if stringer.IsBlank(req.UserID) {
		return false
	}
	return true
}

// String 将生成MFA二维码请求参数转换为字符串表示
//
// @description 将MFAQrcodeRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req MFAQrcodeRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
