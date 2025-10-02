package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

func (h *MFAHandler) GenerateRecoverySecret(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "mfaHandler", "generateRecoverySecret",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received mfa generate secret handler...")
	req := new(request.MFARecoverySecretRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("mfa generate request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层生成MFA密钥
	h.service.GenerateRecoverySecret(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

func (h *MFAHandler) VerifyMfaCode(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "mfaHandler", "verifyMfaCode",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received mfa verify code handler...")
	req := new(request.MFAVerifyCodeRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("mfa verify code request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.VerifyCode(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

func (h *MFAHandler) GenerateQrCode(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "mfaHandler", "generateQrCode",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received mfa generate qrcode handler...")
	req := new(request.MFAQrcodeRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("mfa generate qrcode request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.GenerateQrCodeImage(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取MFA相关的路由列表
// @description 获取所有MFA相关的HTTP路由
// @return []commonhttp.RouterHandler 路由处理器列表
func (h *MFAHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "GenerateMFASecret",
			Uri:             "/mfa/recovery/secret",
			Method:          http.MethodPost,
			HandlerFunc:     h.GenerateRecoverySecret,
			Enabled:         true,
			Desc:            "生成MFA恢复密钥",
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
