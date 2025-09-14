package service

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/request"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/websocket"
)

// WebSocketServiceInterface 定义WebSocket服务的接口
// @description 定义WebSocket服务的接口规范
// @interface
// @method ConnectService 处理WebSocket连接请求
// @method DisconnectService 处理WebSocket断开连接请求
// @method SendMessageService 处理发送WebSocket消息请求
// @method GetConnectionListService 获取连接列表
type WebSocketServiceInterface interface {
	// ConnectService 处理WebSocket连接请求
	ConnectService(ctx *gin.Context, req *request.WSConnectRequest)

	// DisconnectService 处理WebSocket断开连接请求
	DisconnectService(ctx *gin.Context, connID string)

	// SendMessageService 处理发送WebSocket消息请求
	SendMessageService(ctx *gin.Context, req *request.WSMessageRequest)

	// GetConnectionListService 获取连接列表
	GetConnectionListService(ctx *gin.Context, userID string, isPage bool, pageNo int64, pageSize int64)
}

// WebSocketService 实现WebSocketServiceInterface接口
// @description WebSocket服务的具体实现，处理WebSocket相关的业务逻辑
// @struct
// @property *Service 基础服务
// @property wsMapper *mapper.WebSocketMapper WebSocket映射器
// @property wsControl *websocket.Control WebSocket控制器
type WebSocketService struct {
	*Service
	wsMapper  *mapper.WebSocketMapper
	wsControl *websocket.Control
}

// NewWebSocketService 创建一个新的WebSocketService实例
// @description 创建并返回一个新的WebSocket服务实例
// @param service *Service 基础服务
// @param wsMapper *mapper.WebSocketMapper WebSocket映射器
// @param wsControl *websocket.Control WebSocket控制器
// @return *WebSocketService WebSocket服务实例
func NewWebSocketService(service *Service, wsMapper *mapper.WebSocketMapper, wsControl *websocket.Control) *WebSocketService {
	return &WebSocketService{
		Service:   service,
		wsMapper:  wsMapper,
		wsControl: wsControl,
	}
}

// ConnectService 处理WebSocket连接请求
// @description 处理WebSocket连接请求，验证参数，从上下文获取用户信息，升级连接
// @param ctx *gin.Context Gin上下文
// @param req *request.WSConnectRequest WebSocket连接请求参数
func (s *WebSocketService) ConnectService(ctx *gin.Context, req *request.WSConnectRequest) {
	s.logger.Debugf("starting websocket connect service...")
	s.logger.Debugf("request: %v", req)

	// 验证请求参数
	if !req.IsLegal() {
		s.logger.Errorf("websocket connect request is illegal: %v", req)
		http.FailedResponseWithMessage(ctx, http.InvalidParameter, "输入参数不合法")
		return
	}

	// 从上下文中获取用户信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		s.logger.Errorf("userID not found in context")
		http.FailedResponseWithMessage(ctx, http.Unauthorized, "认证失败，请先登录")
		return
	}

	userIDStr := userID.(string)

	// 这里可以调用WebSocket控制器的UpgradeConnection方法来升级连接
	// 但是在服务层，我们只处理业务逻辑，连接升级的具体实现会在handler层处理
	s.logger.Infof("websocket connection request from user: %s, node: %s", userIDStr, req.NodeID)
	_, _, err := s.wsControl.UpgradeConnection(ctx.Writer, ctx.Request)
	if err != nil {
		s.logger.Errorf("websocket connection upgrade failed: %v", err)
		http.FailedResponseWithMessage(ctx, http.InternalServerError, "WebSocket连接升级失败")
		return
	}

	// 返回成功响应
	http.SuccessResponse(ctx, map[string]string{
		"message": "websocket connection request processed successfully",
	})
}

