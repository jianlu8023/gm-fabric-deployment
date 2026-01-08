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

// UserHandler 用户处理器结构体
//
// @description 处理用户相关的HTTP请求
// @struct
type UserHandler struct {
	*Handler                     // Handler 基础处理器，提供日志功能
	service  service.UserService // service 用户服务，处理用户相关的业务逻辑
}

// NewUserHandler 创建用户处理器
//
// @param baseHandler *Handler 基础处理器
// @param service service.UserService 用户服务
// @return *UserHandler 用户处理器实例
func NewUserHandler(baseHandler *Handler, userService service.UserService) *UserHandler {
	return &UserHandler{
		Handler: baseHandler,
		service: userService,
	}
}

// RegisterUserHandler 用户注册处理函数
//
// @description 处理用户注册的HTTP请求，从请求体中解析用户注册信息，调用service方法处理业务逻辑，并通过HTTP响应返回结果
// @method POST
// @url /user/register
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *UserHandler) RegisterUserHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "userHandler", "registerUserHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received register user handler...")

	req := new(request.UserRegisterRequest)
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
		h.logger.Errorf("register user request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.RegisterUser(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// LoginUserHandler 用户登录处理函数
//
// @description 处理用户登录的HTTP请求，从请求体中解析用户登录信息，调用service方法处理业务逻辑，并通过HTTP响应返回结果
// @method POST
// @url /user/login
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *UserHandler) LoginUserHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "userHandler", "loginUserHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received login user handler...")

	req := new(request.UserLoginRequest)
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
		h.logger.Errorf("login user request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.LoginUser(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取用户相关路由列表
//
// @description 返回所有用户相关的HTTP路由配置，包括路由名称、URI、请求方法、处理函数、描述等信息
// @return []commonhttp.RouterHandler 用户路由处理器列表
func (h *UserHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "user",
			Uri:             "/user/",
			Method:          http.MethodPost,
			HandlerFunc:     h.pingHandler,
			Enabled:         true,
			Desc:            "用户模块",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "user",
			Uri:             "/user/",
			Method:          http.MethodGet,
			HandlerFunc:     h.pingHandler,
			Enabled:         true,
			Desc:            "用户模块",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "registerUser",
			Uri:             "/user/register",
			Method:          http.MethodPost,
			HandlerFunc:     h.RegisterUserHandler,
			Enabled:         true,
			Desc:            "注册新用户",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "loginUser",
			Uri:             "/user/login",
			Method:          http.MethodPost,
			HandlerFunc:     h.LoginUserHandler,
			Enabled:         true,
			Desc:            "用户登录",
			EnableJWtVerify: false,
		},
	}
}
