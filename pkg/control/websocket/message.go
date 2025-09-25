package websocket

import (
	"time"
)

// MessageHandler 消息处理器接口
type MessageHandler interface {
	// HandleMessage 处理接收到的消息
	HandleMessage(conn *Connection, msg Message) error
}

// Message 消息结构体
type Message struct {
	Type      int       `json:"type"`      // 消息类型
	Content   []byte    `json:"content"`   // 消息内容
	Timestamp time.Time `json:"timestamp"` // 时间戳
	To        string    `json:"to"`        // 目标节点ID
	From      string    `json:"from"`      // 发送节点ID
}

// Response 响应结构体
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
