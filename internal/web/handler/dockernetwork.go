package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/common/http/binding"
	"net/http"
	"strings"
)

type DockerNetworkHandler struct {
	*Handler
	service *service.DockerNetworkService
}

func NewDockerNetworkHandler(baseHandler *Handler,
	service *service.DockerNetworkService,
) *DockerNetworkHandler {
	return &DockerNetworkHandler{
		Handler: baseHandler,
		service: service,
	}
}

func (h *DockerNetworkHandler) DockerNetworkList(ctx *gin.Context) {
	h.logger.Debugf("received docker network list handler...")
	req := new(request.DockerNetworkListRequest)
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

	h.service.DockerNetworkList(ctx, req)
}

func (h *DockerNetworkHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "DockerNetworkList",
			Method:          http.MethodGet,
			Uri:             "docker/network/list",
			HandlerFunc:     h.DockerNetworkList,
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "获取docker网络列表",
		},
	}
}
