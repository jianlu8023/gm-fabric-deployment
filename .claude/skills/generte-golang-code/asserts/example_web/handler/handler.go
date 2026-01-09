package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// Handler 基础处理器
//
// struct
type Handler struct {
	logger *zap.SugaredLogger // logger 日志
}

// NewHandler 创建基础处理器
//
// @param logger *zap.SugaredLogger: 日志
//
// @return *Handler: 基础处理器
func NewHandler(logger *zap.SugaredLogger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// Ping ping路由handler
//
// @param ctx *gin.Context: gin上下文
func (h *Handler) Ping(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "handler", "ping")
	defer span.End()

	h.logger.Debugf("receive /ping request from %s", ctx.RemoteIP())

	commonhttp.SuccessResponse(ctx, "pong")
	span.SetStatus(codes.Ok, "ok")
	return

}

// Routers 路由
//
// @return []commonhttp.RouterHandler: 路由定义
func (h *Handler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "ping",
			Uri:             "/ping",
			Method:          http.MethodGet,
			HandlerFunc:     h.Ping,
			Enabled:         true,
			Desc:            "ping",
			EnableJWtVerify: false,
		},
	}
}
