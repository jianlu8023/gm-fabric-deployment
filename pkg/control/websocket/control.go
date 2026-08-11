package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"

	"github.com/gorilla/websocket"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

// Control 管理WebSocket连接的控制结构体
//
// @description WebSocket子控制器，负责连接升级、读写循环、心跳保活、消息分发与广播。
// @description 所有对底层websocket.Conn的写入统一由writePump串行执行，其他goroutine通过Send通道投递消息。
// @struct
type Control struct {
	logger          *zap.SugaredLogger
	config          *config.WebSocketConfig
	connections     concurrent.Map[string, *Connection]
	ctx             context.Context
	cancel          context.CancelFunc
	once            sync.Once
	onConnected     func(conn *Connection)
	onDisconnected  func(conn *Connection)
	onMessage       func(conn *Connection, message Message)
	onTextMessage   func(conn *Connection, message Message)
	onBinaryMessage func(conn *Connection, message Message)
	onPingMessage   func(conn *Connection, message Message)
	onPongMessage   func(conn *Connection, message Message)
	onCloseMessage  func(conn *Connection, message Message)
	callbackMux     sync.RWMutex                        // 保护on*回调函数的读写
	messageHandlers concurrent.Map[int, MessageHandler] // 消息类型处理器映射，并发安全
	upgrader        websocket.Upgrader
	pingInterval    time.Duration // ping间隔
	pongWait        time.Duration // pong等待时间
	writeWait       time.Duration // 写入等待时间
	maxMessageSize  int64         // 最大消息大小
	sendBufferSize  int           // 每连接发送通道缓冲大小
}

// NewWebsocketControl 创建一个新的WebSocket控制器
//
// @param wsConfig *config.WebSocketConfig WebSocket配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control WebSocket控制器，若配置未启用则返回nil
func NewWebsocketControl(wsConfig *config.WebSocketConfig, loggerControl *logger.Control) *Control {
	wsConfig = check.IF[*config.WebSocketConfig](wsConfig == nil,
		getDefaultConfig(),
		wsConfig,
	)
	if !wsConfig.Enabled {
		return nil
	}

	// 参数兜底，避免配置缺失时使用零值
	if wsConfig.PingInterval <= 0 {
		wsConfig.PingInterval = 54 * time.Second
	}
	if wsConfig.PongWait <= 0 {
		wsConfig.PongWait = 60 * time.Second
	}
	if wsConfig.WriteWait <= 0 {
		wsConfig.WriteWait = 10 * time.Second
	}
	if wsConfig.MaxMessageSize <= 0 {
		wsConfig.MaxMessageSize = 4096
	}
	if wsConfig.SendBufferSize <= 0 {
		wsConfig.SendBufferSize = 256
	}
	readBufferSize := wsConfig.ReadBufferSize
	if readBufferSize <= 0 {
		readBufferSize = 1024
	}
	writeBufferSize := wsConfig.WriteBufferSize
	if writeBufferSize <= 0 {
		writeBufferSize = 1024
	}
	handshakeTimeout := wsConfig.HandshakeTimeout
	if handshakeTimeout <= 0 {
		handshakeTimeout = 10 * time.Second
	}

	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	// wsLogger := loggerControl.GenLogger(logger.ModuleWebSocket)
	wsLogger := loggerControl.GenLogger("")
	wsLogger.Infof("[websocket/control] starting new websocket control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 构造Origin校验函数：白名单为空时放行（开发环境），非空时严格匹配
	checkOrigin := func(r *http.Request) bool {
		if len(wsConfig.CheckOrigins) == 0 {
			return true
		}
		origin := r.Header.Get("Origin")
		for _, allowed := range wsConfig.CheckOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	}

	// 初始化upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:   readBufferSize,
		WriteBufferSize:  writeBufferSize,
		CheckOrigin:      checkOrigin,
		HandshakeTimeout: handshakeTimeout,
	}

	// 初始化WebsocketControl
	wc := &Control{
		logger:          wsLogger,
		config:          wsConfig,
		connections:     concurrentmap.NewRWMap[string, *Connection](),
		ctx:             ctx,
		cancel:          cancel,
		upgrader:        upgrader,
		messageHandlers: concurrentmap.NewRWMap[int, MessageHandler](),
		pingInterval:    wsConfig.PingInterval,
		pongWait:        wsConfig.PongWait,
		writeWait:       wsConfig.WriteWait,
		maxMessageSize:  wsConfig.MaxMessageSize,
		sendBufferSize:  wsConfig.SendBufferSize,
	}

	wc.registerDefaultCallback()

	wsLogger.Infof("[websocket/control] websocket control started successfully")
	return wc
}

