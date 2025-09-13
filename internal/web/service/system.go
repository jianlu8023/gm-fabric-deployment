package service

import (
	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	systeminfo "github.com/jianlu8023/gm-fabric-deployment/pkg/system/info"
)

// SystemService 系统服务
// @description 提供系统相关的服务功能，如获取系统概览信息
// @struct
// @property *Service 基础服务
// @property mapper *mapper.SystemMapper 系统映射器
type SystemService struct {
	*Service
	mapper *mapper.SystemMapper
}

// GetSystemOverview 获取系统概览信息
// @description 获取系统的基本信息，包括操作系统、CPU、磁盘和内存信息
// @param ctx *gin.Context Gin上下文
func (s *SystemService) GetSystemOverview(ctx *gin.Context) {
	s.logger.Debugf("received get system overview request...")

	os := systeminfo.InitOS()
	cpu, err := systeminfo.InitCPU()
	if err != nil {
		s.logger.Errorf("get cpu info failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.NormalFailed, "获取系统cpu信息失败")
		return
	}

	disk, err := systeminfo.InitDisk()
	if err != nil {
		s.logger.Errorf("get disk info failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.NormalFailed, "获取系统disk信息失败")
		return
	}
	ram, err := systeminfo.InitRAM()
	if err != nil {
		s.logger.Errorf("get ram info failed: %v", err)
		webhttp.FailedResponseWithMessage(ctx, webhttp.NormalFailed, "获取系统ram信息失败")
		return
	}

	webhttp.SuccessResponse(ctx, gin.H{
		"version": "1.0.0",
		"os":      os,
		"cpu":     cpu,
		"disk":    disk,
		"ram":     ram,
	})
}

// NewSystemService 创建系统服务实例
// @description 创建并返回一个新的系统服务实例
// @param service *Service 基础服务
// @param mapper *mapper.SystemMapper 系统映射器
// @return *SystemService 系统服务实例
func NewSystemService(service *Service, mapper *mapper.SystemMapper) *SystemService {
	return &SystemService{
		Service: service,
		mapper:  mapper,
	}
}
