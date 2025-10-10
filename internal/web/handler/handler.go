package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// Handler 基础处理器结构体
//
// @description 所有处理器的基础结构体，提供共享的日志功能
// @struct
type Handler struct {
	logger *zap.SugaredLogger // logger 日志记录器，用于记录处理器的运行日志
}

// NewHandler 创建基础处理器
//
// @description 创建并返回一个新的基础处理器实例
// @param logger *zap.SugaredLogger 日志记录器实例
// @return *Handler 基础处理器实例
func NewHandler(logger *zap.SugaredLogger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// pingHandler ping请求处理函数
//
// @description 处理ping请求，返回pong响应
// @method GET,POST
// @url /ping
// @param ctx *gin.Context Gin上下文
func (h *Handler) pingHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "baseHandler", "ping",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received ping handler...")
	commonhttp.SuccessResponse(ctx, gin.H{
		"request_id": requestid.Get(ctx),
		"message":    "pong",
	})
}

// healthHandler 健康检查处理函数
//
// @description 处理健康检查请求，返回ok响应
// @method GET
// @url /health
// @param ctx *gin.Context Gin上下文
func (h *Handler) healthHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "baseHandler", "health",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received health handler...")
	commonhttp.SuccessResponse(ctx, "ok")
}

// Routers 获取基础路由列表
//
// @description 获取并返回基础路由处理器列表，包括ping和健康检查路由
// @return []commonhttp.RouterHandler 基础路由处理器列表
func (h *Handler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:        "pingGet",
			Uri:         "/ping",
			Method:      http.MethodGet,
			HandlerFunc: h.pingHandler,
			Enabled:     true,
			Desc:        "ping的请求",
		},
		&commonhttp.MyRouter{
			Name:        "pingPost",
			Uri:         "/ping",
			Method:      http.MethodPost,
			HandlerFunc: h.pingHandler,
			Enabled:     true,
			Desc:        "ping的请求",
		},
		&commonhttp.MyRouter{
			Name:        "health",
			Uri:         "/health",
			Method:      http.MethodGet,
			HandlerFunc: h.healthHandler,
			Enabled:     true,
			Desc:        "检查系统是否正常对外服务",
		},
	}

}
