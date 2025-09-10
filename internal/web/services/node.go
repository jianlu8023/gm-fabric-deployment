package services

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/requests"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
)

type NodeService struct {
	*Service
	nodeMapper *mapper.NodeMapper
}

func (s *NodeService) NodeListService(ctx *gin.Context, req *requests.NodeListRequest) http.Error {

	s.logger.Debugf("starting node list service...")
	s.logger.Debugf("request: %v", req)

	s.nodeMapper.InsertOrUpdate(nil)

	return http.NewError(http.NormalFailed, "")
}

func NeeNodeService(service *Service, nodeMapper *mapper.NodeMapper) *NodeService {
	return &NodeService{
		Service:    service,
		nodeMapper: nodeMapper,
	}
}
