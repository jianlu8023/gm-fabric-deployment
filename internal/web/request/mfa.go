package request

import (
	"errors"
	
	"github.com/go-playground/validator/v10"
)

// MFAGenerateSecretRequest 生成MFA密钥请求
// @description 生成MFA密钥的请求参数
// @param userID string 用户ID
// @param provider string MFA提供商 (可选, 默认:google)
type MFAGenerateSecretRequest struct {
	UserID   string `json:"user_id" form:"user_id" binding:"required,min=1,max=100"`
	Provider string `json:"provider" form:"provider" binding:"omitempty,oneof=google microsoft"`
}

// Validate 验证生成MFA密钥请求参数
// @return error 验证错误
func (req *MFAGenerateSecretRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return err
	}
	// 如果未指定提供商，默认使用google
	if req.Provider == "" {
		req.Provider = "google"
	}
	return nil
}

// IsLegal 检查生成MFA密钥请求参数是否合法
// @return bool 是否合法
func (req *MFAGenerateSecretRequest) IsLegal() bool {
	return req.Validate() == nil
}

// MFAVerifyCodeRequest 验证MFA代码请求
// @description 验证MFA代码的请求参数
// @param userID string 用户ID
// @param secret string MFA密钥
// @param code string MFA代码
type MFAVerifyCodeRequest struct {
	UserID string `json:"user_id" form:"user_id" binding:"required,min=1,max=100"`
	Secret string `json:"secret" form:"secret" binding:"required,min=16,max=100"`
	Code   string `json:"code" form:"code" binding:"required,len=6,numeric"`
}

// Validate 验证MFA代码请求参数
// @return error 验证错误
func (req *MFAVerifyCodeRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return err
	}
	// 确保code是纯数字
	for _, char := range req.Code {
		if char < '0' || char > '9' {
			return errors.New("code must be numeric")
		}
	}
	return nil
}

// IsLegal 检查MFA代码请求参数是否合法
// @return bool 是否合法
func (req *MFAVerifyCodeRequest) IsLegal() bool {
	return req.Validate() == nil
}
