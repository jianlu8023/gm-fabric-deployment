package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/request"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	controlwebsocket "github.com/jianlu8023/golang-example/pkg/control/websocket"
)

// WebSocketService WebSocket服务接口
//
// @description 定义WebSocket服务的接口规范
// @interface
type WebSocketService interface {
	// ConnectService 处理WebSocket连接请求
	//
	// @param ctx *gin.Context HTTP上下文
	// @param req *request.WSConnectRequest WebSocket连接请求参数
	// @param userId any 用户ID
	ConnectService(ctx *gin.Context, req *request.WSConnectRequest, userId any)

	// DisconnectService 处理WebSocket断开连接请求
	//
	// @param ctx *gin.Context HTTP上下文
	// @param req *request.WSDisconnectRequest WebSocket断开连接请求参数
	DisconnectService(ctx *gin.Context, req *request.WSDisconnectRequest)

	// SendMessageService 处理发送WebSocket消息请求
	//
	// @param ctx *gin.Context HTTP上下文
	// @param req *request.WSMessageRequest WebSocket消息请求参数
	SendMessageService(ctx *gin.Context, req *request.WSMessageRequest)

	// GetConnectionListService 获取连接列表
	//
	// @param ctx *gin.Context HTTP上下文
	// @param req *request.WSConnectionListRequest 连接列表请求参数
	// @param userID string 用户ID
	GetConnectionListService(ctx *gin.Context, req *request.WSConnectionListRequest, userID string)
}

// webSocketService WebSocket服务结构体
//
// @description WebSocket服务的具体实现，处理WebSocket相关的业务逻辑
// @struct
type webSocketService struct {
	*Service                            // Service 基础服务，提供日志功能
	wsMapper  mapper.WebSocketMapper    // wsMapper WebSocket映射器，用于数据访问
	wsControl *controlwebsocket.Control // wsControl WebSocket控制器，用于WebSocket操作
}

// NewWebSocketService 创建一个新的WebSocketService实例
//
// @description 创建并返回一个新的WebSocket服务实例
// @param service *Service 基础服务
// @param wsMapper mapper.WebSocketMapper WebSocket映射器
// @param wsControl *controlwebsocket.Control WebSocket控制器
// @return WebSocketService WebSocket服务实例
func NewWebSocketService(service *Service, wsMapper mapper.WebSocketMapper, wsControl *controlwebsocket.Control) WebSocketService {
	ws := &webSocketService{
		Service:   service,
		wsMapper:  wsMapper,
		wsControl: wsControl,
	}

	// 初始化消息处理功能
	ws.initializeMessageHandling()

	return ws
}

// initializeMessageHandling 初始化消息处理功能
//
// @description 初始化WebSocket消息处理功能，设置各种回调函数
func (s *webSocketService) initializeMessageHandling() {
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
	// 添加关闭消息回调
	s.wsControl.SetOnCloseMessage(s.onCloseMessageReceived)

	// 注册默认消息处理器
	s.registerDefaultMessageHandlers()

	s.logger.Infof("[websocket service] message handling initialized successfully")
}

// registerDefaultMessageHandlers 注册默认消息处理器
//
// @description 注册默认的WebSocket消息处理器
func (s *webSocketService) registerDefaultMessageHandlers() {
	// 注册文本消息处理器

	s.logger.Debugf("[websocket service] default message handlers registered")
}

// onConnectionEstablished 连接建立回调
//
// @description WebSocket连接建立时的回调函数，发送欢迎消息
// @param conn *controlwebsocket.Connection 连接对象
func (s *webSocketService) onConnectionEstablished(conn *controlwebsocket.Connection) {
	s.logger.Infof("[websocket service] new connection established: %s", conn.NodeID)

	// 可以在这里添加连接建立时的业务逻辑
	// 例如：记录连接日志、发送欢迎消息等

	// 发送欢迎消息
	welcomeMessage := map[string]interface{}{
		"type": "welcome",
		"data": map[string]interface{}{
			"connectionId": conn.NodeID,
			"timestamp":    conn.CreatedAt.Unix(),
			"message":      "Welcome to WebSocket server",
		},
	}

	if data, err := s.marshalMessage(welcomeMessage); err == nil {
		// 使用正确的Message结构体和SendMessage方法
		message := controlwebsocket.Message{
			Type:    websocket.TextMessage,
			Content: data,
		}
		s.wsControl.SendMessage(conn.NodeID, message)
	}
}