// StartUp 启动WebSocket控制器
//
// @param failedFunc func(err error) 启动失败时的回调函数
func (wc *Control) StartUp(failedFunc func(err error)) {
	wc.once.Do(func() {
		if wc.config.Enabled {
			wc.logger.Infof("[websocket/control] websocket control started...")
			// 启动超时扫描（仅清理超时连接，ping由writePump per-connection发送）
			wc.startHeartBeatScan()
		}
	})
}

// Shutdown 关闭WebSocket控制器
//
// @description 取消上下文并关闭所有连接，使用Connection.Close()的sync.Once保证不会重复关闭
// @return error 关闭过程中的错误
func (wc *Control) Shutdown() error {
	wc.logger.Infof("[websocket/control] shutting down websocket control...")

	// 取消上下文，通知所有读写循环退出
	wc.cancel()

	// 关闭所有连接的CloseChan（Connection.Close内部用sync.Once保护）
	values := wc.connections.Values()
	for _, conn := range values {
		conn.Close()
	}

	wc.logger.Infof("[websocket/control] websocket control shutdown completed")
	return nil
}

// SetOnConnected 设置连接建立时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection) 连接建立时的回调函数
func (wc *Control) SetOnConnected(callback func(conn *Connection)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onConnected = callback
}

// SetOnDisconnected 设置连接关闭时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection) 连接关闭时的回调函数
func (wc *Control) SetOnDisconnected(callback func(conn *Connection)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onDisconnected = callback
}

// SetOnMessage 设置收到消息时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection, message Message) 收到消息时的回调函数
func (wc *Control) SetOnMessage(callback func(conn *Connection, message Message)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onMessage = callback
}

// SetOnTextMessage 设置收到文本消息时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection, message Message) 收到文本消息时的回调函数
func (wc *Control) SetOnTextMessage(callback func(conn *Connection, message Message)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onTextMessage = callback
}

// SetOnBinaryMessage 设置收到二进制消息时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection, message Message) 收到二进制消息时的回调函数
func (wc *Control) SetOnBinaryMessage(callback func(conn *Connection, message Message)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onBinaryMessage = callback
}

// SetOnPingMessage 设置收到ping消息时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection, message Message) 收到ping消息时的回调函数
func (wc *Control) SetOnPingMessage(callback func(conn *Connection, message Message)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onPingMessage = callback
}

// SetOnPongMessage 设置收到pong消息时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection, message Message) 收到pong消息时的回调函数
func (wc *Control) SetOnPongMessage(callback func(conn *Connection, message Message)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onPongMessage = callback
}

// SetOnCloseMessage 设置收到关闭消息时的回调函数
//
// @description 必须在StartUp前调用，避免与读循环并发
// @param callback func(conn *Connection, message Message) 收到关闭消息时的回调函数
func (wc *Control) SetOnCloseMessage(callback func(conn *Connection, message Message)) {
	wc.callbackMux.Lock()
	defer wc.callbackMux.Unlock()
	wc.onCloseMessage = callback
}

