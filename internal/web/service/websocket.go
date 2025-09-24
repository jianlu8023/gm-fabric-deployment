package service

import (
	"fmt"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/request"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/websocket"
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
	ws := &WebSocketService{
		Service:   service,
		wsMapper:  wsMapper,
		wsControl: wsControl,
	}
	
	// 初始化消息处理功能
	ws.initializeMessageHandling()
	
	return ws
}

// initializeMessageHandling 初始化消息处理功能
func (s *WebSocketService) initializeMessageHandling() {
	if s.wsControl == nil {
		s.logger.Warnf("[websocket service] wsControl is nil, skipping message handling initialization")
		return
	}
	
	s.logger.Infof("[websocket service] initializing message handling...")
	
	// 设置连接建立回调
	s.wsControl.SetOnConnected(s.onConnectionEstablished)
	
	// 设置连接断开回调
	s.wsControl.SetOnDisconnected(s.onConnectionClosed)
	
	// 设置消息接收回调
	s.wsControl.SetOnMessage(s.onMessageReceived)
	s.wsControl.SetOnTextMessage(s.onTextMessageReceived)
	s.wsControl.SetOnBinaryMessage(s.onBinaryMessageReceived)
	s.wsControl.SetOnPingMessage(s.onPingMessageReceived)
	s.wsControl.SetOnPongMessage(s.onPongMessageReceived)
	
	// 注册默认消息处理器
	s.registerDefaultMessageHandlers()
	
	s.logger.Infof("[websocket service] message handling initialized successfully")
}

// registerDefaultMessageHandlers 注册默认消息处理器
func (s *WebSocketService) registerDefaultMessageHandlers() {
	// 注册文本消息处理器
	textHandler := websocket.NewDefaultTextMessageHandler(s.logger)
	s.wsControl.RegisterMessageHandler(1, textHandler) // websocket.TextMessage
	
	// 注册二进制消息处理器
	binaryHandler := websocket.NewDefaultBinaryMessageHandler(s.logger)
	s.wsControl.RegisterMessageHandler(2, binaryHandler) // websocket.BinaryMessage
	
	// 注册ping消息处理器
	pingHandler := websocket.NewDefaultPingMessageHandler(s.logger)
	s.wsControl.RegisterMessageHandler(9, pingHandler) // websocket.PingMessage
	
	s.logger.Debugf("[websocket service] default message handlers registered")
}

// onConnectionEstablished 连接建立回调
func (s *WebSocketService) onConnectionEstablished(conn *websocket.Connection) {
	s.logger.Infof("[websocket service] new connection established: %s", conn.ID)
	
	// 可以在这里添加连接建立时的业务逻辑
	// 例如：记录连接日志、发送欢迎消息等
	
	// 发送欢迎消息
	welcomeMessage := map[string]interface{}{
		"type": "welcome",
		"data": map[string]interface{}{
			"connectionId": conn.ID,
			"timestamp":   conn.CreatedAt.Unix(),
			"message":     "Welcome to WebSocket server",
		},
	}
	
	if data, err := s.marshalMessage(welcomeMessage); err == nil {
		s.wsControl.SendTo(conn.ID, data)
	}
}

// onConnectionClosed 连接关闭回调
func (s *WebSocketService) onConnectionClosed(conn *websocket.Connection) {
	s.logger.Infof("[websocket service] connection closed: %s", conn.ID)
	
	// 可以在这里添加连接关闭时的业务逻辑
	// 例如：清理资源、更新用户状态等
}

// onMessageReceived 通用消息接收回调
func (s *WebSocketService) onMessageReceived(conn *websocket.Connection, message []byte) {
	s.logger.Debugf("[websocket service] received message from %s: %s", conn.ID, string(message))
	
	// 可以在这里添加通用的消息处理逻辑
	// 例如：消息日志记录、统计等
}

// onTextMessageReceived 文本消息接收回调
func (s *WebSocketService) onTextMessageReceived(conn *websocket.Connection, message []byte) {
	s.logger.Debugf("[websocket service] received text message from %s: %s", conn.ID, string(message))
	
	// 这里可以添加特定的文本消息处理逻辑
	// 例如：解析命令、处理聊天消息等
}

// onBinaryMessageReceived 二进制消息接收回调
func (s *WebSocketService) onBinaryMessageReceived(conn *websocket.Connection, message []byte) {
	s.logger.Debugf("[websocket service] received binary message from %s: %d bytes", conn.ID, len(message))
	
	// 这里可以添加二进制消息处理逻辑
	// 例如：文件上传、图片处理等
}

// onPingMessageReceived ping消息接收回调
func (s *WebSocketService) onPingMessageReceived(conn *websocket.Connection, message []byte) {
	s.logger.Debugf("[websocket service] received ping from %s", conn.ID)
	// ping消息通常用于心跳检测，一般不需要特殊处理
}

