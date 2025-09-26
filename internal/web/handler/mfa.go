package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// MFAHandler MFA处理器
// @description 处理MFA相关的HTTP请求
// @struct
// @property Handler 基础处理器
// @property service *service.MFAService MFA服务
type MFAHandler struct {
	*Handler
	service *service.MFAService
}

func NewMFAHandler(baseHandler *Handler, mfaService *service.MFAService) *MFAHandler {
	return &MFAHandler{
		Handler: baseHandler,
		service: mfaService,
	}
}

func (h *MFAHandler) GenerateMfaSecret(ctx *gin.Context) {
	h.logger.Debugf("received generate MFA secret request")
	var req request.MFAGenerateSecretRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", "error", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "请求参数无效: "+err.Error())
		return
	}

	// 调用服务层生成MFA密钥
	resp, err := h.service.GenerateSecret(&req)
	if err != nil {
		h.logger.Error("Failed to generate MFA secret", "error", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "生成MFA密钥失败: "+err.Error())
		return
	}

	commonhttp.SuccessResponse(ctx, resp)
}

func (h *MFAHandler) VerifyMfaCode(ctx *gin.Context) {
	h.logger.Debugf("received verify MFA code request")
	var req request.MFAVerifyCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", "error", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "请求参数无效: "+err.Error())
		return
	}

	// 调用服务层验证MFA代码
	resp, err := h.service.VerifyCode(&req)
	if err != nil {
		h.logger.Error("Failed to verify MFA code", "error", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "验证MFA代码失败: "+err.Error())
		return
	}

	commonhttp.SuccessResponse(ctx, resp)

}

func (h *MFAHandler) GenerateQrCode(ctx *gin.Context) {
	h.logger.Debugf("received get QR code image request")
	qrCodeURL := ctx.Query("qr_code_url")
	if qrCodeURL == "" {
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "qr_code_url参数不能为空")
		return
	}

	// 调用服务层获取二维码图片
	qrCodeImage, err := h.service.GetQrCodeImage(qrCodeURL)
	if err != nil {
		h.logger.Error("Failed to get QR code image", "error", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取二维码图片失败: "+err.Error())
		return
	}

	commonhttp.SuccessResponse(ctx, gin.H{"qr_code_image": qrCodeImage})
}

// Routers 获取MFA相关的路由列表
// @description 获取所有MFA相关的HTTP路由
// @return []commonhttp.RouterHandler 路由处理器列表
func (h *MFAHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "GenerateMFASecret",
			Uri:             "/mfa/secret",
			Method:          http.MethodPost,
			HandlerFunc:     h.GenerateMfaSecret,
			Enabled:         true,
			Desc:            "生成MFA密钥和二维码",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "VerifyMFACode",
			Uri:             "mfa/verify",
			Method:          http.MethodPost,
			HandlerFunc:     h.VerifyMfaCode,
			Enabled:         true,
			Desc:            "验证MFA代码",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:        "GetQRCodeImage",
			Uri:         "/mfa/qrcode",
			Method:      http.MethodGet,
			HandlerFunc: h.GenerateQrCode,
			Enabled:     true,
			Desc:        "获取MFA二维码图片",
		},
	}
}
