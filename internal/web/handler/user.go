package handler

import (
	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http/binding"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
)

type UserHandler struct {
	*Handler
	userService *service.UserService
}

func NewUserHandler(baseHandler *Handler, userService *service.UserService) *UserHandler {
	return &UserHandler{
		Handler:     baseHandler,
		userService: userService,
	}
}

func (u *UserHandler) RegisterUserHandler(ctx *gin.Context) {
	u.logger.Debugf("register user handler...")
	req := new(request.UserRegisterRequest)

	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		u.logger.Errorf("bind user register request failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidParameter, "绑定请求参数失败")
		return
	}

	u.userService.RegisterUser(ctx, req)
}

// LoginUserHandler 处理用户登录请求
func (u *UserHandler) LoginUserHandler(ctx *gin.Context) {
	u.logger.Debugf("login user handler...")
	req := new(request.UserLoginRequest)

	if err := binding.BindJSON(ctx, req); err != nil {
		u.logger.Errorf("bind user login request failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.InvalidParameter, "绑定请求参数失败")
		return
	}

	u.userService.LoginUser(ctx, req)
}
