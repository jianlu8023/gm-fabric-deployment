package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Libp2pNodeService 节点服务结构体
//
// @description 实现NodeServiceInterface接口，处理节点相关业务逻辑
// @struct
type Libp2pNodeService struct {
	*Service              // Service 基础服务，提供日志功能
	mapper               *mapper.Libp2pNodeMapper  // mapper 节点映射器，用于数据访问
	libp2pControl        *libp2p.Control           // libp2pControl libp2p控制器，用于节点操作
}

// NewLibp2pNodeService 创建新的节点服务实例
//
// @description 创建并返回一个新的节点服务实例
// @param service *Service 基础服务
// @param nodeMapper *mapper.Libp2pNodeMapper 节点映射器
// @param libp2pControl *libp2p.Control libp2p控制器
// @return *Libp2pNodeService 节点服务实例
func NewLibp2pNodeService(service *Service,
	nodeMapper *mapper.Libp2pNodeMapper,
	libp2pControl *libp2p.Control,
) *Libp2pNodeService {
	return &Libp2pNodeService{
		Service:       service,
		mapper:        nodeMapper,
		libp2pControl: libp2pControl,
	}
}

// Libp2pNodeList 获取节点列表服务
//
// @description 查询节点列表并返回分页结果
// @param ctx *gin.Context Gin上下文
// @param req *request.Libp2pNodeListRequest 节点列表请求参数
func (s *Libp2pNodeService) Libp2pNodeList(ctx *gin.Context, req *request.Libp2pNodeListRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "libp2pNodeService", "libp2pNodeList",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received libp2p node list request with params: %v", req)

	s.logger.Debugf("starting call mapper to query info...")
	page, err := s.mapper.NodeList(ctx.Request.Context(), model.Libp2pNode{}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("query node list err: %v", err)
		commonhttp.FailedResponse(ctx, commonhttp.NewError(commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed))
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	nodeResponse, err := response.ToLibp2pNodeListResponse(page)
	if err != nil {
		s.logger.Errorf("convert node list err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	commonhttp.SuccessResponse(ctx, nodeResponse)
	span.SetStatus(codes.Ok, "success")
}

// Libp2pNodeMyself 获取当前节点信息
//
// @description 查询当前节点信息并返回结果
// @param ctx *gin.Context HTTP上下文
// @param req *request.Libp2pNodeMyselfRequest 当前节点信息请求参数
func (s *Libp2pNodeService) Libp2pNodeMyself(ctx *gin.Context, req *request.Libp2pNodeMyselfRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "libp2pNodeService", "libp2pNodeMyself",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received libp2p node myself request with params: %v", req)

	peerId := s.libp2pControl.GetLocalhostPeerID().String()
	myself, err := s.mapper.NodeMyself(ctx.Request.Context(), peerId)
	if err != nil {
		s.logger.Errorf("query node myself err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	myselfResponse, err := response.ToLibp2pNodeMyselfResponse(myself)
	if err != nil {
		s.logger.Errorf("convert node myself err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	s.logger.Debugf("from database query result %v convert result %v", myself, myselfResponse)

	commonhttp.SuccessResponse(ctx, myselfResponse)
	span.SetStatus(codes.Ok, "success")
}
