package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// CaptchaHandler 验证码处理器结构体
//
// @description 处理验证码相关的HTTP请求
// @struct
type CaptchaHandler struct {
	*Handler                         // Handler 基础处理器，提供日志功能
	service  *service.CaptchaService // service 验证码服务，处理验证码相关的业务逻辑
}

// NewCaptchaHandler 创建验证码处理器
//
// @description 创建并返回一个新的验证码处理器实例
// @param baseHandler *Handler 基础处理器
// @param captchaService *service.CaptchaService 验证码服务
// @return *CaptchaHandler 验证码处理器实例
func NewCaptchaHandler(baseHandler *Handler, captchaService *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{
		Handler: baseHandler,
		service: captchaService,
	}
}

// CaptchaServiceInterface 验证码服务接口
//
// @description 定义验证码服务需要实现的方法
// @interface
type CaptchaServiceInterface interface {
	// GenerateCaptcha 生成验证码
	//
	// @param ctx *gin.Context Gin上下文
	// @param req *request.CaptchaGenerateRequest 验证码生成请求
	GenerateCaptcha(ctx *gin.Context, req *request.CaptchaGenerateRequest)
	// ValidateCaptcha 验证验证码
	//
	// @param ctx *gin.Context Gin上下文
	// @param req *request.CaptchaValidateRequest 验证码验证请求
	ValidateCaptcha(ctx *gin.Context, req *request.CaptchaValidateRequest)
}

// GenerateCaptchaHandler 生成验证码的处理函数
//
// @description 处理生成验证码的HTTP请求
// @method GET
// @url /captcha/generate
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *CaptchaHandler) GenerateCaptchaHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "captchaHandler", "generateCaptchaHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received captcha generate handler...")

	// 绑定请求参数
	req := new(request.CaptchaGenerateRequest)
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
		h.logger.Errorf("captcha generate request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层生成验证码
	h.service.GenerateCaptcha(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// ValidateCaptchaHandler 验证验证码的处理函数
//
// @description 处理验证验证码的HTTP请求
// @method POST
// @url /captcha/validate
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *CaptchaHandler) ValidateCaptchaHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "captchaHandler", "validateCaptchaHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received captcha validate handler...")

	// 绑定请求参数
	req := new(request.CaptchaValidateRequest)
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
		h.logger.Errorf("captcha validate request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层验证验证码
	h.service.ValidateCaptcha(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// RefreshCaptchaHandler 刷新验证码的处理函数
//
// @description 处理刷新验证码的HTTP请求
// @method GET
// @url /captcha/refresh
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *CaptchaHandler) RefreshCaptchaHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "captchaHandler", "refreshCaptchaHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received captcha refresh handler...")

	// 绑定请求参数
	req := new(request.CaptchaGenerateRequest)
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
		h.logger.Errorf("captcha refresh request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层生成新的验证码
	h.service.GenerateCaptcha(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 注册验证码相关路由
//
// @description 注册所有验证码相关的HTTP路由
// @return []commonhttp.RouterHandler 验证码路由处理器列表
func (h *CaptchaHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		// 验证码相关路由
		&commonhttp.MyRouter{
			Name:            "generateCaptcha",
			Uri:             "captcha/generate",
			Method:          http.MethodGet,
			HandlerFunc:     h.GenerateCaptchaHandler,
			Enabled:         true,
			Desc:            "生成验证码",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "validateCaptcha",
			Uri:             "captcha/validate",
			Method:          http.MethodPost,
			HandlerFunc:     h.ValidateCaptchaHandler,
			Enabled:         true,
			Desc:            "验证验证码",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "refreshCaptcha",
			Uri:             "captcha/refresh",
			Method:          http.MethodGet,
			HandlerFunc:     h.RefreshCaptchaHandler,
			Enabled:         true,
			Desc:            "刷新验证码",
			EnableJWtVerify: false,
		},
	}
}
