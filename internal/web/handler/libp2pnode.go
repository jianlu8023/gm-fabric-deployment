package handler

import (
	"net/http"
	"strings"

	"github.com/jianlu8023/golang-example/pkg/common/http/binding"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// Libp2pNodeHandler 节点处理器结构体
//
// @description 处理节点相关的HTTP请求
// @struct
type Libp2pNodeHandler struct {
	// Handler 基础处理器，提供日志功能
	*Handler
	// service 节点服务，处理节点相关的业务逻辑
	service *service.Libp2pNodeService
}

// NewLibp2pNodeHandler 创建节点处理器
//
// @param handler *Handler 基础处理器
// @param service *service.Libp2pNodeService 节点服务
// @return *Libp2pNodeHandler 节点处理器实例
func NewLibp2pNodeHandler(handler *Handler, service *service.Libp2pNodeService) *Libp2pNodeHandler {
	return &Libp2pNodeHandler{
		Handler: handler,
		service: service,
	}
}

// Libp2pNodeList 获取节点列表的处理函数
//
// @description 处理获取节点列表的HTTP请求
// @method GET
// @url /api/v1/node/list
// @param page int 页码 (可选, 默认:1)
// @param pageSize int 每页数量 (可选, 默认:10)
// @return JSON 节点列表和分页信息
func (h *Libp2pNodeHandler) Libp2pNodeList(ctx *gin.Context) {
	h.logger.Infof("received libp2p node list handler...")

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
	req := new(request.Libp2pNodeListRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
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

	h.service.Libp2pNodeList(ctx, req)
}

func (h *Libp2pNodeHandler) Libp2pNodeMyself(ctx *gin.Context) {
	h.logger.Infof("received libp2p node myself handler...")
	h.service.Libp2pNodeMyself(ctx)
}

// Routers 获取节点相关路由列表
//
// @return []commonhttp.RouterHandler 节点路由处理器列表
func (h *Libp2pNodeHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "libp2pNodeList",
			Uri:             "libp2p/list",
			Method:          http.MethodGet,
			HandlerFunc:     h.Libp2pNodeList,
			Enabled:         true,
			Desc:            "获取libp2p已发现的节点",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "libp2pNodeMyself",
			Uri:             "libp2p/myself",
			Method:          http.MethodGet,
			HandlerFunc:     h.Libp2pNodeMyself,
			Enabled:         true,
			Desc:            "获取libp2p本机节点",
			EnableJWtVerify: false,
		},
	}
}
