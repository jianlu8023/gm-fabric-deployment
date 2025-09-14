package service

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/websocket"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
)

type DockerImageService struct {
	*Service
	mapper           *mapper.DockerImageMapper
	dockerControl    *docker.Control
	websocketControl *websocket.Control
}

func NewDockerImageService(baseService *Service,
	mapper *mapper.DockerImageMapper,
	dockerControl *docker.Control,
	websocketControl *websocket.Control,
) *DockerImageService {
	return &DockerImageService{
		Service:          baseService,
		mapper:           mapper,
		dockerControl:    dockerControl,
		websocketControl: websocketControl,
	}
}

func (s *DockerImageService) DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest) {
	s.logger.Debugf("received docker image list request with params: %v", req)

	if !req.IsLegal() {
		// 验证失败
		s.logger.Errorf("docker image list request is no legal...")
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

func (s *DockerImageService) DockerImagePull(ctx *gin.Context, req *request.DockerImagePullRequest) {
	s.logger.Debugf("received docker image pull request with params: %v", req)

	if !req.IsLegal() {
		// 验证失败
		s.logger.Errorf("docker image pull request is no legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}

	go func() {
		if err := s.dockerControl.PullImage(req.ImageName); err != nil {
			s.logger.Errorf("docker image pull failed: %v", err)
			// commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, err.Error())
			return
		}
		summary, err := s.dockerControl.GetImage()
		if err != nil {
			s.logger.Errorf("get docker image info failed: %v", err)
			return
		}
		image := model.NewDockerImage()
		image.ImageName = summary.RepoTags[0]
		image.ImageId = summary.ID
		image.ImageCreated = summary.Created
		labels, err := json.Marshal(summary.Labels)
		if err != nil {
			s.logger.Errorf("get docker image info marshal failed: %v", err)
			return
		}
		image.ImageLabels = string(labels)
		image.IsDelete = sql.NullBool{Bool: false, Valid: true}
		image.ImageLocationPeerId = req.PeerId
		if err := s.mapper.InsertOneWithCheck(image); err != nil {
			s.logger.Errorf("save docker image failed: %v", err)

		}
		s.websocketControl.Broadcast([]byte("pull image success"))
	}()
	commonhttp.SuccessResponse(ctx, gin.H{
		"msg": fmt.Sprintf("拉取 %v 镜像成功", req.ImageName),
	})
}
