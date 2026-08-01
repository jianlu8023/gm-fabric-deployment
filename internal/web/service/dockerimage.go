package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/sqlnull"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/ants"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/pkg/control/websocket"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// DockerImageService Docker镜像服务接口
//
// @description 定义Docker镜像服务需要实现的方法
// @interface
type DockerImageService interface {
	// DockerImageList 获取Docker镜像列表
	//
	// @param ctx *gin.Context Gin上下文
	// @param req *request.DockerImageListRequest Docker镜像列表请求参数
	DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest)
	// DockerImagePull 拉取Docker镜像
	//
	// @param ctx *gin.Context Gin上下文
	// @param req *request.DockerImagePullRequest Docker镜像拉取请求参数
	DockerImagePull(ctx *gin.Context, req *request.DockerImagePullRequest)
}

// dockerImageServiceImpl Docker镜像服务结构体
//
// @description 提供Docker镜像相关的服务功能，如镜像列表查询、镜像拉取等
// @struct
type dockerImageServiceImpl struct {
	*Service                                  // Service 基础服务，提供日志功能
	mapper           mapper.DockerImageMapper // mapper Docker镜像映射器，用于数据访问
	dockerControl    *docker.Control          // dockerControl Docker控制器，用于Docker操作
	websocketControl *websocket.Control       // websocketControl WebSocket控制器，用于消息推送
	libp2pControl    *libp2p.Control          // libp2pControl libp2p控制器，用于节点通信
	antsPoolControl  *ants.Control            // antsPoolControl 线程池控制器，用于异步任务
}

// NewDockerImageService 创建Docker镜像服务实例
//
// @description 创建并返回一个新的Docker镜像服务实例
// @param baseService *Service 基础服务
// @param mapper mapper.DockerImageMapper Docker镜像映射器
// @param dockerControl *docker.Control Docker控制器
// @param websocketControl *websocket.Control WebSocket控制器
// @param libp2pControl *libp2p.Control libp2p控制器
// @param antsPoolControl *ants.Control 线程池控制器
// @return *dockerImageServiceImpl Docker镜像服务实例
func NewDockerImageService(baseService *Service, mapper mapper.DockerImageMapper, dockerControl *docker.Control, websocketControl *websocket.Control, libp2pControl *libp2p.Control, antsPoolControl *ants.Control) DockerImageService {
	return &dockerImageServiceImpl{
		Service:          baseService,
		mapper:           mapper,
		dockerControl:    dockerControl,
		websocketControl: websocketControl,
		libp2pControl:    libp2pControl,
		antsPoolControl:  antsPoolControl,
	}
}

// DockerImageList 查询Docker镜像列表
//
// @description 根据查询条件获取Docker镜像列表
// @param ctx *gin.Context Gin上下文
// @param req *request.DockerImageListRequest 镜像列表查询请求参数
func (s *dockerImageServiceImpl) DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "dockerImageServiceImpl", "dockerImageList",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()

	s.logger.Debugf("received docker image list request with params: %v", req)

	page, err := s.mapper.DockerImageList(ctx.Request.Context(), model.DockerImage{
		ImageLocationPeerId: req.PeerId,
	}, req.IsPage, req.PageNo, req.PageSize)
	if err != nil {
		s.logger.Errorf("get docker image list failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取docker镜像失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	listResponse, err := response.NewDockerImageListResponse(page)
	if err != nil {
		s.logger.Errorf("convert docker image list err: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	commonhttp.SuccessResponse(ctx, listResponse)
	span.SetStatus(codes.Ok, "success")
}

// DockerImagePull 拉取Docker镜像
//
// @description 拉取指定名称的Docker镜像到本地或远程节点
// @param ctx *gin.Context Gin上下文
// @param req *request.DockerImagePullRequest 镜像拉取请求参数
func (s *dockerImageServiceImpl) DockerImagePull(ctx *gin.Context, req *request.DockerImagePullRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "dockerImageServiceImpl", "dockerImagePull",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
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
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				return
			}
		} else {
			// 本机拉取镜像

			// 1. 先判断数据库中是否存在镜像
			s.logger.Debugf("check docker images exist status...")
			imageExist, err := s.dockerControl.GetImage(docker.WithImageQueryName(req.ImageName))
			if err != nil {
				s.logger.Errorf("get docker image failed: %v", err)
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				return
			}
			if !stringer.IsBlank(imageExist.ID) {
				// 镜像id不等于空
				s.logger.Warnf("docker image %v already exist", req.ImageName)
				return
			}
			// 2. 拉取镜像
			s.logger.Debugf("docker image pull...")
			if err = s.dockerControl.PullImage(req.ImageName); err != nil {
				s.logger.Errorf("docker image pull failed: %v", err)
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				return
			}
			// 3. 获取镜像信息
			s.logger.Debugf("get docker image info...")
			summary, err := s.dockerControl.GetImage(docker.WithImageQueryName(req.ImageName))
			if err != nil {
				s.logger.Errorf("get docker image info failed: %v", err)
				span.RecordError(err)
				return
			}
			// 4. 保存镜像信息
			s.logger.Debugf("save docker image info...")
			image := model.NewDockerImage()
			image.ImageName = summary.RepoTags
			image.ImageId = summary.ID
			image.ImageCreated = summary.Created
			image.ImageLabels = summary.Labels
			image.IsDelete = sqlnull.FalseToNull()
			image.ImageLocationPeerId = req.PeerId
			if err := s.mapper.InsertOneWithCheck(ctx.Request.Context(), image); err != nil {
				s.logger.Errorf("save docker image failed: %v", err)
				span.RecordError(err)
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
		span.RecordError(err)
		return
	}

	commonhttp.SuccessResponse(ctx, response.NewDockerImagePullResponse(
		fmt.Sprintf("拉取 %v 镜像成功", req.ImageName),
	))
	// span.SetStatus(codes.Ok, "success")
}