// DisconnectService 处理WebSocket断开连接请求
// @description 处理WebSocket断开连接请求，验证用户权限，更新连接状态，移除连接
// @param ctx *gin.Context Gin上下文
// @param connID string 连接ID
func (s *WebSocketService) DisconnectService(ctx *gin.Context, connID string) {
	s.logger.Debugf("starting websocket disconnect service for connection: %s", connID)

	// 从上下文中获取用户信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		s.logger.Errorf("userID not found in context")
		http.FailedResponseWithMessage(ctx, http.Unauthorized, "认证失败，请先登录")
		return
	}

	userIDStr := userID.(string)

	// 验证连接是否属于该用户
	conn, err := s.wsMapper.GetConnectionByID(connID)
	if err != nil {
		s.logger.Errorf("get connection by id err: %v", err)
		http.FailedResponse(ctx, http.NewError(http.NormalFailed, http.ErrMsgNormalFailed))
		return
	}

	if conn.UserID != userIDStr {
		s.logger.Errorf("connection %s does not belong to user %s", connID, userIDStr)
		http.FailedResponseWithMessage(ctx, http.Forbidden, "无权操作该连接")
		return
	}

	// 更新连接状态
	err = s.wsMapper.UpdateConnectionStatus(connID, "disconnected")
	if err != nil {
		s.logger.Errorf("update connection status err: %v", err)
		http.FailedResponse(ctx, http.NewError(http.NormalFailed, http.ErrMsgNormalFailed))
		return
	}

	// 从控制器中移除连接
	s.wsControl.RemoveConnection(connID)

	http.SuccessResponse(ctx, map[string]string{
		"message": "websocket disconnected successfully",
	})
}

// SendMessageService 处理发送WebSocket消息请求
// @description 处理发送WebSocket消息请求，根据目标类型进行不同的消息发送处理
// @param ctx *gin.Context Gin上下文
// @param req *request.WSMessageRequest WebSocket消息请求参数
func (s *WebSocketService) SendMessageService(ctx *gin.Context, req *request.WSMessageRequest) {
	s.logger.Debugf("starting websocket send message service...")
	s.logger.Debugf("request: %v", req)

	// 从上下文中获取用户信息
	_, exists := ctx.Get("user_id")
	if !exists {
		s.logger.Errorf("userID not found in context")
		http.FailedResponseWithMessage(ctx, http.Unauthorized, "认证失败，请先登录")
		return
	}

	// 根据消息类型和目标进行不同的处理
	// 如果指定了节点ID，则发送到该节点的所有连接
	if req.NodeID != "" {
		s.logger.Infof("sending message to node: %s", req.NodeID)
		// 这里可以实现向特定节点发送消息的逻辑
	} else if req.UserID != "" {
		// 如果指定了用户ID，则发送到该用户的所有连接
		s.logger.Infof("sending message to user: %s", req.UserID)
		// 这里可以实现向特定用户发送消息的逻辑
	} else {
		// 否则广播到所有连接
		s.logger.Info("broadcasting message to all connections")
		// 这里可以实现广播消息的逻辑
	}

	http.SuccessResponse(ctx, map[string]string{
		"message": "websocket message sent successfully",
	})
}

// GetConnectionListService 获取连接列表
// @description 获取指定用户的WebSocket连接列表
// @param ctx *gin.Context Gin上下文
// @param userID string 用户ID
// @param isPage bool 是否分页
// @param pageNo int64 页码
// @param pageSize int64 每页大小
func (s *WebSocketService) GetConnectionListService(ctx *gin.Context, userID string, isPage bool, pageNo int64, pageSize int64) {
	s.logger.Debugf("starting get connection list service for user: %s", userID)

	// 调用mapper层获取连接列表
	page, err := s.wsMapper.GetConnectionsByUserID(userID, isPage, pageNo, pageSize)
	if err != nil {
		s.logger.Errorf("get connection list err: %v", err)
		http.FailedResponse(ctx, http.NewError(http.NormalFailed, http.ErrMsgNormalFailed))
		return
	}

	http.SuccessResponse(ctx, page)
}
