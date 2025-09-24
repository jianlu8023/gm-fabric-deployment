package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"

	"github.com/gorilla/websocket"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

// MessageHandler 消息处理器接口
type MessageHandler interface {
	// HandleMessage 处理接收到的消息
	HandleMessage(conn *Connection, messageType int, message []byte) error
}

// Message 消息结构体
type Message struct {
	Type      int       `json:"type"`      // 消息类型
	Content   []byte    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
	From      string    `json:"from"`      // 发送方ID
	To        string    `json:"to"`        // 接收方ID（可选）
}

// Response 响应结构体
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Connection 表示一个WebSocket连接
type Connection struct {
	Conn         *websocket.Conn
	ID           string
	UserID       string // 用户ID
	NodeID       string // 节点ID
	Send         chan []byte
	CloseChan    chan struct{}
	LastPingTime time.Time // 最后ping时间
	CreatedAt    time.Time // 创建时间
}

// Control 管理WebSocket连接的控制结构体
type Control struct {
	logger          *zap.SugaredLogger
	config          *config.HttpServerConfig
	connections     map[string]*Connection
	connectionsMux  sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	once            sync.Once
	onConnected     func(conn *Connection)
	onDisconnected  func(conn *Connection)
	onMessage       func(conn *Connection, message []byte)
	onTextMessage   func(conn *Connection, message []byte)
	onBinaryMessage func(conn *Connection, message []byte)
	onPingMessage   func(conn *Connection, message []byte)
	onPongMessage   func(conn *Connection, message []byte)
	onCloseMessage  func(conn *Connection, message []byte)
	messageHandlers map[int]MessageHandler // 消息类型处理器映射
	upgrader        websocket.Upgrader
	pingInterval    time.Duration // ping间隔
	pongWait        time.Duration // pong等待时间
	writeWait       time.Duration // 写入等待时间
	maxMessageSize  int64         // 最大消息大小
}

// NewWebsocketControl 创建一个新的WebSocket控制器
// @param serverConfig *config.HttpServerConfig HTTP服务配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control WebSocket控制器
func NewWebsocketControl(serverConfig *config.HttpServerConfig, loggerControl *logger.Control) *Control {
	if serverConfig == nil {
		serverConfig = getDefaultConfig()
	}
	if !serverConfig.Enabled {
		return nil
	}
	wsLogger := loggerControl.GenLogger(logger.ModuleWebSocket)
	wsLogger.Infof("[control] starting new websocket control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		HandshakeTimeout: 10 * time.Second,
	}

	// 初始化WebsocketControl
	wc := &Control{
		logger:          wsLogger,
		config:          serverConfig,
		connections:     make(map[string]*Connection),
		ctx:             ctx,
		cancel:          cancel,
		upgrader:        upgrader,
		messageHandlers: make(map[int]MessageHandler),
		pingInterval:    54 * time.Second, // 心跳间隔
		pongWait:        60 * time.Second, // pong等待时间
		writeWait:       10 * time.Second, // 写入等待时间
		maxMessageSize:  512,              // 最大消息大小
	}

	wsLogger.Infof("[control] websocket control started successfully")
	return wc
}

// UpgradeConnection 将HTTP连接升级为WebSocket连接
// @param w http.ResponseWriter HTTP响应写入器
// @param r *http.Request HTTP请求
// @return string 连接ID
// @return *Connection 封装后的连接对象
// @return error 升级过程中的错误
func (wc *Control) UpgradeConnection(w http.ResponseWriter, r *http.Request) (string, *Connection, error) {
	// 执行HTTP到WebSocket的升级
	conn, err := wc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wc.logger.Debugf("[control] websocket upgrade error: %v", err)
		return "", nil, err
	}

	// 生成连接ID
	connID := uuid.GetUUID()

	// 添加连接到控制器
	connection := wc.AddConnection(connID, conn)

	wc.logger.Debugf("[control] websocket connection upgraded successfully, id: %s", connID)
	return connID, connection, nil
}

