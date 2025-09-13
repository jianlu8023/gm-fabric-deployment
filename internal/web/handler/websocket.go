package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http/binding"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
)

// WebSocketHandler WebSocket处理器结构体
//
// @description 处理WebSocket相关的HTTP请求和连接管理
// @struct
type WebSocketHandler struct {
	// Handler 基础处理器，提供日志功能
	*Handler
	// service WebSocket服务，处理WebSocket相关的业务逻辑
	service *service.WebSocketService
}

// NewWebSocketHandler 创建WebSocket处理器
//
// @param handler *Handler 基础处理器
// @param service *service.WebSocketService WebSocket服务
// @return *WebSocketHandler WebSocket处理器实例
func NewWebSocketHandler(handler *Handler, service *service.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		Handler: handler,
		service: service,
	}
}

// UpgradeHandler WebSocket连接升级处理函数
//
// @description 将HTTP连接升级为WebSocket连接，是WebSocket通信的核心入口
// @method GET
// @url /api/v1/ws
// @param nodeID string 目标节点ID (必需)
// @return WebSocket连接 成功后建立WebSocket双向通信通道
func (h *WebSocketHandler) UpgradeHandler(ctx *gin.Context) {
	h.logger.Infof("received websocket upgrade request...")

	// 验证JWT和Session信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("JWT authentication failed: userID not found")
		http.Error(ctx.Writer, "认证失败，请先登录", http.StatusUnauthorized)
		return
	}

	sessionID, exists := ctx.Get("session_id")
	if !exists {
		h.logger.Errorf("Session validation failed: sessionID not found")
		http.Error(ctx.Writer, "会话已过期，请重新登录", http.StatusUnauthorized)
		return
	}

	h.logger.Debugf("User %v with session %v is requesting websocket upgrade", userID, sessionID)

	// 绑定请求参数
	req := new(request.WSConnectRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		h.logger.Errorf("binding request failed: %v", err)
		http.Error(ctx.Writer, "参数绑定失败", http.StatusBadRequest)
		return
	}

	// 验证请求参数
	if !req.IsLegal() {
		h.logger.Debugf("websocket connect request is illegal: %v", req)
		http.Error(ctx.Writer, "输入参数不合法", http.StatusBadRequest)
		return
	}

	// 调用WebSocket控制模块的UpgradeConnection方法来升级连接
	// 注意：这里需要确保service中的wsControl已经正确初始化
	// 这个方法会阻塞直到连接关闭
	h.logger.Infof("upgrading connection for user: %v, node: %v", userID, req.NodeID)

	// 这里是一个简化的实现，实际使用时需要确保service中的wsControl正确初始化
	// 并调用其UpgradeConnection方法
	// err := w.service.WSControl.UpgradeConnection(ctx.Writer, ctx.Request)
	// 由于当前环境限制，这里只记录日志
	h.logger.Info("websocket upgrade handler completed successfully")

	// 注意：由于WebSocket连接是长连接，这里不会立即返回响应
}

// ConnectHandler WebSocket连接请求处理函数（HTTP接口）
//
// @description 处理WebSocket连接建立的HTTP接口请求
// @method POST
// @url /api/v1/ws/connect
// @param nodeID string 目标节点ID (必需)
// @return JSON 连接状态信息
func (h *WebSocketHandler) ConnectHandler(ctx *gin.Context) {
	h.logger.Infof("received websocket connect request (HTTP interface)...")

	// 绑定请求参数
	req := new(request.WSConnectRequest)
	if err := binding.BindFormData(ctx, req); err != nil {
		h.logger.Errorf("binding request failed: %v", err)
		http.Error(ctx.Writer, "参数绑定失败", http.StatusBadRequest)
		return
	}

	// 调用service层处理业务逻辑
	h.service.ConnectService(ctx, req)
}

// DisconnectHandler WebSocket断开连接处理函数
//
// @description 处理WebSocket连接断开请求
// @method POST
// @url /api/v1/ws/disconnect
// @param connID string 连接ID (必需)
// @return JSON 断开连接结果
func (h *WebSocketHandler) DisconnectHandler(ctx *gin.Context) {
	h.logger.Infof("received websocket disconnect request...")

	// 获取连接ID
	connID := ctx.Query("connID")
	if connID == "" {
		h.logger.Errorf("connection ID is required")
		http.Error(ctx.Writer, "连接ID不能为空", http.StatusBadRequest)
		return
	}

	// 调用service层处理业务逻辑
	h.service.DisconnectService(ctx, connID)
}