// onPongMessageReceived pong消息接收回调
func (s *WebSocketService) onPongMessageReceived(conn *websocket.Connection, message []byte) {
	s.logger.Debugf("[websocket service] received pong from %s", conn.ID)
	// pong消息用于响应ping，表示连接正常
}

// marshalMessage 序列化消息
func (s *WebSocketService) marshalMessage(message interface{}) ([]byte, error) {
	return commonhttp.Marshal(message)
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
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "输入参数不合法")
		return
	}

	// 从上下文中获取用户信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		s.logger.Errorf("userID not found in context")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "认证失败，请先登录")
		return
	}

	userIDStr := userID.(string)

	// 这里可以调用WebSocket控制器的UpgradeConnection方法来升级连接
	// 但是在服务层，我们只处理业务逻辑，连接升级的具体实现会在handler层处理
	s.logger.Infof("websocket connection request from user: %s, node: %s", userIDStr, req.NodeID)
	_, _, err := s.wsControl.UpgradeConnection(ctx.Writer, ctx.Request)
	if err != nil {
		s.logger.Errorf("websocket connection upgrade failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, "WebSocket连接升级失败")
		return
	}

	// 返回成功响应
	commonhttp.SuccessResponse(ctx, map[string]string{
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
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "认证失败，请先登录")
		return
	}

	userIDStr := userID.(string)

	// 验证连接是否属于该用户
	conn, err := s.wsMapper.GetConnectionByID(connID)
	if err != nil {
		s.logger.Errorf("get connection by id err: %v", err)
		commonhttp.FailedResponse(ctx, commonhttp.NewError(commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed))
		return
	}

	if conn.UserID != userIDStr {
		s.logger.Errorf("connection %s does not belong to user %s", connID, userIDStr)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Forbidden, "无权操作该连接")
		return
	}

	// 更新连接状态
	err = s.wsMapper.UpdateConnectionStatus(connID, "disconnected")
	if err != nil {
		s.logger.Errorf("update connection status err: %v", err)
		commonhttp.FailedResponse(ctx, commonhttp.NewError(commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed))
		return
	}

	// 从控制器中移除连接
	s.wsControl.RemoveConnection(connID)

	commonhttp.SuccessResponse(ctx, map[string]string{
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
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		s.logger.Errorf("userID not found in context")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "认证失败，请先登录")
		return
	}

	currentUserIDStr := currentUserID.(string)

	// 验证请求参数
	if req.Message == "" {
		s.logger.Errorf("message content is empty")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "消息内容不能为空")
		return
	}

	// 准备消息数据
	messageData, err := s.marshalMessage(map[string]interface{}{
		"type":      "message",
		"from":      currentUserIDStr,
		"content":   req.Message,
		"timestamp": s.getCurrentTimestamp(),
	})
	if err != nil {
		s.logger.Errorf("failed to marshal message: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, "消息序列化失败")
		return
	}

	var successCount int
	var targetInfo string

	// 根据消息类型和目标进行不同的处理
	if req.ConnID != "" {
		// 发送到特定连接
		s.logger.Infof("sending message to connection: %s", req.ConnID)
		if s.wsControl.SendTo(req.ConnID, messageData) {
			successCount = 1
		}
		targetInfo = fmt.Sprintf("connection %s", req.ConnID)
	} else if req.NodeID != "" {
		// 发送到特定节点的所有连接
		s.logger.Infof("sending message to node: %s", req.NodeID)
		successCount = s.wsControl.BroadcastByNodeID(req.NodeID, messageData)
		targetInfo = fmt.Sprintf("node %s (%d connections)", req.NodeID, successCount)
	} else if req.UserID != "" {
		// 发送到特定用户的所有连接
		s.logger.Infof("sending message to user: %s", req.UserID)
		successCount = s.wsControl.BroadcastByUserID(req.UserID, messageData)
		targetInfo = fmt.Sprintf("user %s (%d connections)", req.UserID, successCount)
	} else {
		// 广播到所有连接
		s.logger.Info("broadcasting message to all connections")
		s.wsControl.Broadcast(messageData)
		successCount = s.wsControl.GetConnectionCount()
		targetInfo = fmt.Sprintf("all connections (%d total)", successCount)
	}

	// 返回成功响应
	response := map[string]interface{}{
		"message":     "websocket message sent successfully",
		"target":      targetInfo,
		"successCount": successCount,
		"timestamp":   s.getCurrentTimestamp(),
	}

	commonhttp.SuccessResponse(ctx, response)
}

// getCurrentTimestamp 获取当前时间戳
func (s *WebSocketService) getCurrentTimestamp() int64 {
	return time.Now().Unix()
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
		commonhttp.FailedResponse(ctx, commonhttp.NewError(commonhttp.NormalFailed, commonhttp.ErrMsgNormalFailed))
		return
	}

	commonhttp.SuccessResponse(ctx, page)
}
