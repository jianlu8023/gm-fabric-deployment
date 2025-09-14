package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
)

type DockerImageService struct {
	*Service
	mapper        *mapper.DockerImageMapper
	dockerControl *docker.Control
}

func (s *DockerImageService) DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest) {
	s.logger.Debugf("received docker image list request with params: %v", req)

	if !req.IsLegal() {
		// 验证失败
		s.logger.Errorf("docker image list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}

	page, err := s.mapper.DockerImageList(model.DockerImage{
		ImageLocationPeerId: req.PeerId,
	}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("get docker image list failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取docker镜像失败")
		return
	}

	commonhttp.SuccessResponse(ctx, page)

}

func NewDockerImageService(baseService *Service,
	mapper *mapper.DockerImageMapper,
	dockerControl *docker.Control,
) *DockerImageService {
	return &DockerImageService{
		Service:       baseService,
		mapper:        mapper,
		dockerControl: dockerControl,
	}
}
