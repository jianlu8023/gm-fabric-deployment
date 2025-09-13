package service

import (
	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model/node"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
)

// NodeServiceInterface 节点服务接口
// @description 定义节点服务的接口
// @interface
// @method NodeList 获取节点列表
type NodeServiceInterface interface {
	NodeList(ctx *gin.Context, req *request.NodeListRequest)
}

// NodeService 节点服务实现
// @description 实现NodeServiceInterface接口，处理节点相关业务逻辑
// @struct
// @property *Service 基础服务
// @property nodeMapper *mapper.NodeMapper 节点映射器
type NodeService struct {
	*Service
	nodeMapper *mapper.NodeMapper
}

// NodeList 获取节点列表服务
// @description 查询节点列表并返回分页结果
// @param ctx *gin.Context Gin上下文
// @param req *request.NodeListRequest 节点列表请求参数
func (s *NodeService) NodeList(ctx *gin.Context, req *request.NodeListRequest) {
	s.logger.Debugf("starting node list service...")
	s.logger.Debugf("request: %v", req)

	s.logger.Debugf("starting call mapper to query info...")
	page, err := s.nodeMapper.NodeList(node.Info{}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("query node list err: %v", err)
		webhttp.FailedResponse(ctx, webhttp.NewError(webhttp.NormalFailed, webhttp.ErrMsgNormalFailed))
		return
	}
	webhttp.SuccessResponse(ctx, page)
}

// NeeNodeService 创建节点服务实例
// @description 创建并返回一个新的节点服务实例
// @param service *Service 基础服务
// @param nodeMapper *mapper.NodeMapper 节点映射器
// @return *NodeService 节点服务实例
func NeeNodeService(service *Service, nodeMapper *mapper.NodeMapper) *NodeService {
	return &NodeService{
		Service:    service,
		nodeMapper: nodeMapper,
	}
}