// registerDefaultCallback 注册默认回调（仅打印日志）
func (wc *Control) registerDefaultCallback() {
	wc.logger.Debugf("[websocket/control] registering default callback...")
	wc.SetOnConnected(func(conn *Connection) {
		wc.logger.Debugf("[websocket/control] websocket connection %s connected", conn.NodeID)
	})
	wc.SetOnDisconnected(func(conn *Connection) {
		wc.logger.Debugf("[websocket/control] websocket connection %s disconnected", conn.NodeID)
	})
	wc.SetOnMessage(func(conn *Connection, msg Message) {
		wc.logger.Debugf("[websocket/control] websocket connection %s received message: %s", conn.NodeID, string(msg.Content))
	})
	wc.SetOnTextMessage(func(conn *Connection, msg Message) {
		wc.logger.Debugf("[websocket/control] websocket connection %s received text message: %s", conn.NodeID, string(msg.Content))
	})
	wc.SetOnBinaryMessage(func(conn *Connection, msg Message) {
		wc.logger.Debugf("[websocket/control] websocket connection %s received binary message, size: %d", conn.NodeID, len(msg.Content))
	})
	wc.SetOnPingMessage(func(conn *Connection, msg Message) {
		wc.logger.Debugf("[websocket/control] websocket connection %s received ping message", conn.NodeID)
	})
	wc.SetOnPongMessage(func(conn *Connection, msg Message) {
		wc.logger.Debugf("[websocket/control] websocket connection %s received pong message", conn.NodeID)
	})
	wc.SetOnCloseMessage(func(conn *Connection, message Message) {
		wc.logger.Debugf("[websocket/control] websocket connection %s received close message", conn.NodeID)
	})
}

// UpgradeConnection 将HTTP连接升级为WebSocket连接
//
// @param w http.ResponseWriter HTTP响应写入器
// @param r *http.Request HTTP请求
// @param connID string 连接ID，为空时自动生成UUID
// @param userID string 关联用户ID，由handler层认证后传入
// @param sessionID string 关联会话ID
// @return string 连接ID
// @return *Connection 封装后的连接对象
// @return error 升级过程中的错误
func (wc *Control) UpgradeConnection(w http.ResponseWriter, r *http.Request, connID string, userID string, sessionID string) (string, *Connection, error) {
	// 执行HTTP到WebSocket的升级
	conn, err := wc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wc.logger.Debugf("[websocket/control] websocket upgrade error: %v", err)
		return "", nil, err
	}

	if stringer.IsBlank(connID) {
		// 生成连接ID
		connID = uuid.GetUUID()
	}

	// 添加连接到控制器
	connection := wc.AddConnection(connID, conn, userID, sessionID, r.RemoteAddr, r.UserAgent())

	wc.logger.Debugf("[websocket/control] websocket connection upgraded successfully, id: %s", connID)
	return connID, connection, nil
}

// RegisterMessageHandler 注册特定类型的消息处理器
//
// @description 按消息类型注册处理器，必须在StartUp前调用；运行时并发安全
// @param messageType int 消息类型
// @param handler MessageHandler 消息处理器
func (wc *Control) RegisterMessageHandler(messageType int, handler MessageHandler) {
	wc.messageHandlers.Put(messageType, handler)
	wc.logger.Debugf("[websocket/control] registered message handler for type: %d", messageType)
}

