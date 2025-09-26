package service

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/ants"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DockerImageService struct {
	*Service
	mapper           *mapper.DockerImageMapper
	dockerControl    *docker.Control
	websocketControl *websocket.Control
	libp2pControl    *libp2p.Control
	antsPoolControl  *ants.Control
}

func NewDockerImageService(baseService *Service,
	mapper *mapper.DockerImageMapper,
	dockerControl *docker.Control,
	websocketControl *websocket.Control,
	libp2pControl *libp2p.Control,
	antsPoolControl *ants.Control,
) *DockerImageService {
	return &DockerImageService{
		Service:          baseService,
		mapper:           mapper,
		dockerControl:    dockerControl,
		websocketControl: websocketControl,
		libp2pControl:    libp2pControl,
		antsPoolControl:  antsPoolControl,
	}
}

func (s *DockerImageService) DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest) {
	s.logger.Debugf("received docker image list request with params: %v", req)

	page, err := s.mapper.DockerImageList(model.DockerImage{
		ImageLocationPeerId: req.PeerId,
	}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("get docker image list failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取docker镜像失败")
		return
	}

	listResponse, err := response.NewDockerImageListResponse(page)
	if err != nil {
		s.logger.Errorf("convert docker image list err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		return
	}

	commonhttp.SuccessResponse(ctx, listResponse)
}

func (s *DockerImageService) DockerImagePull(ctx *gin.Context, req *request.DockerImagePullRequest) {
	s.logger.Debugf("received docker image pull request with params: %v", req)

	if err := s.antsPoolControl.Submit(func() {
		s.logger.Debugf("starting to pull doker image: %v", req.ImageName)
		if !stringer.CompareIgnoreCase(req.PeerId, s.libp2pControl.GetLocalhostPeerID().String()) {
			// 需要发送libp2p消息到指定节点
			dockerPullMsg := &libp2p.Message{
				Type: libp2p.MsgDockerImagePull,
				Content: []byte(libp2p.DockerImagePullContent{
					ImageName: req.ImageName,
				}.String()),
				To:   peer.ID(req.PeerId),
				From: s.libp2pControl.GetLocalhostPeerID(),
			}
			if err := s.libp2pControl.SendMessageToPeer(dockerPullMsg.To, dockerPullMsg); err != nil {
				s.logger.Errorf("send docker pull message to peer failed: %v", err)
				return
			}
		} else {
			// 本机拉取镜像

			// 1. 先判断数据库中是否存在镜像
			s.logger.Debugf("check docker images exist status...")
			imageExist, err := s.dockerControl.GetImage(docker.WithImageQueryName(req.ImageName))
			if err != nil {
				s.logger.Errorf("get docker image failed: %v", err)
				return
			}
			if !stringer.IsBlank(imageExist.ID) {
				// 镜像id不等于空
				s.logger.Warnf("docker image %v already exist", req.ImageName)
				return
			}
			// 2. 拉取镜像
			s.logger.Debugf("docker image pull...")
			if err := s.dockerControl.PullImage(req.ImageName); err != nil {
				s.logger.Errorf("docker image pull failed: %v", err)
				return
			}
			// 3. 获取镜像信息
			s.logger.Debugf("get docker image info...")
			summary, err := s.dockerControl.GetImage(docker.WithImageQueryName(req.ImageName))
			if err != nil {
				s.logger.Errorf("get docker image info failed: %v", err)
				return
			}
			// 4. 保存镜像信息
			s.logger.Debugf("save docker image info...")
			image := model.NewDockerImage()
			image.ImageName = summary.RepoTags[0]
			image.ImageId = summary.ID
			datetime, err := humantime.ParseTimeLocal(fmt.Sprintf("%v", summary.Created))
			if err != nil {
				s.logger.Errorf("parse time on local failed: %v", err)
				return
			}
			image.ImageCreated = datetime
			labels, err := json.MarshalString(summary.Labels)
			if err != nil {
				s.logger.Errorf("get docker image info marshal failed: %v", err)
				return
			}
			image.ImageLabels = labels
			image.IsDelete = sql.NullBool{Bool: false, Valid: true}
			image.ImageLocationPeerId = req.PeerId
			if err := s.mapper.InsertOneWithCheck(image); err != nil {
				s.logger.Errorf("save docker image failed: %v", err)
				return
			}
			s.logger.Debugf("save docker image success, send websocket message to client...")
			s.websocketControl.Broadcast(websocket.Message{
				Content: []byte("pull image success"),
			})
		}
	}); err != nil {
		s.logger.Errorf("submit docker image pull task failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, commonhttp.ErrMsgInternalServerError)
		return
	}

	commonhttp.SuccessResponse(ctx, response.NewDockerImagePullResponse(
		fmt.Sprintf("拉取 %v 镜像成功", req.ImageName),
	))
}
