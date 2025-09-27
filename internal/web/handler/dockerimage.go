package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
)

// DockerImageHandler Docker镜像处理器结构体
//
// @description 处理Docker镜像相关的HTTP请求
// @struct
type DockerImageHandler struct {
	*Handler
	service *service.DockerImageService
}

// NewDockerImageHandler 创建Docker镜像处理器
//
// @param baseHandler *Handler 基础处理器
// @param service *service.DockerImageService Docker镜像服务
// @return *DockerImageHandler Docker镜像处理器实例
func NewDockerImageHandler(baseHandler *Handler,
	service *service.DockerImageService,
) *DockerImageHandler {
	return &DockerImageHandler{
		Handler: baseHandler,
		service: service,
	}
}

// DockerImageServiceInterface Docker镜像服务接口
//
// @description 定义Docker镜像服务需要实现的方法
// @interface
type DockerImageServiceInterface interface {
	DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest)
	DockerImagePull(ctx *gin.Context, req *request.DockerImagePullRequest)
}

// DockerImageList 处理Docker镜像列表请求
//
// @description 处理获取Docker镜像列表的HTTP请求
// @method GET
// @url /api/v1/docker/image/list
// @param peer_id string 节点ID (可选)
// @param is_page bool 是否分页 (可选)
// @param page_no int 页码 (可选，当is_page为true时必填)
// @param page_size int 每页大小 (可选，当is_page为true时必填)
// @return JSON 镜像列表数据
func (h *DockerImageHandler) DockerImageList(ctx *gin.Context) {
	h.logger.Debugf("received docker image list handler...")

	req := new(request.DockerImageListRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages,
		)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, strings.Join(msg, ","))
		return
	}
	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("docker image list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.DockerImageList(ctx, req)
}

// DockerImagePull 处理Docker镜像拉取请求
//
// @description 处理拉取Docker镜像的HTTP请求
// @method POST
// @url /api/v1/docker/image/pull
// @param peer_id string 节点ID (必需)
// @param image_name string 镜像名称 (必需)
// @return JSON 镜像拉取结果
func (h *DockerImageHandler) DockerImagePull(ctx *gin.Context) {
	h.logger.Debugf("received docker image pull handler...")

	req := new(request.DockerImagePullRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, strings.Join(msg, ","))
		return
	}
	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("docker image pull request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.DockerImagePull(ctx, req)
}

// Routers 获取Docker镜像相关路由列表
//
// @return []commonhttp.RouterHandler Docker镜像路由处理器列表
func (h *DockerImageHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "DockerImageList",
			Uri:             "docker/image/list",
			Method:          http.MethodGet,
			HandlerFunc:     h.DockerImageList,
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "获取docker镜像列表",
		},
		&commonhttp.MyRouter{
			Name:            "DockerImagePull",
			Uri:             "docker/image/pull",
			Method:          http.MethodPost,
			HandlerFunc:     h.DockerImagePull,
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "拉取docker镜像",
		},
	}
}
