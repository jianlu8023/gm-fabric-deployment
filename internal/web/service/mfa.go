package service

import (
	"errors"
	
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	"github.com/jianlu8023/golang-example/pkg/control/mfa"
)

// MFAService MFA服务
// @description MFA相关的业务服务
// @struct
// @property Service 基础服务
// @property mfaControl *mfa.Control MFA控制器

type MFAService struct {
	*Service
	mfaControl *mfa.Control
}

// NewMFAService 创建MFA服务实例
// @description 创建并返回一个新的MFA服务实例
// @param logger *zap.SugaredLogger 日志记录器
// @param mfaControl *mfa.Control MFA控制器
// @return *MFAService MFA服务实例
func NewMFAService(baseService *Service, mfaControl *mfa.Control) *MFAService {
	return &MFAService{
		Service:    baseService,
		mfaControl: mfaControl,
	}
}

// GenerateSecret 生成MFA密钥
// @description 为用户生成MFA密钥、二维码URL和二维码图片
// @param req *request.MFAGenerateSecretRequest 生成密钥请求
// @return *response.MFASecretResponse MFA密钥响应
// @return error 错误信息
func (s *MFAService) GenerateSecret(req *request.MFAGenerateSecretRequest) (*response.MFASecretResponse, error) {
	if !req.IsLegal() {
		s.logger.Error("Invalid MFAGenerateSecretRequest", "error", req.Validate())
		return nil, req.Validate()
	}

	// 使用MFA控制器生成密钥
	secret, qrCodeURL, err := s.mfaControl.GenerateSecret(req.UserID)
	if err != nil {
		s.logger.Error("Failed to generate MFA secret", "error", err, "userID", req.UserID, "provider", req.Provider)
		return nil, err
	}

	// 生成二维码图片
	qrCodeImage := s.mfaControl.GetQrCodeImage(qrCodeURL)

	// 创建并返回响应
	resp := response.NewMFASecretResponse(secret, qrCodeURL, qrCodeImage, req.Provider)
	s.logger.Info("MFA secret generated successfully", "userID", req.UserID, "provider", req.Provider)
	return resp, nil
}

// VerifyCode 验证MFA代码
// @description 验证用户提供的MFA代码是否有效
// @param req *request.MFAVerifyCodeRequest 验证代码请求
// @return *response.MFAVerifyResponse MFA验证响应
// @return error 错误信息
func (s *MFAService) VerifyCode(req *request.MFAVerifyCodeRequest) (*response.MFAVerifyResponse, error) {
	if !req.IsLegal() {
		s.logger.Error("Invalid MFAVerifyCodeRequest", "error", req.Validate())
		return nil, req.Validate()
	}

	// 验证MFA代码
	valid:= s.mfaControl.VerifyCode(req.Secret, req.Code)
	
	// 创建并返回响应
	resp := response.NewMFAVerifyResponse(valid, "")
	status := "failed"
	if valid {
		status = "success"
	}
	s.logger.Info("MFA code verification", "userID", req.UserID, "status", status)
	return resp, nil
}

// GetQrCodeImage 获取MFA二维码图片
// @description 根据二维码URL生成二维码图片
// @param qrCodeURL string 二维码URL
// @return string 二维码图片的Base64编码
// @return error 错误信息
func (s *MFAService) GetQrCodeImage(qrCodeURL string) (string, error) {
	if qrCodeURL == "" {
		s.logger.Error("Empty QR code URL")
		return "", errors.New("empty QR code URL")
	}

	// 使用MFA控制器获取二维码图片
	qrCodeImage := s.mfaControl.GetQrCodeImage(qrCodeURL)
	if qrCodeImage == "" {
		s.logger.Error("Failed to generate QR code image")
		return "", errors.New("failed to generate QR code image")
	}

	s.logger.Info("QR code image generated successfully")
	return qrCodeImage, nil
}
