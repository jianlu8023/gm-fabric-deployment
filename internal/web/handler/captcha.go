package handler

import (
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// CaptchaHandler 验证码处理器
// @description 处理验证码相关的HTTP请求
// @property *Handler 基础处理器
// @property captchaService *service.CaptchaService 验证码服务
type CaptchaHandler struct {
	*Handler
	captchaService *service.CaptchaService
}

// NewCaptchaHandler 创建验证码处理器
// @param baseHandler 基础处理器
// @param captchaService 验证码服务
// @return *CaptchaHandler 验证码处理器实例
func NewCaptchaHandler(baseHandler *Handler, captchaService *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{
		Handler:        baseHandler,
		captchaService: captchaService,
	}
}

// GenerateCaptchaHandler 生成验证码的处理函数
// @description 处理生成验证码的HTTP请求
// @method GET
// @url /api/v1/captcha/generate
// @param captchaType string 验证码类型 (可选, 默认:string)
// @param width int 验证码图片宽度 (可选, 默认:240)
// @param height int 验证码图片高度 (可选, 默认:80)
// @return JSON 验证码信息和图片
func (c *CaptchaHandler) GenerateCaptchaHandler(ctx *gin.Context) {
	c.logger.Debugf("生成验证码处理函数被调用")

	// 绑定请求参数
	req := new(request.GenerateCaptchaRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		c.logger.Errorf("绑定生成验证码请求参数失败: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "绑定生成验证码参数失败")
		return
	}

	// 调用服务层生成验证码
	c.captchaService.GenerateCaptcha(ctx, req)
}

// ValidateCaptchaHandler 验证验证码的处理函数
// @description 处理验证验证码的HTTP请求
// @method POST
// @url /api/v1/captcha/validate
// @param captchaId string 验证码ID (必需)
// @param code string 用户输入的验证码 (必需)
// @return JSON 验证结果
func (c *CaptchaHandler) ValidateCaptchaHandler(ctx *gin.Context) {
	c.logger.Debugf("验证验证码处理函数被调用")

	// 绑定请求参数
	req := new(request.CaptchaRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		c.logger.Errorf("绑定验证验证码请求参数失败: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "绑定验证验证码参数失败")
		return
	}

	// 调用服务层验证验证码
	c.captchaService.ValidateCaptcha(ctx, req)
}

// RefreshCaptchaHandler 刷新验证码的处理函数
// @description 处理刷新验证码的HTTP请求
// @method GET
// @url /api/v1/captcha/refresh
// @param captchaType string 验证码类型 (可选, 默认:string)
// @param width int 验证码图片宽度 (可选, 默认:240)
// @param height int 验证码图片高度 (可选, 默认:80)
// @return JSON 新的验证码信息和图片
func (c *CaptchaHandler) RefreshCaptchaHandler(ctx *gin.Context) {
	c.logger.Debugf("刷新验证码处理函数被调用")

	// 绑定请求参数
	req := new(request.GenerateCaptchaRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		c.logger.Errorf("绑定刷新验证码请求参数失败: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "绑定刷新验证码参数失败")
		return
	}

	// 调用服务层生成新的验证码
	c.captchaService.GenerateCaptcha(ctx, req)
}

// Routers 注册验证码相关路由
// @description 注册所有验证码相关的HTTP路由
func (c *CaptchaHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		// 验证码相关路由
		&commonhttp.MyRouter{
			Name:            "generateCaptcha",
			Uri:             "captcha/generate",
			Method:          http.MethodGet,
			HandlerFunc:     c.GenerateCaptchaHandler,
			Enabled:         true,
			Desc:            "生成验证码",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "validateCaptcha",
			Uri:             "captcha/validate",
			Method:          http.MethodPost,
			HandlerFunc:     c.ValidateCaptchaHandler,
			Enabled:         true,
			Desc:            "验证验证码",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "refreshCaptcha",
			Uri:             "captcha/refresh",
			Method:          http.MethodGet,
			HandlerFunc:     c.RefreshCaptchaHandler,
			Enabled:         true,
			Desc:            "刷新验证码",
			EnableJWtVerify: false,
		},
	}
}
