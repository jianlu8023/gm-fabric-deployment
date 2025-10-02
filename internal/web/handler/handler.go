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
	// logger 日志记录器，用于记录处理器的运行日志
	logger *zap.SugaredLogger
}

// NewHandler 创建基础处理器
//
// @param logger *zap.SugaredLogger 日志记录器实例
// @return *Handler 基础处理器实例
func NewHandler(logger *zap.SugaredLogger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// Routers 获取基础路由列表
//
// @return []commonhttp.RouterHandler 基础路由处理器列表
func (h *Handler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:   "pingGet",
			Uri:    "/ping",
			Method: http.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				_, span := tracer.StartSpan(ctx.Request.Context(), "baseHandler", "ping",
					attribute.String("requestId", requestid.Get(ctx)),
				)
				defer span.End()

				h.logger.Debugf("received ping handler...")
				commonhttp.SuccessResponse(ctx, gin.H{
					"request_id": requestid.Get(ctx),
					"message":    "pong",
				})
			},
			Enabled: true,
			Desc:    "ping的请求",
		},
		&commonhttp.MyRouter{
			Name:   "pingPost",
			Uri:    "/ping",
			Method: http.MethodPost,
			HandlerFunc: func(ctx *gin.Context) {
				_, span := tracer.StartSpan(ctx.Request.Context(), "baseHandler", "ping",
					attribute.String("requestId", requestid.Get(ctx)),
				)
				defer span.End()

				h.logger.Debugf("received ping handler...")
				commonhttp.SuccessResponse(ctx, gin.H{
					"request_id": requestid.Get(ctx),
					"message":    "pong",
				})
			},
			Enabled: true,
			Desc:    "ping的请求",
		},
		&commonhttp.MyRouter{
			Name:   "health",
			Uri:    "/health",
			Method: http.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				_, span := tracer.StartSpan(ctx.Request.Context(), "baseHandler", "health",
					attribute.String("requestId", requestid.Get(ctx)),
				)
				defer span.End()

				h.logger.Debugf("received health handler...")
				commonhttp.SuccessResponse(ctx, "ok")
			},
			Enabled: true,
			Desc:    "检查系统是否正常对外服务",
		},
	}

}