// SetOnConnected 设置连接建立时的回调函数
// @param callback func(conn *Connection) 连接建立时的回调函数
func (wc *Control) SetOnConnected(callback func(conn *Connection)) {
	wc.onConnected = callback
}

// SetOnDisconnected 设置连接关闭时的回调函数
// @param callback func(conn *Connection) 连接关闭时的回调函数
func (wc *Control) SetOnDisconnected(callback func(conn *Connection)) {
	wc.onDisconnected = callback
}

// SetOnMessage 设置收到消息时的回调函数
// @param callback func(conn *Connection, message []byte) 收到消息时的回调函数
func (wc *Control) SetOnMessage(callback func(conn *Connection, message []byte)) {
	wc.onMessage = callback
}

// SetOnTextMessage 设置收到文本消息时的回调函数
// @param callback func(conn *Connection, message []byte) 收到文本消息时的回调函数
func (wc *Control) SetOnTextMessage(callback func(conn *Connection, message []byte)) {
	wc.onTextMessage = callback
}

// SetOnBinaryMessage 设置收到二进制消息时的回调函数
// @param callback func(conn *Connection, message []byte) 收到二进制消息时的回调函数
func (wc *Control) SetOnBinaryMessage(callback func(conn *Connection, message []byte)) {
	wc.onBinaryMessage = callback
}

// SetOnPingMessage 设置收到ping消息时的回调函数
// @param callback func(conn *Connection, message []byte) 收到ping消息时的回调函数
func (wc *Control) SetOnPingMessage(callback func(conn *Connection, message []byte)) {
	wc.onPingMessage = callback
}

// SetOnPongMessage 设置收到pong消息时的回调函数
// @param callback func(conn *Connection, message []byte) 收到pong消息时的回调函数
func (wc *Control) SetOnPongMessage(callback func(conn *Connection, message []byte)) {
	wc.onPongMessage = callback
}

// SetOnCloseMessage 设置收到关闭消息时的回调函数
// @param callback func(conn *Connection, message []byte) 收到关闭消息时的回调函数
func (wc *Control) SetOnCloseMessage(callback func(conn *Connection, message []byte)) {
	wc.onCloseMessage = callback
}

// RegisterMessageHandler 注册特定类型的消息处理器
// @param messageType int 消息类型
// @param handler MessageHandler 消息处理器
func (wc *Control) RegisterMessageHandler(messageType int, handler MessageHandler) {
	wc.messageHandlers[messageType] = handler
	wc.logger.Debugf("[control] registered message handler for type: %d", messageType)
}

// AddConnection 添加一个新的WebSocket连接
// @param id 连接ID
// @param conn WebSocket连接
// @return *Connection 封装后的连接对象
func (wc *Control) AddConnection(id string, conn *websocket.Conn) *Connection {
	wc.connectionsMux.Lock()
	defer wc.connectionsMux.Unlock()

	// 创建新的连接对象
	connection := &Connection{
		Conn:         conn,
		ID:           id,
		Send:         make(chan []byte, 100),
		CloseChan:    make(chan struct{}),
		LastPingTime: time.Now(),
		CreatedAt:    time.Now(),
	}

	// 添加到连接映射中
	wc.connections[id] = connection

	// 启动读写协程
	go wc.readPump(connection)
	go wc.writePump(connection)

	// 调用连接建立回调
	if wc.onConnected != nil {
		go wc.onConnected(connection)
	}

	wc.logger.Debugf("[control] websocket connection added, id: %s", id)
	return connection
}

// RemoveConnection 移除一个WebSocket连接
// @param id 连接ID
func (wc *Control) RemoveConnection(id string) {
	wc.connectionsMux.Lock()
	connection, exists := wc.connections[id]
	if exists {
		delete(wc.connections, id)
		close(connection.CloseChan)
	}
	wc.connectionsMux.Unlock()

	if exists {
		wc.logger.Debugf("[control] websocket connection removed, id: %s", id)
		// 调用连接关闭回调
		if wc.onDisconnected != nil {
			go wc.onDisconnected(connection)
		}
	}
}

