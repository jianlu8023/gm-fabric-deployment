package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// MFARecoverySecretRequest 生成MFA密钥请求
// @description 生成MFA密钥的请求参数
// @param userID string 用户ID
// @param provider string MFA提供商 (可选, 默认:google)
type MFARecoverySecretRequest struct {
	UserID string `json:"user_id" form:"userId" binding:"required,min=1,max=100"`
}

func (req MFARecoverySecretRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 检查生成MFA密钥请求参数是否合法
// @return bool 是否合法
func (req MFARecoverySecretRequest) IsLegal() bool {
	if stringer.IsBlank(req.UserID) {
		return false
	}
	return true
}

// MFAVerifyCodeRequest 验证MFA代码请求
// @description 验证MFA代码的请求参数
// @param userID string 用户ID
// @param secret string MFA密钥
// @param code string MFA代码
type MFAVerifyCodeRequest struct {
	UserID string `json:"user_id" form:"userId" binding:"required,min=1,max=100"`
	Code   string `json:"code" form:"code" binding:"required,len=6,numeric"`
}

func (req MFAVerifyCodeRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 检查MFA代码请求参数是否合法
// @return bool 是否合法
func (req MFAVerifyCodeRequest) IsLegal() bool {
	if stringer.IsBlank(req.Code) ||
		stringer.IsBlank(req.UserID) {
		return false
	}
	return true
}

type MFAQrcodeRequest struct {
	UserID string `json:"user_id" form:"userId" binding:"required,min=1,max=100"`
}

func (req MFAQrcodeRequest) IsLegal() bool {
	if stringer.IsBlank(req.UserID) {
		return false
	}
	return true
}

func (req MFAQrcodeRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