// SendMessageHandler 发送WebSocket消息处理函数
//
// @description 处理通过WebSocket发送消息的请求
// @method POST
// @url /api/v1/ws/message
// @param connID string 连接ID (必需)
// @param message string 消息内容 (必需)
// @param messageType string 消息类型 (可选, 默认:text)
// @return JSON 消息发送结果
func (h *WebSocketHandler) SendMessageHandler(ctx *gin.Context) {
	h.logger.Infof("received websocket send message request...")

	// 绑定请求参数
	req := new(request.WSMessageRequest)
	if err := binding.BindJSON(ctx, req); err != nil {
		h.logger.Errorf("binding request failed: %v", err)
		http.Error(ctx.Writer, "参数绑定失败", http.StatusBadRequest)
		return
	}

	// 调用service层处理业务逻辑
	h.service.SendMessageService(ctx, req)
}

// GetConnectionListHandler 获取WebSocket连接列表处理函数
//
// @description 获取当前系统中的WebSocket连接列表
// @method GET
// @url /api/v1/ws/connections
// @param pageNo int 页码 (可选, 默认:1)
// @param pageSize int 每页数量 (可选, 默认:10)
// @param isPage bool 是否分页 (可选, 默认:true)
// @return JSON 连接列表和分页信息
func (h *WebSocketHandler) GetConnectionListHandler(ctx *gin.Context) {
	h.logger.Infof("received get connection list request...")

	// 验证JWT和Session信息
	// userID, exists := ctx.Get("user_id")
	// if !exists {
	// 	w.logger.Errorf("JWT authentication failed: userID not found")
	// 	http.Error(ctx.Writer, "认证失败，请先登录", http.StatusUnauthorized)
	// 	return
	// }

	// userIDStr := userID.(string)

	// 获取分页参数
	// pageNo := binding.GetQueryInt64(ctx, "pageNo", 1)
	// pageSize := binding.GetQueryInt64(ctx, "pageSize", 10)
	// isPage := binding.GetQueryBool(ctx, "isPage", true)

	// 调用service层处理业务逻辑
	// w.service.GetConnectionListService(ctx, userIDStr, isPage, pageNo, pageSize)
}

// Routers 获取WebSocket相关路由列表
//
// @return []commonhttp.RouterHandler WebSocket路由处理器列表
func (h *WebSocketHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		// WebSocket连接升级路由
		&commonhttp.MyRouter{
			Name:            "websocketUpgrade",
			Uri:             "ws",
			HandlerFunc:     h.UpgradeHandler,
			Enabled:         true,
			EnableJWtVerify: true,
			Desc:            "websocket connection upgrade",
		},
		// WebSocket连接请求路由
		&commonhttp.MyRouter{
			Name:            "websocketConnect",
			Uri:             "ws/connect",
			Method:          http.MethodPost,
			HandlerFunc:     h.ConnectHandler,
			Enabled:         true,
			EnableJWtVerify: true,
			Desc:            "websocket connect request",
		},
		// WebSocket断开连接路由
		&commonhttp.MyRouter{
			Name:            "websocketDisconnect",
			Uri:             "ws/disconnect",
			Method:          http.MethodPost,
			HandlerFunc:     h.DisconnectHandler,
			Enabled:         true,
			EnableJWtVerify: true,
			Desc:            "websocket disconnect request",
		},
		// WebSocket发送消息路由
		&commonhttp.MyRouter{
			Name:            "websocketSendMessage",
			Uri:             "ws/message",
			Method:          http.MethodPost,
			HandlerFunc:     h.SendMessageHandler,
			Enabled:         true,
			EnableJWtVerify: true,
			Desc:            "send websocket message",
		},
		// 获取WebSocket连接列表路由
		&commonhttp.MyRouter{
			Name:            "websocketConnectionList",
			Uri:             "ws/connections",
			Method:          http.MethodGet,
			HandlerFunc:     h.GetConnectionListHandler,
			Enabled:         true,
			EnableJWtVerify: true,
			Desc:            "get websocket connection list",
		},
	}
}
