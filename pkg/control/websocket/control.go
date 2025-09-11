package websocket

import (
	"context"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/random/uuid"
	"go.uber.org/zap"
)

// Connection 表示一个WebSocket连接
type Connection struct {
	Conn      *websocket.Conn
	ID        string
	Send      chan []byte
	CloseChan chan struct{}
}

// Control 管理WebSocket连接的控制结构体
type Control struct {
	logger         *zap.SugaredLogger
	config         *config.HttpServerConfig
	connections    map[string]*Connection
	connectionsMux sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	once           sync.Once
	onConnected    func(conn *Connection)
	onDisconnected func(conn *Connection)
	onMessage      func(conn *Connection, message []byte)
	upgrader       websocket.Upgrader
}

// NewWebsocketControl 创建一个新的WebSocket控制器
// @param serverConfig *config.HttpServerConfig HTTP服务配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control WebSocket控制器
func NewWebsocketControl(serverConfig *config.HttpServerConfig, loggerControl *logger.Control) *Control {
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
	}

	// 初始化WebsocketControl
	wc := &Control{
		logger:      wsLogger,
		config:      serverConfig,
		connections: make(map[string]*Connection),
		ctx:         ctx,
		cancel:      cancel,
		upgrader:    upgrader,
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
func (wc *Control) SetOnConnected(callback func(conn *Connection)) {
	wc.onConnected = callback
}

// SetOnDisconnected 设置连接关闭时的回调函数
func (wc *Control) SetOnDisconnected(callback func(conn *Connection)) {
	wc.onDisconnected = callback
}

// SetOnMessage 设置收到消息时的回调函数
func (wc *Control) SetOnMessage(callback func(conn *Connection, message []byte)) {
	wc.onMessage = callback
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
		Conn:      conn,
		ID:        id,
		Send:      make(chan []byte, 100),
		CloseChan: make(chan struct{}),
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

// GetConnectionCount 获取当前连接数
// @return int 连接数
func (wc *Control) GetConnectionCount() int {
	wc.connectionsMux.RLock()
	defer wc.connectionsMux.RUnlock()
	return len(wc.connections)
}

// readPump 从WebSocket连接读取消息
func (wc *Control) readPump(conn *Connection) {
	defer func() {
		wc.RemoveConnection(conn.ID)
		conn.Conn.Close()
	}()

	// 设置读取限制
	conn.Conn.SetReadLimit(512)

	for {
		select {
		case <-wc.ctx.Done():
			// 控制器已关闭
			return
		case <-conn.CloseChan:
			// 连接已关闭
			return
		default:
			// 读取消息
			_, message, err := conn.Conn.ReadMessage()
			if err != nil {
				// 读取错误，关闭连接
				wc.logger.Debugf("[control] websocket read error: %v, closing connection: %s", err, conn.ID)
				return
			}

			// 调用消息回调
			if wc.onMessage != nil {
				go wc.onMessage(conn, message)
			}
		}
	}
}

// writePump 向WebSocket连接写入消息
func (wc *Control) writePump(conn *Connection) {
	defer func() {
		conn.Conn.Close()
	}()

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

			// 写入文本消息
			err := conn.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				// 写入错误，关闭连接
				wc.logger.Debugf("[control] websocket write error: %v, closing connection: %s", err, conn.ID)
				return
			}
		}
	}
}

// StartUp 启动WebSocket控制器
func (wc *Control) StartUp(failedFunc func(err error)) {
	wc.once.Do(func() {
		wc.logger.Infof("[control] websocket control started...")
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