// GetConnection 获取指定ID的WebSocket连接
// @param id 连接ID
// @return *Connection 连接对象，如果不存在则返回nil
func (wc *Control) GetConnection(id string) *Connection {
	wc.connectionsMux.RLock()
	defer wc.connectionsMux.RUnlock()
	return wc.connections[id]
}

// Broadcast 向所有连接广播消息
// @param message 消息内容
func (wc *Control) Broadcast(message []byte) {
	wc.connectionsMux.RLock()
	// 创建连接副本，避免在遍历过程中锁定时间过长
	connections := make([]*Connection, 0, len(wc.connections))
	for _, conn := range wc.connections {
		connections = append(connections, conn)
	}
	wc.connectionsMux.RUnlock()

	// 向所有连接发送消息
	for _, conn := range connections {
		select {
		case conn.Send <- message:
			// 消息成功发送到通道
		default:
			// 通道已满，关闭连接
			close(conn.CloseChan)
			wc.logger.Debugf("[control] websocket connection buffer full, closing: %s", conn.ID)
		}
	}
}

// SendTo 向指定连接发送消息
// @param id 连接ID
// @param message 消息内容
// @return bool 消息是否成功发送
func (wc *Control) SendTo(id string, message []byte) bool {
	wc.connectionsMux.RLock()
	connection, exists := wc.connections[id]
	wc.connectionsMux.RUnlock()

	if !exists {
		wc.logger.Debugf("[control] websocket connection not found, id: %s", id)
		return false
	}

	select {
	case connection.Send <- message:
		// 消息成功发送到通道
		return true
	default:
		// 通道已满，关闭连接
		close(connection.CloseChan)
		wc.logger.Debugf("[control] websocket connection buffer full, closing: %s", connection.ID)
		return false
	}
}

// SendPing 发送ping消息到指定连接
// @param id string 连接ID
// @param message []byte ping消息内容
// @return bool 是否成功发送
func (wc *Control) SendPing(id string, message []byte) bool {
	wc.connectionsMux.RLock()
	connection, exists := wc.connections[id]
	wc.connectionsMux.RUnlock()

	if !exists {
		wc.logger.Debugf("[control] websocket connection not found for ping, id: %s", id)
		return false
	}

	connection.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait))
	err := connection.Conn.WriteMessage(websocket.PingMessage, message)
	if err != nil {
		wc.logger.Debugf("[control] failed to send ping to connection: %s, error: %v", id, err)
		return false
	}

	wc.logger.Debugf("[control] sent ping to connection: %s", id)
	return true
}

// SendPong 发送pong消息到指定连接
// @param id string 连接ID
// @param message []byte pong消息内容
// @return bool 是否成功发送
func (wc *Control) SendPong(id string, message []byte) bool {
	wc.connectionsMux.RLock()
	connection, exists := wc.connections[id]
	wc.connectionsMux.RUnlock()

	if !exists {
		wc.logger.Debugf("[control] websocket connection not found for pong, id: %s", id)
		return false
	}

	connection.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait))
	err := connection.Conn.WriteMessage(websocket.PongMessage, message)
	if err != nil {
		wc.logger.Debugf("[control] failed to send pong to connection: %s, error: %v", id, err)
		return false
	}

	wc.logger.Debugf("[control] sent pong to connection: %s", id)
	return true
}

// SendBinary 发送二进制消息到指定连接
// @param id string 连接ID
// @param message []byte 二进制消息内容
// @return bool 是否成功发送
func (wc *Control) SendBinary(id string, message []byte) bool {
	wc.connectionsMux.RLock()
	connection, exists := wc.connections[id]
	wc.connectionsMux.RUnlock()

	if !exists {
		wc.logger.Debugf("[control] websocket connection not found for binary message, id: %s", id)
		return false
	}

	select {
	case connection.Send <- message:
		// 消息成功发送到通道
		return true
	default:
		// 通道已满，关闭连接
		close(connection.CloseChan)
		wc.logger.Debugf("[control] websocket connection buffer full for binary message, closing: %s", connection.ID)
		return false
	}
}

