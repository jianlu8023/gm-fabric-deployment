package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model/node"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
)

type NodeServiceInterface interface {
	NodeList(ctx *gin.Context, req *request.NodeListRequest)
}

type NodeService struct {
	*Service
	nodeMapper *mapper.NodeMapper
}

func (s *NodeService) NodeListService(ctx *gin.Context, req *request.NodeListRequest) {
	s.logger.Debugf("starting node list service...")
	s.logger.Debugf("request: %v", req)

	s.logger.Debugf("starting call mapper to query info...")
	page, err := s.nodeMapper.NodeList(node.Info{}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("query node list err: %v", err)
		http.FailedResponse(ctx, http.NewError(http.NormalFailed, http.ErrMsgNormalFailed))
		return
	}
	http.SuccessResponse(ctx, page)
}

func NeeNodeService(service *Service, nodeMapper *mapper.NodeMapper) *NodeService {
	return &NodeService{
		Service:    service,
		nodeMapper: nodeMapper,
	}
}
