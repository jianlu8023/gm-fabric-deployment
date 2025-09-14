package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
)

// NodeServiceInterface 节点服务接口
// @description 定义节点服务的接口
// @interface
// @method NodeList 获取节点列表
type NodeServiceInterface interface {
	Libp2pNodeList(ctx *gin.Context, req *request.Libp2pNodeListRequest)
}

// Libp2pNodeService 节点服务实现
// @description 实现NodeServiceInterface接口，处理节点相关业务逻辑
// @struct
// @property *Service 基础服务
// @property nodeMapper *mapper.Libp2pNodeMapper 节点映射器
type Libp2pNodeService struct {
	*Service
	nodeMapper    *mapper.Libp2pNodeMapper
	libp2pControl *libp2p.Control
}

// NeeNodeService 创建节点服务实例
// @description 创建并返回一个新的节点服务实例
// @param service *Service 基础服务
// @param nodeMapper *mapper.Libp2pNodeMapper 节点映射器
// @param libp2pControl *libp2p.Control libp2p控制器
// @return *Libp2pNodeService 节点服务实例
func NeeNodeService(service *Service,
	nodeMapper *mapper.Libp2pNodeMapper,
	libp2pControl *libp2p.Control,
) *Libp2pNodeService {
	return &Libp2pNodeService{
		Service:       service,
		nodeMapper:    nodeMapper,
		libp2pControl: libp2pControl,
	}
}

// Libp2pNodeList 获取节点列表服务
// @description 查询节点列表并返回分页结果
// @param ctx *gin.Context Gin上下文
// @param req *request.Libp2pNodeListRequest 节点列表请求参数
func (s *Libp2pNodeService) Libp2pNodeList(ctx *gin.Context, req *request.Libp2pNodeListRequest) {
	s.logger.Debugf("received libp2p node list request with params: %v", req)

	s.logger.Debugf("starting call mapper to query info...")
	page, err := s.nodeMapper.NodeList(model.Libp2pNode{}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("query node list err: %v", err)
		http.FailedResponse(ctx, http.NewError(http.NormalFailed, http.ErrMsgNormalFailed))
		return
	}
	http.SuccessResponse(ctx, page)
}
