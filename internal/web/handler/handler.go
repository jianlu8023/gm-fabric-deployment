package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
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
			Name:   "ping",
			Uri:    "/ping",
			Method: http.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				h.logger.Debugf("ping handler...")
				webhttp.SuccessResponse(ctx, gin.H{
					"request_id": requestid.Get(ctx),
					"message":    "pong",
				})
			},
			Enabled: true,
			Desc:    "ping",
		},
		&commonhttp.MyRouter{
			Name:   "health",
			Uri:    "/health",
			Method: http.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				h.logger.Debugf("health handler...")
				webhttp.SuccessResponse(ctx, "ok")
			},
			Enabled: true,
			Desc:    "health check",
		},
	}

}
