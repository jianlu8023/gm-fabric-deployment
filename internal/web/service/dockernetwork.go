package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
)

type DockerNetworkService struct {
	*Service
	mapper        *mapper.DockerNetworkMapper
	dockerControl *docker.Control
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
func (s *DockerNetworkService) DockerNetworkList(ctx *gin.Context, req *request.DockerNetworkListRequest) {
	s.logger.Debugf("received docker network list request with params: %v", req)

	page, err := s.mapper.DockerNetworkList(model.DockerNetwork{
		NetworkLocationPeerId: req.PeerId,
	}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("get docker network list failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取docker网络失败")
		return
	}

	listResponse, err := response.NewDockerNetworkListResponse(page)
	if err != nil {
		s.logger.Errorf("convert docker network list err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		return
	}

	commonhttp.SuccessResponse(ctx, listResponse)

}