// SendMessage 发送指定类型的消息到指定连接
// @param id string 连接ID
// @param messageType int 消息类型
// @param message []byte 消息内容
// @return bool 是否成功发送
func (wc *Control) SendMessage(id string, messageType int, message []byte) bool {
	switch messageType {
	case websocket.TextMessage:
		return wc.SendTo(id, message)
	case websocket.BinaryMessage:
		return wc.SendBinary(id, message)
	case websocket.PingMessage:
		return wc.SendPing(id, message)
	case websocket.PongMessage:
		return wc.SendPong(id, message)
	case websocket.CloseMessage:
		return wc.SendClose(id, string(message))
	default:
		wc.logger.Warnf("[control] unsupported message type: %d", messageType)
		return false
	}
}

// SendClose 发送关闭消息到指定连接
// @param id string 连接ID
// @param reason string 关闭原因
// @return bool 是否成功发送
func (wc *Control) SendClose(id string, reason string) bool {
	wc.connectionsMux.RLock()
	connection, exists := wc.connections[id]
	wc.connectionsMux.RUnlock()

	if !exists {
		wc.logger.Debugf("[control] websocket connection not found for close message, id: %s", id)
		return false
	}

	deadline := time.Now().Add(wc.writeWait)
	err := connection.Conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason),
		deadline,
	)
	if err != nil {
		wc.logger.Debugf("[control] failed to send close message to connection: %s, error: %v", id, err)
		return false
	}

	wc.logger.Debugf("[control] sent close message to connection: %s", id)
	return true
}

// GetConnectionByUserID 根据用户ID获取连接
// @param userID string 用户ID
// @return *Connection 连接对象，如果不存在则返回nill
func (wc *Control) GetConnectionByUserID(userID string) *Connection {
	wc.connectionsMux.RLock()
	defer wc.connectionsMux.RUnlock()

	for _, conn := range wc.connections {
		if conn.UserID == userID {
			return conn
		}
	}
	return nil
}

