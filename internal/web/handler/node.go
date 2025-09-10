package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http/binding"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
)

type NodeHandler struct {
	*Handler
	service *service.NodeService
}

func NewNodeHandler(handler *Handler, service *service.NodeService) *NodeHandler {
	return &NodeHandler{
		Handler: handler,
		service: service,
	}
}

func (n *NodeHandler) NodeListHandler(ctx *gin.Context) {
	n.logger.Infof("received node list request...")
	req := new(request.NodeListRequest)
	if err := binding.BindFormData(ctx, req); err != nil {
		n.logger.Errorf("binding request failed: %v", err)
		http.FailedResponse(ctx, http.NewError(http.InvalidParameter, http.ErrMsgInvalidParameter))
		return
	}
	if !req.IsLegal() {
		n.logger.Debugf("node list request is illegal: %v", req)
		http.FailedResponseWithMessage(ctx, http.InvalidParameter, "输入参数不合法")
		return
	}

	n.service.NodeListService(ctx, req)
}
