package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

// Example 展示如何使用增强的WebSocket控制器
func Example() {
	// 创建配置
	serverConfig := &config.HttpServerConfig{
		Enabled: true,
	}

	// 创建日志控制器（这里需要根据实际项目进行调整）
	loggerControl := &logger.Control{}

	// 创建WebSocket控制器
	wsControl := NewWebsocketControl(serverConfig, loggerControl)
	if wsControl == nil {
		fmt.Println("Failed to create websocket control")
		return
	}

	// 设置连接建立回调
	wsControl.SetOnConnected(func(conn *Connection) {
		fmt.Printf("New connection established: %s\n", conn.ID)

		// 可以在这里设置连接的用户ID和节点ID
		// wsControl.SetConnectionUserID(conn.ID, "user123")
		// wsControl.SetConnectionNodeID(conn.ID, "node456")

		// 发送欢迎消息
		welcomeMsg := Response{
			Success: true,
			Message: "Welcome to WebSocket server",
			Data: map[string]interface{}{
				"connectionId": conn.ID,
				"timestamp":    conn.CreatedAt.Unix(),
			},
		}

		if data, err := json.Marshal(welcomeMsg); err == nil {
			wsControl.SendTo(conn.ID, data)
		}
	})

	// 设置连接断开回调
	wsControl.SetOnDisconnected(func(conn *Connection) {
		fmt.Printf("Connection disconnected: %s\n", conn.ID)
	})

	// 设置通用消息回调
	wsControl.SetOnMessage(func(conn *Connection, message []byte) {
		fmt.Printf("Received message from %s: %s\n", conn.ID, string(message))
	})

	// 设置文本消息回调
	wsControl.SetOnTextMessage(func(conn *Connection, message []byte) {
		fmt.Printf("Received text message from %s: %s\n", conn.ID, string(message))
	})

	// 设置二进制消息回调
	wsControl.SetOnBinaryMessage(func(conn *Connection, message []byte) {
		fmt.Printf("Received binary message from %s: %d bytes\n", conn.ID, len(message))
	})

	// 设置ping消息回调
	wsControl.SetOnPingMessage(func(conn *Connection, message []byte) {
		fmt.Printf("Received ping from %s\n", conn.ID)
	})

	// 设置pong消息回调
	wsControl.SetOnPongMessage(func(conn *Connection, message []byte) {
		fmt.Printf("Received pong from %s\n", conn.ID)
	})

	// 注册消息处理器
	// 创建并注册文本消息处理器
	textHandler := NewDefaultTextMessageHandler(wsControl.logger)
	wsControl.RegisterMessageHandler(websocket.TextMessage, textHandler)

	// 创建并注册二进制消息处理器
	binaryHandler := NewDefaultBinaryMessageHandler(wsControl.logger)
	wsControl.RegisterMessageHandler(websocket.BinaryMessage, binaryHandler)

	// 创建并注册ping消息处理器
	pingHandler := NewDefaultPingMessageHandler(wsControl.logger)
	wsControl.RegisterMessageHandler(websocket.PingMessage, pingHandler)

	// 启动控制器
	wsControl.StartUp(func(err error) {
		fmt.Printf("WebSocket control startup failed: %v\n", err)
	})

	fmt.Println("WebSocket control started successfully")
	fmt.Printf("Current connections: %d\n", wsControl.GetConnectionCount())

	// 演示一些API的使用
	demonstrateUsage(wsControl)

	// 关闭控制器
	// wsControl.Shutdown()
}

