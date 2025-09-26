package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
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

// NewSystemHandler 创建系统处理器
//
// @description 创建并返回一个新的系统处理器实例
// @param handler *Handler 基础handler
// @param service *service.SystemService 系统相关服务
// @return *SystemHandler 系统处理器实例
func NewSystemHandler(handler *Handler, service *service.SystemService) *SystemHandler {
	return &SystemHandler{
		Handler: handler,
		service: service,
	}
}

// GetSystemOverview 获取系统概览信息处理函数
//
// @description 获取系统的整体运行状态和关键指标
// @method GET
// @url /api/v1/system/overview
// @return JSON 系统概览信息，包括节点数量、容器状态、资源使用情况等
func (h *SystemHandler) GetSystemOverview(ctx *gin.Context) {
	h.logger.Debugf("received system overview handler...")

	// 调用服务层方法
	h.service.GetSystemOverview(ctx)
}

// GetSystemInitStatus 获取系统初始化状态处理函数
//
// @description 检查系统是否已经完成初始化配置
// @method GET
// @url /api/v1/system/init
// @return JSON 系统初始化状态信息
func (h *SystemHandler) GetSystemInitStatus(ctx *gin.Context) {
	h.logger.Debugf("received system init status handler...")

	// 调用服务层方法
	h.service.GetSystemInitStatus(ctx)
}

// Routers 获取系统相关路由列表
//
// @description 返回系统相关的所有HTTP路由定义
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
		&commonhttp.MyRouter{
			Name:            "systemInitStatus",
			Uri:             "system/init",
			Enabled:         true,
			EnableJWtVerify: false,
			Method:          http.MethodGet,
			Desc:            "获取baas是否进行初始化",
			HandlerFunc:     h.GetSystemInitStatus,
		},
	}
}
