package service

import (
	"github.com/gin-gonic/gin"
	systemcpu "github.com/jianlu8023/go-tools/v2/pkg/system/cpu"
	systemdisk "github.com/jianlu8023/go-tools/v2/pkg/system/disk"
	systemos "github.com/jianlu8023/go-tools/v2/pkg/system/os"
	systemram "github.com/jianlu8023/go-tools/v2/pkg/system/ram"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/version"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// SystemService 系统服务结构体
//
// @description 提供系统相关的服务功能，如获取系统概览信息
// @struct
type SystemService struct {
	*Service                      // Service 基础服务，提供日志功能
	mapper   *mapper.SystemMapper // mapper 系统映射器，用于数据访问
}

// NewSystemService 创建系统服务实例
//
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

// GetSystemOverview 获取系统概览信息
//
// @description 获取系统的基本信息，包括操作系统、CPU、磁盘和内存信息
// @param ctx *gin.Context Gin上下文
// @param req *request.SystemOverviewRequest 系统概览请求参数
func (s *SystemService) GetSystemOverview(ctx *gin.Context, req *request.SystemOverviewRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "systemService", "getSystemOverview",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received system overview request with params: %v", req)

	os := systemos.SystemOsInfo()
	cpu, err := systemcpu.SystemCpuInfo()
	if err != nil {
		s.logger.Errorf("get cpu info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统cpu信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	disk, err := systemdisk.SystemDiskInfo()
	if err != nil {
		s.logger.Errorf("get disk info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统disk信息失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	ram, err := systemram.SystemRamInfo()
	if err != nil {
		s.logger.Errorf("get ram info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统ram信息失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	commonhttp.SuccessResponse(ctx, gin.H{
		"version": version.Version,
		"os":      os,
		"cpu":     cpu,
		"disk":    disk,
		"ram":     ram,
	})
	span.SetStatus(codes.Ok, "success")
}

// GetSystemInitStatus 获取系统初始化状态
//
// @description 检查系统是否已经完成初始化配置
// @param ctx *gin.Context Gin上下文
// @param req *request.SystemInitStatusRequest 系统初始化状态请求参数
func (s *SystemService) GetSystemInitStatus(ctx *gin.Context, req *request.SystemInitStatusRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "systemService", "getSystemInitStatus",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received system init status request with params: %v", req)

	init, err := s.mapper.GetSystemInit()
	if err != nil {
		s.logger.Errorf("get system init status failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "获取系统初始化状态失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	commonhttp.SuccessResponse(ctx, response.NewSystemInitStatusResponse(init))
	span.SetStatus(codes.Ok, "success")
}
