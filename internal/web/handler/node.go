package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	webhhtp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http/binding"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
)

// NodeHandler 节点处理器结构体
//
// @description 处理节点相关的HTTP请求
// @struct
type NodeHandler struct {
	// Handler 基础处理器，提供日志功能
	*Handler
	// service 节点服务，处理节点相关的业务逻辑
	service *service.NodeService
}

// NewNodeHandler 创建节点处理器
//
// @param handler *Handler 基础处理器
// @param service *service.NodeService 节点服务
// @return *NodeHandler 节点处理器实例
func NewNodeHandler(handler *Handler, service *service.NodeService) *NodeHandler {
	return &NodeHandler{
		Handler: handler,
		service: service,
	}
}

// NodeListHandler 获取节点列表的处理函数
//
// @description 处理获取节点列表的HTTP请求
// @method GET
// @url /api/v1/node/list
// @param page int 页码 (可选, 默认:1)
// @param pageSize int 每页数量 (可选, 默认:10)
// @return JSON 节点列表和分页信息
func (h *NodeHandler) NodeListHandler(ctx *gin.Context) {
	h.logger.Infof("received node list request...")

	// 验证JWT和Session信息
	// userID, exists := ctx.Get("user_id")
	// if !exists {
	// 	h.logger.Errorf("JWT authentication failed: userID not found")
	// 	webhhtp.FailedResponseWithMessage(ctx, webhhtp.Unauthorized, "认证失败，请先登录")
	// 	return
	// }

	// sessionID, exists := ctx.Get("session_id")
	// if !exists {
	// 	h.logger.Errorf("Session validation failed: sessionID not found")
	// 	webhhtp.FailedResponseWithMessage(ctx, webhhtp.SessionExpired, "会话已过期，请重新登录")
	// 	return
	// }

	// h.logger.Debugf("User %v with session %v is accessing node list", userID, sessionID)

	// 验证通过后继续处理请求
	req := new(request.NodeListRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		h.logger.Errorf("binding request failed: %v", err)
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Infof("validation errors: %v", messages)
		webhhtp.FailedResponse(ctx, webhhtp.NewError(webhhtp.InvalidParameter, webhhtp.ErrMsgInvalidParameter))
		return
	}
	if !req.IsLegal() {
		h.logger.Debugf("node list request is illegal: %v", req)
		webhhtp.FailedResponseWithMessage(ctx, webhhtp.InvalidParameter, "输入参数不合法")
		return
	}

	h.service.NodeList(ctx, req)
}

// Routers 获取节点相关路由列表
//
// @return []commonhttp.RouterHandler 节点路由处理器列表
func (h *NodeHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "nodeList",
			Uri:             "node/list",
			Method:          http.MethodGet,
			HandlerFunc:     h.NodeListHandler,
			Enabled:         true,
			Desc:            "获取libp2p已发现的节点",
			EnableJWtVerify: false,
		},
	}

}
