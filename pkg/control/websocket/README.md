# WebSocket 控制器功能增强

## 概述

本次更新为 WebSocket 控制器添加了完整的消息接收和处理功能，大大增强了 WebSocket 的能力。之前的控制器只能发送消息，现在可以接收和处理各种类型的 WebSocket 消息。

## 主要增强功能

### 1. 消息接收功能

- **多类型消息支持**: 支持文本消息、二进制消息、ping/pong 消息的接收和处理
- **回调机制**: 提供多种回调函数来处理不同类型的消息
- **消息处理器**: 支持注册自定义消息处理器

### 2. 新增 API 方法

#### 消息处理回调设置
```go
// 设置不同类型的消息回调
SetOnTextMessage(callback func(conn *Connection, message []byte))
SetOnBinaryMessage(callback func(conn *Connection, message []byte))
SetOnPingMessage(callback func(conn *Connection, message []byte))
SetOnPongMessage(callback func(conn *Connection, message []byte))
SetOnCloseMessage(callback func(conn *Connection, message []byte))
```

#### 消息处理器注册
```go
// 注册特定类型的消息处理器
RegisterMessageHandler(messageType int, handler MessageHandler)
```

#### 消息发送方法
```go
SendPing(id string, message []byte) bool           // 发送ping消息
SendPong(id string, message []byte) bool           // 发送pong消息
SendBinary(id string, message []byte) bool         // 发送二进制消息
SendMessage(id string, messageType int, message []byte) bool // 发送指定类型消息
SendClose(id string, reason string) bool           // 发送关闭消息
```

#### 连接管理方法
```go
GetConnectionByUserID(userID string) *Connection                    // 根据用户ID获取连接
GetConnectionsByUserID(userID string) []*Connection                 // 获取用户的所有连接
GetConnectionsByNodeID(nodeID string) []*Connection                 // 获取节点的所有连接
BroadcastByUserID(userID string, message []byte) int               // 向特定用户广播
BroadcastByNodeID(nodeID string, message []byte) int               // 向特定节点广播
SetConnectionUserID(id string, userID string) bool                 // 设置连接的用户ID
SetConnectionNodeID(id string, nodeID string) bool                 // 设置连接的节点ID
```

#### 心跳机制
```go
StartHeartbeat()  // 启动心跳检测
```

### 3. 增强的连接结构

```go
type Connection struct {
    Conn         *websocket.Conn
    ID           string
    UserID       string    // 用户ID
    NodeID       string    // 节点ID
    Send         chan []byte
    CloseChan    chan struct{}
    LastPingTime time.Time // 最后ping时间
    CreatedAt    time.Time // 创建时间
}
```

### 4. 消息处理器接口

```go
type MessageHandler interface {
    HandleMessage(conn *Connection, messageType int, message []byte) error
}
```

### 5. 结构化消息支持

```go
type Message struct {
    Type      int       `json:"type"`      // 消息类型
    Content   []byte    `json:"content"`   // 消息内容
    Timestamp time.Time `json:"timestamp"` // 时间戳
    From      string    `json:"from"`      // 发送方ID
    To        string    `json:"to"`        // 接收方ID（可选）
}
```

## 新增文件

### 1. `/pkg/control/websocket/handlers.go`
- 默认文本消息处理器 (`DefaultTextMessageHandler`)
- 默认二进制消息处理器 (`DefaultBinaryMessageHandler`)  
- 默认ping消息处理器 (`DefaultPingMessageHandler`)

### 2. `/pkg/control/websocket/example.go`
- 完整的使用示例
- 自定义消息处理器示例
- API 使用演示

## 使用示例

### 基本使用

```go
// 创建控制器
wsControl := NewWebsocketControl(config, logger)

// 设置回调
wsControl.SetOnTextMessage(func(conn *Connection, message []byte) {
    log.Printf("Received text: %s", string(message))
})

// 注册处理器
textHandler := NewDefaultTextMessageHandler(logger)
wsControl.RegisterMessageHandler(websocket.TextMessage, textHandler)

// 启动控制器
wsControl.StartUp(nil)
```

### Service 层集成

service 层已经更新，自动初始化消息处理功能：

```go
// 创建 WebSocket 服务时自动设置所有回调和处理器
wsService := NewWebSocketService(service, mapper, wsControl)
```

### Handler 层更新

handler 层的 `UpgradeHandler` 已更新，正确处理连接升级和用户/节点ID设置：

```go
// 升级连接并设置用户信息
connID, connection, err := h.service.wsControl.UpgradeConnection(ctx.Writer, ctx.Request)
h.service.wsControl.SetConnectionUserID(connID, userID.(string))
h.service.wsControl.SetConnectionNodeID(connID, req.NodeID)
```

## 心跳机制

控制器现在包含完整的心跳机制：

- 自动发送 ping 消息检测连接状态
- 配置化的 ping 间隔和 pong 等待时间
- 自动清理超时连接
- 连接状态实时监控

## 与原有 ws.go 的整合

参考了 `ws.go` 中的优秀设计：

- 消息接口设计
- 不同类型消息的处理模式
- 客户端管理机制
- 心跳检测逻辑

## 配置选项

控制器支持以下配置：

```go
pingInterval    time.Duration // ping间隔 (默认: 54秒)
pongWait        time.Duration // pong等待时间 (默认: 60秒)
writeWait       time.Duration // 写入等待时间 (默认: 10秒)
maxMessageSize  int64         // 最大消息大小 (默认: 512字节)
```

## 向后兼容性

所有原有的API保持不变，新功能作为扩展添加，确保现有代码无需修改即可继续工作。

## 测试和验证

建议进行以下测试：

1. 基本消息收发测试
2. 不同类型消息处理测试
3. 心跳机制测试
4. 连接管理功能测试
5. 广播功能测试
6. 错误处理测试

## 未来改进建议

1. 添加消息持久化功能
2. 实现消息队列支持
3. 添加连接池管理
4. 实现集群支持
5. 添加监控和统计功能

---

通过这次增强，WebSocket 控制器现在具备了完整的双向通信能力，可以满足各种实时通信需求。