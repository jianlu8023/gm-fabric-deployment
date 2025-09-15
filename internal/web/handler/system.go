package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
)

// SystemHandler 系统处理器结构体
//
// @description 处理系统相关的HTTP请求
// @struct
type SystemHandler struct {
	// Handler 基础处理器，提供日志功能
	*Handler
	// service 系统服务，处理系统相关的业务逻辑
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

// GetSystemOverview 获取系统概览信息
//
// @description 获取系统的整体运行状态和关键指标
// @method GET
// @url /api/v1/system/overview
// @return JSON 系统概览信息，包括节点数量、容器状态、资源使用情况等
func (h *SystemHandler) GetSystemOverview(ctx *gin.Context) {
	h.logger.Debugf("received system overview handler...")
	h.service.GetSystemOverview(ctx)
}

// Routers 获取系统相关路由列表
//
// @return []commonhttp.RouterHandler 系统路由处理器列表
func (h *SystemHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "systemOverview",
			Uri:             "system/overview",
			Enabled:         true,
			EnableJWtVerify: false,
			Method:          http.MethodGet,
			Desc:            "获取系统的总览",
			HandlerFunc:     h.GetSystemOverview,
		},
	}
}