// AddConnection 添加一个新的WebSocket连接
//
// @param id string 连接ID
// @param conn *websocket.Conn 底层WebSocket连接
// @param userID string 关联用户ID
// @param sessionID string 关联会话ID
// @param remoteAddr string 客户端地址
// @param userAgent string 客户端User-Agent
// @return *Connection 封装后的连接对象
func (wc *Control) AddConnection(id string, conn *websocket.Conn, userID string, sessionID string, remoteAddr string, userAgent string) *Connection {
	// 创建新的连接对象
	connection := &Connection{
		Conn:         conn,
		NodeID:       id,
		UserID:       userID,
		SessionID:    sessionID,
		RemoteAddr:   remoteAddr,
		UserAgent:    userAgent,
		Send:         make(chan Message, wc.sendBufferSize),
		CloseChan:    make(chan struct{}),
		LastPongTime: time.Now(),
		CreatedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// 添加到连接映射中
	wc.connections.Put(id, connection)

	// 启动读写协程
	go wc.readPump(connection)
	go wc.writePump(connection)

	// 调用连接建立回调
	wc.callbackMux.RLock()
	onConnected := wc.onConnected
	wc.callbackMux.RUnlock()
	if onConnected != nil {
		go onConnected(connection)
	}

	wc.logger.Debugf("[websocket/control] websocket connection added, id: %s", id)
	return connection
}

// RemoveConnection 移除一个WebSocket连接
//
// @description 从连接映射中删除并关闭CloseChan，Connection.Close内部用sync.Once保证幂等
// @param id string 连接ID
func (wc *Control) RemoveConnection(id string) {
	connection, exists := wc.connections.Get(id)
	if !exists {
		return
	}
	wc.connections.Del(id)

	// 关闭CloseChan（sync.Once保护，重复调用安全）
	connection.Close()

	wc.logger.Debugf("[websocket/control] websocket connection removed, id: %s", id)

	// 调用连接关闭回调
	wc.callbackMux.RLock()
	onDisconnected := wc.onDisconnected
	wc.callbackMux.RUnlock()
	if onDisconnected != nil {
		go onDisconnected(connection)
	}
}

// GetConnection 获取指定ID的WebSocket连接
//
// @param id string 连接ID
// @return *Connection 连接对象，如果不存在则返回nil
func (wc *Control) GetConnection(id string) *Connection {
	connection, exists := wc.connections.Get(id)
	if !exists {
		return nil
	}
	return connection
}

// Broadcast 向所有连接广播消息
//
// @description 向所有连接的Send通道投递消息，缓冲满则关闭对应连接；不排除任何连接
// @param message Message 消息内容
func (wc *Control) Broadcast(message Message) {
	wc.broadcastExcept("", message)
}

// BroadcastExcept 向除指定连接外的所有连接广播消息
//
// @description 用于群聊场景避免回显发送者
// @param excludeID string 需要排除的连接ID，为空时不排除任何连接
// @param message Message 消息内容
func (wc *Control) BroadcastExcept(excludeID string, message Message) {
	wc.broadcastExcept(excludeID, message)
}

// broadcastExcept 内部广播实现
//
// @param excludeID string 需要排除的连接ID
// @param message Message 消息内容
func (wc *Control) broadcastExcept(excludeID string, message Message) {
	// 填充时间戳
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	iter := wc.connections.Iterator()
	defer func() {
		_ = iter.Close()
	}()

	var fullIDs []string
	for iter.HasNext() {
		next := iter.Value()
		conn := next.Value
		if conn.NodeID == excludeID {
			continue
		}
		select {
		case conn.Send <- message:
		default:
			// 缓冲满，记录待清理；不能在迭代过程中删除，避免影响迭代器
			wc.logger.Debugf("[websocket/control] websocket connection buffer full, will close: %s", conn.NodeID)
			fullIDs = append(fullIDs, conn.NodeID)
		}
	}

	// 迭代结束后统一清理缓冲满的连接
	for _, id := range fullIDs {
		wc.RemoveConnection(id)
	}
}

// SendMessage 发送消息到指定连接
//
// @description 向目标连接的Send通道投递消息，由目标连接的writePump串行写入；
// @description 缓冲满时返回ErrSendBufferFull，不阻塞调用方
// @param id string 目标连接ID
// @param message Message 消息内容
// @return error 发送失败错误（连接不存在或缓冲满）
func (wc *Control) SendMessage(id string, message Message) error {
	connection, exists := wc.connections.Get(id)
	if !exists {
		wc.logger.Debugf("[websocket/control] websocket connection not found, id: %s", id)
		return ErrConnectionNotFound
	}

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	select {
	case connection.Send <- message:
		return nil
	default:
		wc.logger.Debugf("[websocket/control] websocket connection send buffer full, id: %s", id)
		return ErrSendBufferFull
	}
}

// startHeartBeatScan 启动心跳超时扫描
//
// @description 仅扫描并清理pong超时的连接，ping由每个连接的writePump周期性发送
func (wc *Control) startHeartBeatScan() {
	go func() {
		ticker := time.NewTicker(wc.pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-wc.ctx.Done():
				return
			case <-ticker.C:
				wc.scanTimeoutConnections()
			}
		}
	}()
	wc.logger.Infof("[websocket/control] heartbeat scan started with interval: %v", wc.pingInterval)
}

