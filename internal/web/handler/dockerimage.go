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

type DockerImageHandler struct {
	*Handler
	service *service.DockerImageService
}

func NewDockerImageHandler(baseHandler *Handler,
	service *service.DockerImageService,
) *DockerImageHandler {
	return &DockerImageHandler{
		Handler: baseHandler,
		service: service,
	}
}

type DockerImageServiceInterface interface {
	DockerImageList(ctx *gin.Context, req *request.DockerImageListRequest)
	DockerImagePull(ctx *gin.Context, req *request.DockerImagePullRequest)
}

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
