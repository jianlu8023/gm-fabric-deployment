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

// DockerNetworkHandler Docker网络处理器结构体
//
// @description 处理Docker网络相关的HTTP请求
// @struct
type DockerNetworkHandler struct {
	*Handler                              // Handler 基础处理器，提供日志功能
	service  service.DockerNetworkService // service Docker网络服务，处理Docker网络相关的业务逻辑
}

// NewDockerNetworkHandler 创建Docker网络处理器
//
// @description 创建并返回一个新的Docker网络处理器实例
// @param baseHandler *Handler 基础handler
// @param service service.DockerNetworkService Docker网络相关服务
// @return *DockerNetworkHandler Docker网络处理器实例
func NewDockerNetworkHandler(baseHandler *Handler,
	service service.DockerNetworkService,
) *DockerNetworkHandler {
	return &DockerNetworkHandler{
		Handler: baseHandler,
		service: service,
	}
}

// DockerNetworkList 获取Docker网络列表处理函数
//
// @description 获取Docker网络列表信息
// @method GET
// @url /docker/network/list
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *DockerNetworkHandler) DockerNetworkList(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "dockerNetworkHandler", "dockerNetworkList",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received docker network list handler...")
	req := new(request.DockerNetworkListRequest)
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
		h.logger.Errorf("docker network list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.DockerNetworkList(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取Docker网络相关路由列表
//
// @description 返回Docker网络相关的所有HTTP路由定义
// @return []commonhttp.RouterHandler Docker网络路由处理器列表
func (h *DockerNetworkHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:        "DockerNetworkList",
			Method:      http.MethodGet,
			Uri:         "docker/network/list",
			HandlerFunc: h.DockerNetworkList,
			Enabled:     true,
			EnableAuth:  false,
			Desc:        "获取docker网络列表",
		},
	}
}