// demonstrateUsage 演示WebSocket控制器的各种功能
func demonstrateUsage(wsControl *Control) {
	fmt.Println("\n=== WebSocket Control API Usage Examples ===")

	// 1. 获取连接数
	fmt.Printf("Total connections: %d\n", wsControl.GetConnectionCount())

	// 2. 广播消息
	broadcastMsg := Response{
		Success: true,
		Message: "Broadcast message to all connections",
		Data:    "Hello everyone!",
	}
	if data, err := json.Marshal(broadcastMsg); err == nil {
		wsControl.Broadcast(data)
		fmt.Println("Broadcasted message to all connections")
	}

	// 3. 向特定用户广播
	userMsg := Response{
		Success: true,
		Message: "Message for specific user",
		Data:    "Hello user123!",
	}
	if data, err := json.Marshal(userMsg); err == nil {
		count := wsControl.BroadcastByUserID("user123", data)
		fmt.Printf("Broadcasted message to %d connections of user123\n", count)
	}

	// 4. 向特定节点广播
	nodeMsg := Response{
		Success: true,
		Message: "Message for specific node",
		Data:    "Hello node456!",
	}
	if data, err := json.Marshal(nodeMsg); err == nil {
		count := wsControl.BroadcastByNodeID("node456", data)
		fmt.Printf("Broadcasted message to %d connections of node456\n", count)
	}

	// 5. 发送ping到特定连接
	// wsControl.SendPing("connection_id", []byte("custom_ping"))

	// 6. 发送不同类型的消息
	// testMsg := []byte("test message")
	// wsControl.SendMessage("connection_id", websocket.TextMessage, testMsg)
	// wsControl.SendMessage("connection_id", websocket.BinaryMessage, testMsg)

	fmt.Println("API usage demonstration completed")
}

// CustomMessageHandler 自定义消息处理器示例
type CustomMessageHandler struct {
	logger interface {
		Debugf(template string, args ...interface{})
		Infof(template string, args ...interface{})
	}
}

// NewCustomMessageHandler 创建自定义消息处理器
func NewCustomMessageHandler(logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
}) *CustomMessageHandler {
	return &CustomMessageHandler{
		logger: logger,
	}
}

// HandleMessage 处理自定义消息
func (h *CustomMessageHandler) HandleMessage(conn *Connection, messageType int, message []byte) error {
	h.logger.Infof("[custom] handling message type %d from connection %s", messageType, conn.ID)

	// 在这里添加你的自定义消息处理逻辑
	switch messageType {
	case websocket.TextMessage:
		// 处理文本消息
		h.logger.Debugf("[custom] processing text: %s", string(message))

	case websocket.BinaryMessage:
		// 处理二进制消息
		h.logger.Debugf("[custom] processing binary data: %d bytes", len(message))

	default:
		h.logger.Debugf("[custom] unsupported message type: %d", messageType)
	}

	return nil
}

// GetUsageInstructions 获取使用说明
func GetUsageInstructions() string {
	return `
WebSocket 控制器增强功能使用说明：

1. 消息接收功能：
   - 支持文本消息、二进制消息、ping/pong消息的接收和处理
   - 提供了多种回调函数来处理不同类型的消息
   - 支持自定义消息处理器

2. 新增的API方法：
   - SetOnTextMessage()：设置文本消息回调
   - SetOnBinaryMessage()：设置二进制消息回调
   - SetOnPingMessage()：设置ping消息回调
   - SetOnPongMessage()：设置pong消息回调
   - RegisterMessageHandler()：注册消息处理器
   - SendPing()：发送ping消息
   - SendPong()：发送pong消息
   - SendBinary()：发送二进制消息
   - SendMessage()：发送指定类型的消息
   - GetConnectionByUserID()：根据用户ID获取连接
   - GetConnectionsByUserID()：获取用户的所有连接
   - GetConnectionsByNodeID()：获取节点的所有连接
   - BroadcastByUserID()：向特定用户广播
   - BroadcastByNodeID()：向特定节点广播
   - SetConnectionUserID()：设置连接的用户ID
   - SetConnectionNodeID()：设置连接的节点ID
   - StartHeartbeat()：启动心跳检测

3. 心跳机制：
   - 自动发送ping消息检测连接状态
   - 支持配置ping间隔和pong等待时间
   - 自动清理超时连接

4. 消息处理器：
   - 支持注册不同类型的消息处理器
   - 提供了默认的消息处理器实现
   - 可以实现MessageHandler接口创建自定义处理器

5. 使用示例：
   // 创建控制器
   wsControl := NewWebsocketControl(config, logger)
   
   // 设置回调
   wsControl.SetOnTextMessage(func(conn *Connection, message []byte) {
       // 处理文本消息
   })
   
   // 注册处理器
   handler := NewDefaultTextMessageHandler(logger)
   wsControl.RegisterMessageHandler(websocket.TextMessage, handler)
   
   // 启动控制器
   wsControl.StartUp(nil)
`
}