// SendMessageService 处理发送WebSocket消息请求
//
// @description 处理发送WebSocket消息请求，根据目标类型进行不同的消息发送处理
// @param ctx *gin.Context Gin上下文
// @param req *request.WSMessageRequest WebSocket消息请求参数
func (s *webSocketService) SendMessageService(ctx *gin.Context, req *request.WSMessageRequest) {
	s.logger.Debugf("starting websocket send message service...")
	s.logger.Debugf("request: %v", req)

	// 准备消息数据
	messageData, err := s.marshalMessage(map[string]interface{}{
		"type":      "message",
		"content":   req.Message,
		"timestamp": time.GetCurrentTimestamp(),
	})
	if err != nil {
		s.logger.Errorf("failed to marshal message: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, "消息序列化失败")
		return
	}

	// 创建Message结构体
	message := controlwebsocket.Message{
		Type:    websocket.TextMessage,
		Content: messageData,
	}

	var targetInfo string

	// 根据消息类型和目标进行不同的处理
	if !stringer.IsBlank(req.NodeID) {
		// 发送到特定节点的所有连接
		s.logger.Infof("sending message to node: %s", req.NodeID)
		s.wsControl.SendMessage(req.NodeID, message)
		targetInfo = fmt.Sprintf("node %s", req.NodeID)
	} else {
		// 广播到所有连接
		s.logger.Info("broadcasting message to all connections")
		s.wsControl.Broadcast(message)
		targetInfo = fmt.Sprintf("all connections (%d total)", s.wsControl.GetConnectionCount())
	}

	// 返回成功响应
	response := map[string]interface{}{
		"message":   "websocket message sent successfully",
		"target":    targetInfo,
		"timestamp": time.GetCurrentTimestamp(),
	}

	commonhttp.SuccessResponse(ctx, response)
}

// onConnectionClosed 连接关闭回调
//
// @description WebSocket连接关闭时的回调函数
// @param conn *controlwebsocket.Connection 连接对象
func (s *webSocketService) onConnectionClosed(conn *controlwebsocket.Connection) {
	s.logger.Infof("[websocket service] connection closed: %s", conn.NodeID)

	// 可以在这里添加连接关闭时的业务逻辑
	// 例如：清理资源、更新用户状态等
}

// onMessageReceived 通用消息接收回调
//
// @description 接收WebSocket消息的通用回调函数
// @param conn *controlwebsocket.Connection 连接对象
// @param message controlwebsocket.Message 消息对象
func (s *webSocketService) onMessageReceived(conn *controlwebsocket.Connection, message controlwebsocket.Message) {
	s.logger.Debugf("[websocket service] received message from %s: %s", conn.NodeID, string(message.Content))

	// 可以在这里添加通用的消息处理逻辑
	// 例如：消息日志记录、统计等
}

// onTextMessageReceived 文本消息接收回调
//
// @description 接收WebSocket文本消息的回调函数
// @param conn *controlwebsocket.Connection 连接对象
// @param message controlwebsocket.Message 消息对象
func (s *webSocketService) onTextMessageReceived(conn *controlwebsocket.Connection, message controlwebsocket.Message) {
	s.logger.Debugf("[websocket service] received text message from %s: %s", conn.NodeID, string(message.Content))

	// 这里可以添加特定的文本消息处理逻辑
	// 例如：解析命令、处理聊天消息等
}

// onBinaryMessageReceived 二进制消息接收回调
//
// @description 接收WebSocket二进制消息的回调函数
// @param conn *controlwebsocket.Connection 连接对象
// @param message controlwebsocket.Message 消息对象
func (s *webSocketService) onBinaryMessageReceived(conn *controlwebsocket.Connection, message controlwebsocket.Message) {
	s.logger.Debugf("[websocket service] received binary message from %s: %d bytes", conn.NodeID, len(message.Content))

	// 这里可以添加二进制消息处理逻辑
	// 例如：文件上传、图片处理等
}

// onPingMessageReceived ping消息接收回调
//
// @description 接收WebSocket ping消息的回调函数
// @param conn *controlwebsocket.Connection 连接对象
// @param message controlwebsocket.Message 消息对象
func (s *webSocketService) onPingMessageReceived(conn *controlwebsocket.Connection, message controlwebsocket.Message) {
	s.logger.Debugf("[websocket service] received ping from %s", conn.NodeID)
	// ping消息通常用于心跳检测，一般不需要特殊处理
}

// onPongMessageReceived pong消息接收回调
//
// @description 接收WebSocket pong消息的回调函数
// @param conn *controlwebsocket.Connection 连接对象
// @param message controlwebsocket.Message 消息对象
func (s *webSocketService) onPongMessageReceived(conn *controlwebsocket.Connection, message controlwebsocket.Message) {
	s.logger.Debugf("[websocket service] received pong from %s", conn.NodeID)
	// pong消息用于响应ping，表示连接正常
}

// onCloseMessageReceived 关闭消息接收回调
//
// @description 接收WebSocket关闭消息的回调函数
// @param conn *controlwebsocket.Connection 连接对象
// @param message controlwebsocket.Message 消息对象
func (s *webSocketService) onCloseMessageReceived(conn *controlwebsocket.Connection, message controlwebsocket.Message) {
	s.logger.Debugf("[websocket service] received close message from %s", conn.NodeID)
	// 处理关闭消息，可以在这里进行最后的资源清理
}

// marshalMessage 序列化消息
//
// @description 将消息对象序列化为字节数据
// @param message interface{} 消息对象
// @return []byte 序列化后的字节数据
// @return error 错误信息
func (s *webSocketService) marshalMessage(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// ConnectService 处理WebSocket连接请求
//
// @description 处理WebSocket连接请求，验证参数，从上下文获取用户信息，升级连接
// @param ctx *gin.Context Gin上下文
// @param req *request.WSConnectRequest WebSocket连接请求参数
// @param userId any 用户ID
func (s *webSocketService) ConnectService(ctx *gin.Context, req *request.WSConnectRequest, userId any) {
	s.logger.Debugf("starting websocket connect service...")
	s.logger.Debugf("request: %v", req)

	userIDStr := userId.(string)

	// 调用WebSocket控制器升级连接
	s.logger.Infof("websocket connection request from user: %s, node: %s", userIDStr, req.NodeID)
	_, _, err := s.wsControl.UpgradeConnection(ctx.Writer, ctx.Request, req.NodeID)
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
//
// @description 处理WebSocket断开连接请求，验证用户权限，更新连接状态，移除连接
// @param ctx *gin.Context Gin上下文
// @param req *request.WSDisconnectRequest WebSocket断开连接请求参数
func (s *webSocketService) DisconnectService(ctx *gin.Context, req *request.WSDisconnectRequest) {
	s.logger.Debugf("starting websocket disconnect service for connection: %s", req.NodeID)

	connection := s.wsControl.GetConnection(req.NodeID)

	if stringer.IsBlank(connection.NodeID) {
		s.logger.Warnf("websocket connection not found for connection: %s", req.NodeID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InternalServerError, "WebSocket连接不存在")
		return
	}

	// 从控制器中移除连接
	s.wsControl.RemoveConnection(req.NodeID)

	commonhttp.SuccessResponse(ctx, map[string]string{
		"message": "websocket disconnected successfully",
	})
}

// GetConnectionListService 获取连接列表
//
// @description 获取指定用户的WebSocket连接列表
// @param ctx *gin.Context Gin上下文
// @param req *request.WSConnectionListRequest 连接列表请求参数
// @param userID string 用户ID
func (s *webSocketService) GetConnectionListService(ctx *gin.Context, req *request.WSConnectionListRequest, userID string) {
	s.logger.Debugf("starting get connection list service for user: %s", userID)

	// 从wsControl中获取所有连接信息
	connections := s.wsControl.GetAllConnections()
	if connections == nil {
		s.logger.Debugf("no connections found")
		connections = make([]*controlwebsocket.Connection, 0)
	}

	// 如果需要分页，进行分页处理
	var result interface{}
	if req.IsPage {
		// 计算分页
		total := len(connections)
		start := (req.PageNo - 1) * req.PageSize
		end := start + req.PageSize

		if start >= total {
			connections = make([]*controlwebsocket.Connection, 0)
		} else if end > total {
			connections = connections[start:]
		} else {
			connections = connections[start:end]
		}

		// 构造分页结果
		result = map[string]interface{}{
			"list":     connections,
			"total":    total,
			"pageNo":   req.PageNo,
			"pageSize": req.PageSize,
			"isPage":   true,
		}
	} else {
		// 不分页，直接返回所有连接
		result = connections
	}

	commonhttp.SuccessResponse(ctx, result)
}
