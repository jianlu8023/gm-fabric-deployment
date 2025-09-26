package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/captcha"
)

// CaptchaService 验证码服务
// @description 处理验证码生成和验证的服务
// @struct
// @property *Service 基础服务
// @property captchaControl *captcha.Control 验证码控制器
type CaptchaService struct {
	*Service
	captchaControl *captcha.Control
}

// NewCaptchaService 创建验证码服务
// @description 创建并返回一个新的验证码服务实例
// @param baseService *Service 基础服务
// @param captchaControl *captcha.Control 验证码控制器
// @return *CaptchaService 验证码服务实例
func NewCaptchaService(baseService *Service, captchaControl *captcha.Control) *CaptchaService {
	return &CaptchaService{
		Service:        baseService,
		captchaControl: captchaControl,
	}
}

// GenerateCaptcha 生成验证码
// @description 生成验证码图片并返回验证码ID和Base64编码的图片数据
// @tags 验证码
// @router /api/captcha/generate [post]
// @param req body request.CaptchaGenerateRequest true "生成验证码请求参数"
// @success 200 {object} response.CaptchaGenerateResponse "验证码生成成功"
// @failure 400 {object} response.ErrorResponse "参数错误"
// @failure 500 {object} response.ErrorResponse "验证码生成失败"
// @param ctx *gin.Context Gin上下文
// @param req *request.CaptchaGenerateRequest 生成验证码请求参数
func (s *CaptchaService) GenerateCaptcha(ctx *gin.Context, req *request.CaptchaGenerateRequest) {
	s.logger.Debugf("received captcha generate request with params: %v", req)

	id, b64s, _, err := s.captchaControl.GenerateCaptcha()
	if err != nil {
		s.logger.Errorf("generate captcha failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.CaptchaGenerationFailed, commonhttp.ErrMsgCaptchaGenerationFailed)
		return
	}

	// 返回响应
	commonhttp.SuccessResponse(ctx, response.NewCaptchaGenerateResponse(id, b64s))
}

// ValidateCaptcha 验证验证码
// @description 验证用户输入的验证码是否正确有效
// @tags 验证码
// @router /api/captcha/validate [post]
// @param req body request.CaptchaValidateRequest true "验证码验证请求参数"
// @success 200 {object} response.CaptchaValidateResponse "验证码验证成功"
// @failure 400 {object} response.ErrorResponse "参数错误"
// @failure 401 {object} response.ErrorResponse "验证码不正确或已过期"
// @param ctx *gin.Context Gin上下文
// @param req *request.CaptchaValidateRequest 验证码验证请求参数
func (s *CaptchaService) ValidateCaptcha(ctx *gin.Context, req *request.CaptchaValidateRequest) {
	s.logger.Debugf("received captcha validate request with params: %v", req)

	// 验证验证码
	if ok, err := s.captchaControl.ValidateCaptcha(req.CaptchaId, req.Code); err == nil {
		if !ok {
			s.logger.Errorf("validate captcha failed: %s", req.CaptchaId)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidCaptcha, "验证码不正确或已过期")
			return
		}
	} else {
		// 发生错误
		s.logger.Errorf("validate captcha has error: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidCaptcha, commonhttp.ErrMsgInvalidCaptcha)
		return
	}

	// 验证成功
	s.logger.Debugf("validate captcha success: %s", req.CaptchaId)
	commonhttp.SuccessResponse(ctx, response.NewCaptchaValidateResponse(true, req.CaptchaId))
}