// GetConnectionsByUserID 根据用户ID获取所有连接（一个用户可能有多个连接）
// @param userID string 用户ID
// @return []*Connection 连接列表
func (wc *Control) GetConnectionsByUserID(userID string) []*Connection {
	wc.connectionsMux.RLock()
	defer wc.connectionsMux.RUnlock()

	var connections []*Connection
	for _, conn := range wc.connections {
		if conn.UserID == userID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// GetConnectionsByNodeID 根据节点ID获取所有连接
// @param nodeID string 节点ID
// @return []*Connection 连接列表
func (wc *Control) GetConnectionsByNodeID(nodeID string) []*Connection {
	wc.connectionsMux.RLock()
	defer wc.connectionsMux.RUnlock()

	var connections []*Connection
	for _, conn := range wc.connections {
		if conn.NodeID == nodeID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// BroadcastByUserID 向指定用户的所有连接广播消息
// @param userID string 用户ID
// @param message []byte 消息内容
// @return int 成功发送的连接数
func (wc *Control) BroadcastByUserID(userID string, message []byte) int {
	connections := wc.GetConnectionsByUserID(userID)
	successCount := 0

	for _, conn := range connections {
		if wc.SendTo(conn.ID, message) {
			successCount++
		}
	}

	wc.logger.Debugf("[control] broadcasted message to %d connections of user: %s", successCount, userID)
	return successCount
}

// BroadcastByNodeID 向指定节点的所有连接广播消息
// @param nodeID string 节点ID
// @param message []byte 消息内容
// @return int 成功发送的连接数
func (wc *Control) BroadcastByNodeID(nodeID string, message []byte) int {
	connections := wc.GetConnectionsByNodeID(nodeID)
	successCount := 0

	for _, conn := range connections {
		if wc.SendTo(conn.ID, message) {
			successCount++
		}
	}

	wc.logger.Debugf("[control] broadcasted message to %d connections of node: %s", successCount, nodeID)
	return successCount
}

// SetConnectionUserID 设置连接的用户ID
// @param id string 连接ID
// @param userID string 用户ID
// @return bool 是否设置成功
func (wc *Control) SetConnectionUserID(id string, userID string) bool {
	wc.connectionsMux.Lock()
	defer wc.connectionsMux.Unlock()

	if connection, exists := wc.connections[id]; exists {
		connection.UserID = userID
		wc.logger.Debugf("[control] set userID %s for connection: %s", userID, id)
		return true
	}
	return false
}

// SetConnectionNodeID 设置连接的节点ID
// @param id string 连接ID
// @param nodeID string 节点ID
// @return bool 是否设置成功
func (wc *Control) SetConnectionNodeID(id string, nodeID string) bool {
	wc.connectionsMux.Lock()
	defer wc.connectionsMux.Unlock()

	if connection, exists := wc.connections[id]; exists {
		connection.NodeID = nodeID
		wc.logger.Debugf("[control] set nodeID %s for connection: %s", nodeID, id)
		return true
	}
	return false
}

// StartHeartbeat 启动心跳检测
func (wc *Control) StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(wc.pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-wc.ctx.Done():
				return
			case <-ticker.C:
				wc.pingAllConnections()
			}
		}
	}()
	wc.logger.Infof("[control] heartbeat started with interval: %v", wc.pingInterval)
}

// pingAllConnections 向所有连接发送ping
func (wc *Control) pingAllConnections() {
	wc.connectionsMux.RLock()
	connections := make([]*Connection, 0, len(wc.connections))
	for _, conn := range wc.connections {
		connections = append(connections, conn)
	}
	wc.connectionsMux.RUnlock()

	for _, conn := range connections {
		// 检查连接是否超时
		if time.Since(conn.LastPingTime) > wc.pongWait {
			wc.logger.Debugf("[control] connection %s ping timeout, removing", conn.ID)
			wc.RemoveConnection(conn.ID)
			continue
		}

		// 发送ping
		wc.SendPing(conn.ID, []byte("ping"))
	}
}

// @return int 连接数
func (wc *Control) GetConnectionCount() int {
	wc.connectionsMux.RLock()
	defer wc.connectionsMux.RUnlock()
	return len(wc.connections)
}

// readPump 从WebSocket连接读取消息（内部方法）
// @param conn *Connection WebSocket连接对象
func (wc *Control) readPump(conn *Connection) {
	defer func() {
		wc.RemoveConnection(conn.ID)
		conn.Conn.Close()
	}()

	// 设置读取限制和超时
	conn.Conn.SetReadLimit(wc.maxMessageSize)
	conn.Conn.SetReadDeadline(time.Now().Add(wc.pongWait))

	// 设置pong处理器（用于心跳检测）
	conn.Conn.SetPongHandler(func(string) error {
		conn.Conn.SetReadDeadline(time.Now().Add(wc.pongWait))
		conn.LastPingTime = time.Now()
		wc.logger.Debugf("[control] received pong from connection: %s", conn.ID)
		return nil
	})

	// 设置ping处理器
	conn.Conn.SetPingHandler(func(message string) error {
		conn.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait))
		if err := conn.Conn.WriteMessage(websocket.PongMessage, []byte(message)); err != nil {
			wc.logger.Debugf("[control] failed to send pong: %v", err)
			return err
		}
		wc.logger.Debugf("[control] sent pong to connection: %s", conn.ID)
		return nil
	})

	// 设置关闭处理器
	conn.Conn.SetCloseHandler(func(code int, text string) error {
		wc.logger.Debugf("[control] received close message from connection: %s, code: %d, text: %s", conn.ID, code, text)
		return nil
	})

	for {
		select {
		case <-wc.ctx.Done():
			// 控制器已关闭
			wc.logger.Debugf("[control] context cancelled, closing read pump for connection: %s", conn.ID)
			return
		case <-conn.CloseChan:
			// 连接已关闭
			wc.logger.Debugf("[control] connection close channel triggered for: %s", conn.ID)
			return
		default:
			// 读取消息
			messageType, message, err := conn.Conn.ReadMessage()
			if err != nil {
				// 检查是否是正常关闭
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					wc.logger.Debugf("[control] websocket unexpected close error: %v, connection: %s", err, conn.ID)
				} else {
					wc.logger.Debugf("[control] websocket read error: %v, closing connection: %s", err, conn.ID)
				}
				return
			}

			// 处理不同类型的消息
			wc.handleReceivedMessage(conn, messageType, message)
		}
	}
}

