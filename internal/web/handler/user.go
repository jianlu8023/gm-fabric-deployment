package handler

import (
	"net/http"

	"github.com/jianlu8023/golang-example/pkg/common/http/binding"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// UserHandler 用户处理器结构体
//
// @description 处理用户相关的HTTP请求
// @struct
type UserHandler struct {
	// Handler 基础处理器，提供日志功能
	*Handler
	// service 用户服务，处理用户相关的业务逻辑
	service *service.UserService
}

// NewUserHandler 创建用户处理器
//
// @param baseHandler *Handler 基础处理器
// @param service *service.UserService 用户服务
// @return *UserHandler 用户处理器实例
func NewUserHandler(baseHandler *Handler, userService *service.UserService) *UserHandler {
	return &UserHandler{
		Handler: baseHandler,
		service: userService,
	}
}

// RegisterUserHandler 用户注册处理函数
//
// @description 处理用户注册的HTTP请求
// @method POST
// @url /api/v1/user/register
// @param username string 用户名 (必需)
// @param password string 密码 (必需)
// @param confirmPassword string 确认密码 (必需)
// @param email string 邮箱 (必需)
// @param phone string 手机号 (可选)
// @return JSON 注册结果
func (h *UserHandler) RegisterUserHandler(ctx *gin.Context) {
	h.logger.Debugf("received register user handler...")

	req := new(request.UserRegisterRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		h.logger.Errorf("bind user register request failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "绑定请求参数失败")
		return
	}
	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("register user request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.RegisterUser(ctx, req)
}

// LoginUserHandler 用户登录处理函数
//
// @description 处理用户登录的HTTP请求
// @method POST
// @url /api/v1/user/login
// @param username string 用户名 (必需)
// @param password string 密码 (必需)
// @param captchaId string 验证码ID (必需)
// @param code string 验证码 (必需)
// @return JSON 登录结果和JWT令牌
func (h *UserHandler) LoginUserHandler(ctx *gin.Context) {
	h.logger.Debugf("received login user handler...")
	req := new(request.UserLoginRequest)

	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		h.logger.Errorf("bind user login request failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "绑定请求参数失败")
		return
	}
	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("login user request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.LoginUser(ctx, req)
}

// Routers 获取用户相关路由列表
//
// @return []commonhttp.RouterHandler 用户路由处理器列表
func (h *UserHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "registerUser",
			Uri:             "user/register",
			Method:          http.MethodPost,
			HandlerFunc:     h.RegisterUserHandler,
			Enabled:         true,
			Desc:            "注册新用户",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "loginUser",
			Uri:             "user/login",
			Method:          http.MethodPost,
			HandlerFunc:     h.LoginUserHandler,
			Enabled:         true,
			Desc:            "用户登录",
			EnableJWtVerify: false,
		},
	}
}
