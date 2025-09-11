package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
)

type SSEHandler struct {
	*Handler
	sseService *service.SSEService
}

func NewSSEHandler(baseHandler *Handler, sseService *service.SSEService) *SSEHandler {
	return &SSEHandler{
		Handler:    baseHandler,
		sseService: sseService,
	}
}

func (h *SSEHandler) SSE(ctx *gin.Context) {
	h.logger.Debugf("starting sse handler...")
	h.sseService.SSE(ctx)
}
