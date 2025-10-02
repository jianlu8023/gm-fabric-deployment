package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
)

// SSEHandler SSE处理器结构体
//
// @description 处理Server-Sent Events相关的HTTP请求
// @struct
type SSEHandler struct {
	// Handler 基础处理器，提供日志功能
	*Handler
	// sseService SSE服务，处理SSE相关的业务逻辑
	sseService *service.SSEService
}

// NewSSEHandler 创建SSE处理器
//
// @param baseHandler *Handler 基础处理器
// @param sseService *service.SSEService SSE服务
// @return *SSEHandler SSE处理器实例
func NewSSEHandler(baseHandler *Handler, sseService *service.SSEService) *SSEHandler {
	return &SSEHandler{
		Handler:    baseHandler,
		sseService: sseService,
	}
}

// SSE 处理SSE连接的函数
//
// @description 建立Server-Sent Events连接，用于服务器向客户端推送实时消息
// @method GET
// @url /api/v1/sse
// @return text/event-stream SSE事件流
func (h *SSEHandler) SSE(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "sseHandler", "sse",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received sse handler...")
	h.sseService.SSE(ctx)
}

// Routers 获取SSE相关路由列表
//
// @return []commonhttp.RouterHandler SSE路由处理器列表
func (h *SSEHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:        "sse",
			Uri:         "/sse",
			Method:      http.MethodGet,
			HandlerFunc: h.SSE,
			Enabled:     true,
			Desc:        "sse",
		},
	}
}