// scanTimeoutConnections 扫描并清理pong超时的连接
//
// @description 先收集超时连接ID，迭代结束后批量RemoveConnection，避免迭代中删除
func (wc *Control) scanTimeoutConnections() {
	iter := wc.connections.Iterator()
	defer func() {
		_ = iter.Close()
	}()

	var timeoutIDs []string
	for iter.HasNext() {
		next := iter.Value()
		conn := next.Value
		if time.Since(conn.LastPongTime) > wc.pongWait {
			wc.logger.Debugf("[websocket/control] connection %s pong timeout, removing", conn.NodeID)
			timeoutIDs = append(timeoutIDs, conn.NodeID)
		}
	}

	for _, id := range timeoutIDs {
		wc.RemoveConnection(id)
	}
}

// GetConnectionCount 获取当前连接数
//
// @return int 当前连接数
func (wc *Control) GetConnectionCount() int {
	return wc.connections.Len()
}

// GetAllConnections 获取所有连接
//
// @description 返回当前所有连接的快照切片
// @return []*Connection 连接切片
func (wc *Control) GetAllConnections() []*Connection {
	wc.logger.Debugf("[websocket/control] getting all connections...")
	return wc.connections.Values()
}

// readPump 从WebSocket连接读取消息（内部方法）
//
// @description 每连接一个goroutine，阻塞读ReadMessage；通过SetReadDeadline让读超时自动解除阻塞，
// @description 配合CloseChan关闭与ctx取消实现退出。退出时负责关闭底层conn与清理连接。
// @param conn *Connection WebSocket连接对象
func (wc *Control) readPump(conn *Connection) {
	defer func() {
		// 只在readPump退出时关闭底层连接，避免与writePump双重关闭
		wc.RemoveConnection(conn.NodeID)
		if err := conn.Conn.Close(); err != nil {
			wc.logger.Errorf("[websocket/control] failed to close connection %s: %v", conn.NodeID, err)
		}
	}()

	// 设置读取限制和初始超时
	conn.Conn.SetReadLimit(wc.maxMessageSize)
	if err := conn.Conn.SetReadDeadline(time.Now().Add(wc.pongWait)); err != nil {
		wc.logger.Errorf("[websocket/control] failed to set read deadline for connection %s: %v", conn.NodeID, err)
		return
	}

	// 设置pong处理器（用于心跳保活：收到pong时刷新读超时与LastPongTime）
	conn.Conn.SetPongHandler(func(string) error {
		if err := conn.Conn.SetReadDeadline(time.Now().Add(wc.pongWait)); err != nil {
			wc.logger.Errorf("[websocket/control] failed to set read deadline for connection %s: %v", conn.NodeID, err)
			return err
		}
		conn.LastPongTime = time.Now()
		wc.logger.Debugf("[websocket/control] received pong from connection: %s", conn.NodeID)
		return nil
	})

	// 设置ping处理器：收到客户端ping时自动回pong
	conn.Conn.SetPingHandler(func(message string) error {
		if err := conn.Conn.SetReadDeadline(time.Now().Add(wc.pongWait)); err != nil {
			wc.logger.Errorf("[websocket/control] failed to set read deadline for connection %s: %v", conn.NodeID, err)
			return err
		}
		// pong通过Send通道由writePump串行写入，避免并发写
		select {
		case conn.Send <- Message{Type: websocket.PongMessage, Content: []byte(message), Timestamp: time.Now()}:
		default:
			wc.logger.Debugf("[websocket/control] send buffer full while handling ping, id: %s", conn.NodeID)
		}
		wc.logger.Debugf("[websocket/control] received ping from connection: %s", conn.NodeID)
		return nil
	})

	// 设置关闭处理器
	conn.Conn.SetCloseHandler(func(code int, text string) error {
		wc.logger.Debugf("[websocket/control] received close message from connection: %s, code: %d, text: %s", conn.NodeID, code, text)
		return nil
	})

	for {
		// 读取消息（阻塞）；通过SetReadDeadline超时或连接关闭解除阻塞
		messageType, message, err := conn.Conn.ReadMessage()
		if err != nil {
			// 检查是否是正常关闭
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				wc.logger.Debugf("[websocket/control] websocket unexpected close error: %v, connection: %s", err, conn.NodeID)
			} else {
				// 读超时也走这里：若ctx未取消且CloseChan未关闭，说明是pong超时
				select {
				case <-wc.ctx.Done():
					wc.logger.Debugf("[websocket/control] context cancelled, closing read pump for connection: %s", conn.NodeID)
				case <-conn.CloseChan:
					wc.logger.Debugf("[websocket/control] connection close channel triggered for: %s", conn.NodeID)
				default:
					wc.logger.Debugf("[websocket/control] websocket read error: %v, closing connection: %s", err, conn.NodeID)
				}
			}
			return
		}

		// 处理不同类型的消息
		wc.handleReceivedMessage(conn, Message{
			Type:      messageType,
			Content:   message,
			Timestamp: time.Now(),
		})
	}
}

