package service

import (
	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/response"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/captcha"
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
// @param req body request.GenerateCaptchaRequest true "生成验证码请求参数"
// @success 200 {object} response.CaptchaResponse "验证码生成成功"
// @failure 400 {object} response.ErrorResponse "参数错误"
// @failure 500 {object} response.ErrorResponse "验证码生成失败"
// @param ctx *gin.Context Gin上下文
// @param req *request.GenerateCaptchaRequest 生成验证码请求参数
func (s *CaptchaService) GenerateCaptcha(ctx *gin.Context, req *request.GenerateCaptchaRequest) {
	s.logger.Debugf("接收到生成验证码请求: %v", req)

	// 验证请求参数
	if !req.IsLegal() {
		s.logger.Errorf("生成验证码请求参数不合法: %v", req)
		webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidParameter, "生成验证码参数不合法")
		return
	}

	id, b64s, answer, err := s.captchaControl.GenerateCaptcha()
	if err != nil {
		s.logger.Errorf("生成验证码失败: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.CaptchaGenerationFailed, webhttp.ErrMsgCaptchaGenerationFailed)
		return
	}

	var _ = answer

	// 返回响应
	resp := response.NewCaptchaResponse(id, b64s, "", "success", "验证码生成成功")
	webhttp.SuccessResponse(ctx, resp)
}

// ValidateCaptcha 验证验证码
// @description 验证用户输入的验证码是否正确有效
// @tags 验证码
// @router /api/captcha/validate [post]
// @param req body request.CaptchaRequest true "验证码验证请求参数"
// @success 200 {object} response.ValidateCaptchaResponse "验证码验证成功"
// @failure 400 {object} response.ErrorResponse "参数错误"
// @failure 401 {object} response.ErrorResponse "验证码不正确或已过期"
// @param ctx *gin.Context Gin上下文
// @param req *request.CaptchaRequest 验证码验证请求参数
func (s *CaptchaService) ValidateCaptcha(ctx *gin.Context, req *request.CaptchaRequest) {
	s.logger.Debugf("接收到验证码验证请求: %v", req)

	// 验证请求参数
	if !req.IsLegal() {
		s.logger.Errorf("验证码验证请求参数不合法: %v", req)
		webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidParameter, "验证码验证参数不合法")
		return
	}

	// 验证验证码ID是否为空
	if req.CaptchaId == "" {
		s.logger.Errorf("验证码ID不能为空")
		webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidParameter, "验证码ID不能为空")
		return
	}

	// 验证验证码
	if ok, err := s.captchaControl.ValidateCaptcha(req.CaptchaId, req.Code); err == nil {
		if !ok {
			s.logger.Errorf("验证码验证失败: %s", req.CaptchaId)
			webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidCaptcha, "验证码不正确或已过期")
			return
		}
	} else {
		// 发生错误
		s.logger.Errorf("验证码验证过程中发生错误: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidCaptcha, webhttp.ErrMsgInvalidCaptcha)
		return
	}

	// 验证成功
	s.logger.Debugf("验证码验证成功: %s", req.CaptchaId)
	resp := response.NewValidateCaptchaResponse(true, "验证码验证成功", req.CaptchaId)
	webhttp.SuccessResponse(ctx, resp)
}

// GetCaptchaImageByType 根据类型获取验证码图片
// @description 快捷方法，根据类型获取验证码图片
// @param captchaType 验证码类型
// @return (string, string, error) 验证码ID、Base64图片、错误
// func (s *CaptchaService) GetCaptchaImageByType(captchaType string) (string, string, error) {
// 	driver := s.getDriver(captchaType, 240, 80)
// 	captcha := base64Captcha.NewCaptcha(driver, s.store)
// 	return captcha.Generate()
// }
