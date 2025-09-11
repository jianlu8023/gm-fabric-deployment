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

	// 验证JWT和Session信息
	userID, exists := ctx.Get("userID")
	if !exists {
		n.logger.Errorf("JWT authentication failed: userID not found")
		http.FailedResponseWithMessage(ctx, http.Unauthorized, "认证失败，请先登录")
		return
	}

	sessionID, exists := ctx.Get("sessionID")
	if !exists {
		n.logger.Errorf("Session validation failed: sessionID not found")
		http.FailedResponseWithMessage(ctx, http.SessionExpired, "会话已过期，请重新登录")
		return
	}

	n.logger.Debugf("User %v with session %v is accessing node list", userID, sessionID)

	// 验证通过后继续处理请求
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