// handleReceivedMessage 处理接收到的消息（内部方法）
//
// @description 先调用通用回调，再按帧类型分发到具体回调与业务处理器
// @param conn *Connection WebSocket连接对象
// @param message Message 消息内容
func (wc *Control) handleReceivedMessage(conn *Connection, message Message) {
	wc.logger.Debugf("[websocket/control] handling message type: %d, from connection: %s, content length: %d", message.Type, conn.NodeID, len(message.Content))

	// 首先调用通用消息回调
	wc.callbackMux.RLock()
	onMessage := wc.onMessage
	wc.callbackMux.RUnlock()
	if onMessage != nil {
		go onMessage(conn, message)
	}

	// 检查是否有注册的特定类型处理器
	if handler, exists := wc.messageHandlers.Get(message.Type); exists {
		go func(h MessageHandler) {
			if err := h.HandleMessage(conn, message); err != nil {
				wc.logger.Errorf("[websocket/control] message handler error: %v, connection: %s", err, conn.NodeID)
			}
		}(handler)
	}

	// 根据消息类型调用不同的回调函数
	switch message.Type {
	case websocket.TextMessage:
		wc.logger.Debugf("[websocket/control] received text message from connection: %s", conn.NodeID)
		wc.callbackMux.RLock()
		onText := wc.onTextMessage
		wc.callbackMux.RUnlock()
		if onText != nil {
			go onText(conn, message)
		}
		wc.handleTextMessage(conn, message)

	case websocket.BinaryMessage:
		wc.logger.Debugf("[websocket/control] received binary message from connection: %s", conn.NodeID)
		wc.callbackMux.RLock()
		onBinary := wc.onBinaryMessage
		wc.callbackMux.RUnlock()
		if onBinary != nil {
			go onBinary(conn, message)
		}
		wc.handleBinaryMessage(conn, message)

	case websocket.PingMessage:
		wc.logger.Debugf("[websocket/control] received ping message from connection: %s", conn.NodeID)
		wc.callbackMux.RLock()
		onPing := wc.onPingMessage
		wc.callbackMux.RUnlock()
		if onPing != nil {
			go onPing(conn, message)
		}
		// ping由readPump的SetPingHandler处理，这里不再重复回复

	case websocket.PongMessage:
		wc.logger.Debugf("[websocket/control] received pong message from connection: %s", conn.NodeID)
		conn.LastPongTime = time.Now()
		wc.callbackMux.RLock()
		onPong := wc.onPongMessage
		wc.callbackMux.RUnlock()
		if onPong != nil {
			go onPong(conn, message)
		}

	case websocket.CloseMessage:
		wc.logger.Debugf("[websocket/control] received close message from connection: %s", conn.NodeID)
		wc.callbackMux.RLock()
		onClose := wc.onCloseMessage
		wc.callbackMux.RUnlock()
		if onClose != nil {
			go onClose(conn, message)
		}

	default:
		wc.logger.Warnf("[websocket/control] unknown message type: %d from connection: %s", message.Type, conn.NodeID)
	}
}

