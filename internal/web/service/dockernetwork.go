package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// DockerNetworkService Docker网络服务接口
//
// @description 定义Docker网络服务需要实现的方法
// @interface
type DockerNetworkService interface {
	// DockerNetworkList 获取Docker网络列表
	//
	// @description 获取Docker网络列表信息
	// @param ctx *gin.Context Gin上下文
	// @param req *request.DockerNetworkListRequest Docker网络列表请求参数
	DockerNetworkList(ctx *gin.Context, req *request.DockerNetworkListRequest)
}

// dockerNetworkServiceImpl Docker网络服务结构体
//
// @description 提供Docker网络相关的服务功能，如获取Docker网络列表
// @struct
type dockerNetworkServiceImpl struct {
	*Service                                 // Service 基础服务，提供日志功能
	mapper        mapper.DockerNetworkMapper // mapper Docker网络映射器，用于数据访问
	dockerControl *docker.Control            // dockerControl Docker控制器，用于Docker操作
}

// NewDockerNetworkService 创建Docker网络服务实例
//
// @description 创建并返回一个新的Docker网络服务实例
// @param baseService *Service 基础服务
// @param mapper mapper.DockerNetworkMapper Docker网络映射器
// @param dockerControl *docker.Control Docker控制器
// @return DockerNetworkService Docker网络服务实例
func NewDockerNetworkService(baseService *Service,
	mapper mapper.DockerNetworkMapper,
	dockerControl *docker.Control,
) DockerNetworkService {
	return &dockerNetworkServiceImpl{
		Service:       baseService,
		mapper:        mapper,
		dockerControl: dockerControl,
	}
}

// DockerNetworkList 获取Docker网络列表
//
// @description 获取Docker网络列表信息
// @param ctx *gin.Context Gin上下文
// @param req *request.DockerNetworkListRequest Docker网络列表请求参数
func (s *dockerNetworkServiceImpl) DockerNetworkList(ctx *gin.Context, req *request.DockerNetworkListRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "dockerNetworkServiceImpl", "dockerNetworkList",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received docker network list request with params: %v", req)

	page, err := s.mapper.DockerNetworkList(model.DockerNetwork{
		NetworkLocationPeerId: req.PeerId,
	}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("get docker network list failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取docker网络失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	listResponse, err := response.NewDockerNetworkListResponse(page)
	if err != nil {
		s.logger.Errorf("convert docker network list err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	commonhttp.SuccessResponse(ctx, listResponse)
	span.SetStatus(codes.Ok, "success")
}
