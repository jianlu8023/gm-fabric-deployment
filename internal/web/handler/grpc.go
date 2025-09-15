package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"net/http"
)

type GrpcHandler struct {
	*Handler
	service *service.GrpcService
}

func NewGrpcHandler(baseHandler *Handler,
	service *service.GrpcService) *GrpcHandler {
	return &GrpcHandler{
		Handler: baseHandler,
		service: service,
	}
}

func (h *GrpcHandler) SendGrpcPingMessage(ctx *gin.Context) {
	h.logger.Debugf("received send grpc ping message handler...")
}

func (h *GrpcHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "SendGrpcPingMessage",
			Uri:             "/grpc/send/message/ping",
			Method:          http.MethodGet,
			HandlerFunc:     h.SendGrpcPingMessage,
			Enabled:         true,
			EnableJWtVerify: false,
			Desc:            "发送Grpc的pingMessage",
		},
	}
}