// handleTextMessage 处理文本消息
//
// @description 根据BusinessType区分业务子类型（心跳/聊天/系统），与WebSocket帧类型Type解耦
// @param conn *Connection WebSocket连接对象
// @param msg Message 消息内容
func (wc *Control) handleTextMessage(conn *Connection, msg Message) {
	wc.logger.Debugf("[websocket/control] received structured message from %s: businessType=%d, from=%s, to=%s",
		conn.NodeID, msg.BusinessType, msg.From, msg.To)

	// 根据业务子类型分发
	switch msg.BusinessType {
	case BusinessTypeHeartbeat: // 心跳消息
		wc.handleHeartbeat(conn)
	case BusinessTypeChat: // 聊天消息
		wc.handleChatMessage(conn, msg)
	case BusinessTypeSystem: // 系统消息
		wc.handleSystemMessage(conn, &msg)
	default:
		wc.logger.Debugf("[websocket/control] unknown business message type: %d", msg.BusinessType)
	}
}

// handleBinaryMessage 处理二进制消息
//
// @description 二进制消息的统一入口，可在此扩展文件传输、图片传输等逻辑
// @param conn *Connection WebSocket连接对象
// @param message Message 消息内容
func (wc *Control) handleBinaryMessage(conn *Connection, message Message) {
	wc.logger.Debugf("[websocket/control] handling binary message from %s, size: %d bytes", conn.NodeID, len(message.Content))
}

// handleHeartbeat 处理心跳消息
//
// @description 刷新LastPongTime并回复心跳响应
// @param conn *Connection WebSocket连接对象
func (wc *Control) handleHeartbeat(conn *Connection) {
	conn.LastPongTime = time.Now()
	// 回复心跳（通过Send通道由writePump写入）
	heartbeatResp := Response{
		Success: true,
		Message: "heartbeat",
		Data:    time.Now().Unix(),
	}
	if data, err := json.Marshal(heartbeatResp); err == nil {
		select {
		case conn.Send <- Message{Type: websocket.TextMessage, Content: data, Timestamp: time.Now()}:
		default:
			wc.logger.Debugf("[websocket/control] send buffer full while handling heartbeat, id: %s", conn.NodeID)
		}
	}
}

