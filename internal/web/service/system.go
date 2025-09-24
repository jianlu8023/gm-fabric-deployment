package service

import (
	"github.com/gin-gonic/gin"
	systeminfo "github.com/jianlu8023/go-tools/v2/pkg/system/info"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/version"
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
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统cpu信息失败")
		return
	}

	disk, err := systeminfo.InitDisk()
	if err != nil {
		s.logger.Errorf("get disk info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统disk信息失败")
		return
	}
	ram, err := systeminfo.InitRAM()
	if err != nil {
		s.logger.Errorf("get ram info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统ram信息失败")
		return
	}

	commonhttp.SuccessResponse(ctx, gin.H{
		"version": version.Version,
		"os":      os,
		"cpu":     cpu,
		"disk":    disk,
		"ram":     ram,
	})
}

func (s *SystemService) GetSystemInitStatus(ctx *gin.Context) {
	s.logger.Debugf("received get system init status request...")

	init, err := s.mapper.GetSystemInit()
	if err != nil {
		s.logger.Errorf("get system init status failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统初始化状态失败")
		return
	}
	commonhttp.SuccessResponse(ctx, gin.H{
		"init": init,
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
