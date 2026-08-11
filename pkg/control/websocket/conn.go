package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection 表示一个WebSocket连接
//
// @description 封装底层websocket.Conn，提供发送通道、关闭信号、心跳时间戳与用户元数据
// @struct
type Connection struct {
	Conn         *websocket.Conn        // 底层WebSocket连接
	NodeID       string                 // 连接节点ID（唯一标识）
	UserID       string                 // 关联用户ID（由handler层认证后传入，便于定向推送与权限控制）
	SessionID    string                 // 关联会话ID
	RemoteAddr   string                 // 客户端地址
	UserAgent    string                 // 客户端User-Agent
	Send         chan Message           // 发送消息通道，由writePump串行消费
	CloseChan    chan struct{}          // 关闭信号通道，关闭后通知读写循环退出
	closeOnce    sync.Once              // 确保CloseChan只被关闭一次，避免重复close导致panic
	LastPongTime time.Time              // 最后一次收到pong的时间，用于心跳超时判定
	CreatedAt    time.Time              // 连接创建时间
	Metadata     map[string]interface{} // 业务扩展元数据
}

// Close 安全关闭连接的CloseChan
//
// @description 使用sync.Once保证CloseChan只被关闭一次，多次调用安全
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		close(c.CloseChan)
	})
}
