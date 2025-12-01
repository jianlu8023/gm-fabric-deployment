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

// DockerImageHandler Docker镜像处理器结构体
//
// @description 处理Docker镜像相关的HTTP请求
// @struct
type DockerImageHandler struct {
	*Handler                             // Handler 基础处理器，提供日志功能
	service  *service.DockerImageService // service Docker镜像服务，处理Docker镜像相关的业务逻辑
}

// NewDockerImageHandler 创建Docker镜像处理器
//
// @param baseHandler *Handler 基础处理器
// @param service *service.DockerImageService Docker镜像服务
// @return *DockerImageHandler Docker镜像处理器实例
func NewDockerImageHandler(baseHandler *Handler,
	service *service.DockerImageService,
) *DockerImageHandler {
	return &DockerImageHandler{
		Handler: baseHandler,
		service: service,
	}
}

// DockerImageServiceInterface Docker镜像服务接口
//
// @description 定义Docker镜像服务需要实现的方法
// @interface
type DockerImageServiceInterface interface {
	// DockerImageList 获取Docker镜像列表
	//
	// @param ctx *gin.Context Gin上下文
	// @param req *request.DockerImageListRequest Docker镜像列表请求参数
	DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest)
	// DockerImagePull 拉取Docker镜像
	//
	// @param ctx *gin.Context Gin上下文
	// @param req *request.DockerImagePullRequest Docker镜像拉取请求参数
	DockerImagePull(ctx *gin.Context, req *request.DockerImagePullRequest)
}

// DockerImageList 处理Docker镜像列表请求
//
// @description 处理获取Docker镜像列表的HTTP请求
// @method GET
// @url /docker/image/list
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *DockerImageHandler) DockerImageList(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "dockerImageHandler", "dockerImageList",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received docker image list handler...")

	req := new(request.DockerImageListRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages,
		)
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
		h.logger.Errorf("docker image list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.DockerImageList(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// DockerImagePull 处理Docker镜像拉取请求
//
// @description 处理拉取Docker镜像的HTTP请求
// @method POST
// @url /docker/image/pull
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *DockerImageHandler) DockerImagePull(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "dockerImageHandler", "dockerImagePull",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received docker image pull handler...")

	req := new(request.DockerImagePullRequest)
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
		h.logger.Errorf("docker image pull request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.DockerImagePull(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取Docker镜像相关路由列表
//
// @description 返回Docker镜像相关的所有HTTP路由定义
// @return []commonhttp.RouterHandler Docker镜像路由处理器列表
func (h *DockerImageHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "DockerImageList",
			Uri:             "docker/image/list",
			Method:          http.MethodGet,
			HandlerFunc:     h.DockerImageList,
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "获取docker镜像列表",
		},
		&commonhttp.MyRouter{
			Name:            "DockerImagePull",
			Uri:             "docker/image/pull",
			Method:          http.MethodPost,
			HandlerFunc:     h.DockerImagePull,
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "拉取docker镜像",
		},
	}
}
