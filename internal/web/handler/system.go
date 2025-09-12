package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
)

// SystemHandler 系统相关接口
type SystemHandler struct {
	*Handler
	service *service.SystemService
}

// NewSystemHandler 创建系统相关接口
// @param handler *Handler 基础handler
// @param service *service.SystemService 系统相关服务
// @return *SystemHandler 系统相关接口
func NewSystemHandler(handler *Handler, service *service.SystemService) *SystemHandler {
	return &SystemHandler{
		Handler: handler,
		service: service,
	}
}

// GetSystemOverview 获取系统概览
// @param ctx *gin.Context 上下文
func (h *SystemHandler) GetSystemOverview(ctx *gin.Context) {
	h.logger.Debugf("get system overview handler...")
	h.service.GetSystemOverview(ctx)
}