// handleReceivedMessage 处理接收到的消息（内部方法）
// @param conn *Connection WebSocket连接对象
// @param messageType int 消息类型
// @param message []byte 消息内容
func (wc *Control) handleReceivedMessage(conn *Connection, messageType int, message []byte) {
	wc.logger.Debugf("[control] handling message type: %d, from connection: %s, content length: %d", messageType, conn.ID, len(message))

	// 首先调用通用消息回调
	if wc.onMessage != nil {
		go wc.onMessage(conn, message)
	}

	// 检查是否有注册的特定类型处理器
	if handler, exists := wc.messageHandlers[messageType]; exists {
		go func() {
			if err := handler.HandleMessage(conn, messageType, message); err != nil {
				wc.logger.Errorf("[control] message handler error: %v, connection: %s", err, conn.ID)
			}
		}()
	}

	// 根据消息类型调用不同的回调函数
	switch messageType {
	case websocket.TextMessage:
		wc.logger.Debugf("[control] received text message from connection: %s", conn.ID)
		if wc.onTextMessage != nil {
			go wc.onTextMessage(conn, message)
		}
		wc.handleTextMessage(conn, message)

	case websocket.BinaryMessage:
		wc.logger.Debugf("[control] received binary message from connection: %s", conn.ID)
		if wc.onBinaryMessage != nil {
			go wc.onBinaryMessage(conn, message)
		}
		wc.handleBinaryMessage(conn, message)

	case websocket.PingMessage:
		wc.logger.Debugf("[control] received ping message from connection: %s", conn.ID)
		if wc.onPingMessage != nil {
			go wc.onPingMessage(conn, message)
		}
		// 自动回复pong
		wc.SendPong(conn.ID, message)

	case websocket.PongMessage:
		wc.logger.Debugf("[control] received pong message from connection: %s", conn.ID)
		if wc.onPongMessage != nil {
			go wc.onPongMessage(conn, message)
		}
		conn.LastPingTime = time.Now()

	case websocket.CloseMessage:
		wc.logger.Debugf("[control] received close message from connection: %s", conn.ID)
		if wc.onCloseMessage != nil {
			go wc.onCloseMessage(conn, message)
		}

	default:
		wc.logger.Warnf("[control] unknown message type: %d from connection: %s", messageType, conn.ID)
	}
}

// handleTextMessage 处理文本消息
func (wc *Control) handleTextMessage(conn *Connection, message []byte) {
	// 尝试解析JSON消息
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		// 如果不是JSON格式，作为普通文本处理
		wc.logger.Debugf("[control] received plain text message from %s: %s", conn.ID, string(message))
		return
	}

	// 如果是结构化消息，进行进一步处理
	wc.logger.Debugf("[control] received structured message from %s: type=%d, from=%s, to=%s",
		conn.ID, msg.Type, msg.From, msg.To)

	// 处理不同类型的结构化消息
	switch msg.Type {
	case 1: // 心跳消息
		wc.handleHeartbeat(conn)
	case 2: // 聊天消息
		wc.handleChatMessage(conn, &msg)
	case 3: // 系统消息
		wc.handleSystemMessage(conn, &msg)
	default:
		wc.logger.Debugf("[control] unknown structured message type: %d", msg.Type)
	}
}

