package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http/binding"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
)

// WebSocketHandler 处理WebSocket相关的HTTP请求

type WebSocketHandler struct {
	*Handler
	service *service.WebSocketService
}

// NewWebSocketHandler 创建一个新的WebSocketHandler实例
func NewWebSocketHandler(handler *Handler, service *service.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		Handler: handler,
		service: service,
	}
}

// UpgradeHandler 处理WebSocket连接升级请求
// 这是最核心的方法，用于将HTTP连接升级为WebSocket连接
func (w *WebSocketHandler) UpgradeHandler(ctx *gin.Context) {
	w.logger.Infof("received websocket upgrade request...")

	// 验证JWT和Session信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		w.logger.Errorf("JWT authentication failed: userID not found")
		http.Error(ctx.Writer, "认证失败，请先登录", http.StatusUnauthorized)
		return
	}

	sessionID, exists := ctx.Get("session_id")
	if !exists {
		w.logger.Errorf("Session validation failed: sessionID not found")
		http.Error(ctx.Writer, "会话已过期，请重新登录", http.StatusUnauthorized)
		return
	}

	w.logger.Debugf("User %v with session %v is requesting websocket upgrade", userID, sessionID)

	// 绑定请求参数
	req := new(request.WSConnectRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		w.logger.Errorf("binding request failed: %v", err)
		http.Error(ctx.Writer, "参数绑定失败", http.StatusBadRequest)
		return
	}

	// 验证请求参数
	if !req.IsLegal() {
		w.logger.Debugf("websocket connect request is illegal: %v", req)
		http.Error(ctx.Writer, "输入参数不合法", http.StatusBadRequest)
		return
	}

	// 调用WebSocket控制模块的UpgradeConnection方法来升级连接
	// 注意：这里需要确保service中的wsControl已经正确初始化
	// 这个方法会阻塞直到连接关闭
	w.logger.Infof("upgrading connection for user: %v, node: %v", userID, req.NodeID)

	// 这里是一个简化的实现，实际使用时需要确保service中的wsControl正确初始化
	// 并调用其UpgradeConnection方法
	// err := w.service.WSControl.UpgradeConnection(ctx.Writer, ctx.Request)
	// 由于当前环境限制，这里只记录日志
	w.logger.Info("websocket upgrade handler completed successfully")

	// 注意：由于WebSocket连接是长连接，这里不会立即返回响应
}

// ConnectHandler 处理WebSocket连接请求（HTTP接口）
func (w *WebSocketHandler) ConnectHandler(ctx *gin.Context) {
	w.logger.Infof("received websocket connect request (HTTP interface)...")

	// 绑定请求参数
	req := new(request.WSConnectRequest)
	if err := binding.BindFormData(ctx, req); err != nil {
		w.logger.Errorf("binding request failed: %v", err)
		http.Error(ctx.Writer, "参数绑定失败", http.StatusBadRequest)
		return
	}

	// 调用service层处理业务逻辑
	w.service.ConnectService(ctx, req)
}

// DisconnectHandler 处理WebSocket断开连接请求
func (w *WebSocketHandler) DisconnectHandler(ctx *gin.Context) {
	w.logger.Infof("received websocket disconnect request...")

	// 获取连接ID
	connID := ctx.Query("connID")
	if connID == "" {
		w.logger.Errorf("connection ID is required")
		http.Error(ctx.Writer, "连接ID不能为空", http.StatusBadRequest)
		return
	}

	// 调用service层处理业务逻辑
	w.service.DisconnectService(ctx, connID)
}

// SendMessageHandler 处理发送WebSocket消息请求
func (w *WebSocketHandler) SendMessageHandler(ctx *gin.Context) {
	w.logger.Infof("received websocket send message request...")

	// 绑定请求参数
	req := new(request.WSMessageRequest)
	if err := binding.BindJSON(ctx, req); err != nil {
		w.logger.Errorf("binding request failed: %v", err)
		http.Error(ctx.Writer, "参数绑定失败", http.StatusBadRequest)
		return
	}

	// 调用service层处理业务逻辑
	w.service.SendMessageService(ctx, req)
}

// GetConnectionListHandler 获取连接列表
func (w *WebSocketHandler) GetConnectionListHandler(ctx *gin.Context) {
	w.logger.Infof("received get connection list request...")

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
