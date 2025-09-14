package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
)

type DockerNetworkService struct {
	*Service
	mapper        *mapper.DockerNetworkMapper
	dockerControl *docker.Control
}

func (s *DockerNetworkService) DockerNetworkList(ctx *gin.Context, req *request.DockerNetworkListRequest) {
	s.logger.Debugf("received docker network list request with params: %v", req)

	if !req.IsLegal() {
		// 验证失败
		s.logger.Errorf("docker network list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}

	page, err := s.mapper.DockerNetworkList(model.DockerNetwork{
		NetworkLocationPeerId: req.PeerId,
	}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("get docker network list failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取docker网络失败")
		return
	}

	commonhttp.SuccessResponse(ctx, page)

}

func NewDockerNetworkService(baseService *Service,
	mapper *mapper.DockerNetworkMapper,
	dockerControl *docker.Control,
) *DockerNetworkService {
	return &DockerNetworkService{
		Service:       baseService,
		mapper:        mapper,
		dockerControl: dockerControl,
	}
}