// handleBinaryMessage 处理二进制消息
func (wc *Control) handleBinaryMessage(conn *Connection, message []byte) {
	wc.logger.Debugf("[control] handling binary message from %s, size: %d bytes", conn.ID, len(message))
	// 这里可以添加二进制消息的特殊处理逻辑
	// 例如：文件传输、图片传输等
}

// handleHeartbeat 处理心跳消息
func (wc *Control) handleHeartbeat(conn *Connection) {
	conn.LastPingTime = time.Now()
	// 回复心跳
	heartbeatResp := Response{
		Success: true,
		Message: "heartbeat",
		Data:    time.Now().Unix(),
	}
	if data, err := json.Marshal(heartbeatResp); err == nil {
		wc.SendTo(conn.ID, data)
	}
}

// handleChatMessage 处理聊天消息
func (wc *Control) handleChatMessage(conn *Connection, msg *Message) {
	wc.logger.Debugf("[control] handling chat message from %s to %s", msg.From, msg.To)

	// 如果指定了接收方，进行点对点发送
	if msg.To != "" && msg.To != "all" {
		// 查找目标连接
		targetConn := wc.GetConnectionByUserID(msg.To)
		if targetConn != nil {
			wc.SendTo(targetConn.ID, msg.Content)
		} else {
			// 目标用户不在线，返回错误
			errorResp := Response{
				Success: false,
				Message: "target user is offline",
			}
			if data, err := json.Marshal(errorResp); err == nil {
				wc.SendTo(conn.ID, data)
			}
		}
	} else {
		// 广播消息
		wc.Broadcast(msg.Content)
	}
}

// handleSystemMessage 处理系统消息
func (wc *Control) handleSystemMessage(conn *Connection, msg *Message) {
	wc.logger.Debugf("[control] handling system message from %s", msg.From)
	// 这里可以处理系统级消息，如：
	// - 用户状态更新
	// - 系统通知
	// - 配置更新等
}

// writePump 向WebSocket连接写入消息（内部方法）
// @param conn *Connection WebSocket连接对象
func (wc *Control) writePump(conn *Connection) {
	defer func() {
		conn.Conn.Close()
	}()

	// 启动心跳ticker
	ticker := time.NewTicker(wc.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-wc.ctx.Done():
			// 控制器已关闭
			conn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		case <-conn.CloseChan:
			// 连接已关闭
			conn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		case message, ok := <-conn.Send:
			// 发送消息
			if !ok {
				// 发送通道已关闭
				conn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}

			// 设置写入超时
			conn.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait))

			// 写入文本消息（默认）
			err := conn.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				// 写入错误，关闭连接
				wc.logger.Debugf("[control] websocket write error: %v, closing connection: %s", err, conn.ID)
				return
			}
		case <-ticker.C:
			// 心跳检测
			conn.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				wc.logger.Debugf("[control] ping error: %v, closing connection: %s", err, conn.ID)
				return
			}
			wc.logger.Debugf("[control] sent ping to connection: %s", conn.ID)
		}
	}
}

// StartUp 启动WebSocket控制器
// @param failedFunc func(err error) 启动失败时的回调函数
func (wc *Control) StartUp(failedFunc func(err error)) {
	wc.once.Do(func() {
		if wc.config.Enabled {
			wc.logger.Infof("[control] websocket control started...")
			// 启动心跳检测
			wc.StartHeartbeat()
		}
	})
}

// Shutdown 关闭WebSocket控制器
// @return error 关闭过程中的错误
func (wc *Control) Shutdown() error {
	wc.logger.Infof("[control] shutting down websocket control...")

	// 取消上下文
	wc.cancel()

	// 关闭所有连接
	wc.connectionsMux.Lock()
	for _, conn := range wc.connections {
		close(conn.CloseChan)
	}
	wc.connectionsMux.Unlock()

	wc.logger.Infof("[control] websocket control shutdown completed")
	return nil
}
