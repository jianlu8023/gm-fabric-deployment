package handler

import (
	"net/http"
	"strings"

	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// CaptchaHandler 验证码处理器
// @description 处理验证码相关的HTTP请求
// @property *Handler 基础处理器
// @property service *service.CaptchaService 验证码服务
type CaptchaHandler struct {
	*Handler
	service *service.CaptchaService
}

// NewCaptchaHandler 创建验证码处理器
// @param baseHandler 基础处理器
// @param service 验证码服务
// @return *CaptchaHandler 验证码处理器实例
func NewCaptchaHandler(baseHandler *Handler, captchaService *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{
		Handler: baseHandler,
		service: captchaService,
	}
}

type CaptchaServiceInterface interface {
	GenerateCaptcha(ctx *gin.Context, req *request.CaptchaGenerateRequest)
	ValidateCaptcha(ctx *gin.Context, req *request.CaptchaValidateRequest)
}

// GenerateCaptchaHandler 生成验证码的处理函数
// @description 处理生成验证码的HTTP请求
// @method GET
// @url /api/v1/captcha/generate
// @param captchaType string 验证码类型 (可选, 默认:string)
// @param width int 验证码图片宽度 (可选, 默认:240)
// @param height int 验证码图片高度 (可选, 默认:80)
// @return JSON 验证码信息和图片
func (h *CaptchaHandler) GenerateCaptchaHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "captchaHandler", "generate")
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
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, strings.Join(msg, ","))
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("captcha generate request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层生成验证码
	h.service.GenerateCaptcha(ctx, req)
}

// ValidateCaptchaHandler 验证验证码的处理函数
// @description 处理验证验证码的HTTP请求
// @method POST
// @url /api/v1/captcha/validate
// @param captchaId string 验证码ID (必需)
// @param code string 用户输入的验证码 (必需)
// @return JSON 验证结果
func (h *CaptchaHandler) ValidateCaptchaHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "captchaHandler", "validate")
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
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, strings.Join(msg, ","))
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("captcha validate request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层验证验证码
	h.service.ValidateCaptcha(ctx, req)
}

// RefreshCaptchaHandler 刷新验证码的处理函数
// @description 处理刷新验证码的HTTP请求
// @method GET
// @url /api/v1/captcha/refresh
// @param captchaType string 验证码类型 (可选, 默认:string)
// @param width int 验证码图片宽度 (可选, 默认:240)
// @param height int 验证码图片高度 (可选, 默认:80)
// @return JSON 新的验证码信息和图片
func (h *CaptchaHandler) RefreshCaptchaHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "captchaHandler", "refresh")
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
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, strings.Join(msg, ","))
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("captcha refresh request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层生成新的验证码
	h.service.GenerateCaptcha(ctx, req)
}

// Routers 注册验证码相关路由
// @description 注册所有验证码相关的HTTP路由
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
