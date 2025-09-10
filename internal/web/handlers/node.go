package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/requests"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/services"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http/binding"
)

type NodeHandler struct {
	*Handler
	service *services.NodeService
}

func NewNodeHandler(handler *Handler, service *services.NodeService) *NodeHandler {
	return &NodeHandler{
		Handler: handler,
		service: service,
	}
}

func (n *NodeHandler) NodeListHandler(ctx *gin.Context) {
	n.logger.Infof("received node list request...")
	req := new(requests.NodeListRequest)
	if err := binding.BindFormData(ctx, req); err != nil {
		n.logger.Errorf("binding request failed: %v", err)
		http.FailedResponse(ctx, http.NewError(http.InvalidParameter, http.ErrMsgInvalidParameter))
		return
	}
	// 暂时没想好怎么返回
	err := n.service.NodeListService(ctx, req)
	http.FailedResponse(ctx, err)
	return

}