// handleChatMessage 处理聊天消息
//
// @description 点对点发送（To非空且非"all"）或广播（排除发送者）
// @param conn *Connection WebSocket连接对象
// @param msg Message 消息内容
func (wc *Control) handleChatMessage(conn *Connection, msg Message) {
	wc.logger.Debugf("[websocket/control] handling chat message from %s to %s", conn.NodeID, msg.To)

	// 标注发送方
	msg.From = conn.NodeID

	// 如果指定了接收方，进行点对点发送
	if msg.To != "" && msg.To != "all" {
		targetConn := wc.GetConnection(msg.To)
		if targetConn != nil {
			if err := wc.SendMessage(targetConn.NodeID, msg); err != nil {
				// 目标连接缓冲满，通知发送方
				errorResp := Response{
					Success: false,
					Message: "target user send buffer full",
				}
				if data, err := json.Marshal(errorResp); err == nil {
					select {
					case conn.Send <- Message{Type: websocket.TextMessage, Content: data, Timestamp: time.Now()}:
					default:
					}
				}
			}
		} else {
			// 目标用户不在线，返回错误
			errorResp := Response{
				Success: false,
				Message: "target user is offline",
			}
			if data, err := json.Marshal(errorResp); err == nil {
				select {
				case conn.Send <- Message{Type: websocket.TextMessage, Content: data, Timestamp: time.Now()}:
				default:
				}
			}
		}
	} else {
		// 群聊广播，排除发送者避免回显
		wc.BroadcastExcept(conn.NodeID, msg)
	}
}

// handleSystemMessage 处理系统消息
//
// @description 系统级消息入口，可扩展用户状态更新、系统通知、配置更新等
// @param conn *Connection WebSocket连接对象
// @param msg *Message 消息内容
func (wc *Control) handleSystemMessage(conn *Connection, msg *Message) {
	wc.logger.Debugf("[websocket/control] handling system message from %s", conn.NodeID)
}

// writePump 向WebSocket连接写入消息（内部方法）
//
// @description per-connection的写入goroutine，串行消费Send通道；周期性发送ping保活；
// @description 所有对conn.Conn的WriteMessage仅在此处调用，保证单连接单writer
// @param conn *Connection WebSocket连接对象
func (wc *Control) writePump(conn *Connection) {
	defer func() {
		// 不在此关闭底层conn，由readPump统一关闭，避免双重关闭
		// 发送关闭帧后退出
		_ = conn.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait))
		_ = conn.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}()

	// 启动ping ticker（per-connection心跳）
	ticker := time.NewTicker(wc.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-wc.ctx.Done():
			// 控制器已关闭
			wc.logger.Debugf("[websocket/control] context cancelled, closing write pump for connection: %s", conn.NodeID)
			return
		case <-conn.CloseChan:
			// 连接已关闭
			wc.logger.Debugf("[websocket/control] connection close channel triggered for write pump: %s", conn.NodeID)
			return
		case message, ok := <-conn.Send:
			// 发送消息
			if !ok {
				// 发送通道已关闭（不应发生，Connection不关闭Send通道，仅关闭CloseChan）
				wc.logger.Debugf("[websocket/control] send channel closed for connection: %s", conn.NodeID)
				return
			}
			// 设置写超时
			if err := conn.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait)); err != nil {
				wc.logger.Errorf("[websocket/control] failed to set write deadline for connection %s: %v", conn.NodeID, err)
				return
			}
			// 直接写入当前连接，不再按message.To路由
			if err := conn.Conn.WriteMessage(message.Type, message.Content); err != nil {
				wc.logger.Debugf("[websocket/control] failed to write message to connection %s: %v", conn.NodeID, err)
				return
			}
			wc.logger.Debugf("[websocket/control] sent message type %d to connection: %s", message.Type, conn.NodeID)

		case <-ticker.C:
			// per-connection心跳：发送ping
			if err := conn.Conn.SetWriteDeadline(time.Now().Add(wc.writeWait)); err != nil {
				wc.logger.Errorf("[websocket/control] failed to set write deadline for ping: %v", err)
				return
			}
			if err := conn.Conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				wc.logger.Debugf("[websocket/control] ping error: %v, closing connection: %s", err, conn.NodeID)
				return
			}
			wc.logger.Debugf("[websocket/control] sent ping to connection: %s", conn.NodeID)
		}
	}
}
