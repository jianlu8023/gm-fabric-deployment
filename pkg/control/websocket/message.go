package websocket

import (
	"time"
)

// 业务消息子类型常量（BusinessType字段使用）
const (
	// BusinessTypeHeartbeat 心跳消息
	BusinessTypeHeartbeat = 1
	// BusinessTypeChat 聊天消息
	BusinessTypeChat = 2
	// BusinessTypeSystem 系统消息
	BusinessTypeSystem = 3
)

// MessageHandler 消息处理器接口
//
// @description 定义处理WebSocket消息的接口，按消息类型注册到控制器
// @interface
type MessageHandler interface {
	// HandleMessage 处理接收到的消息
	//
	// @description 业务侧实现该接口处理特定类型消息
	// @param conn *Connection 消息来源连接
	// @param msg Message 消息内容
	// @return error 处理过程中的错误
	HandleMessage(conn *Connection, msg Message) error
}

// Message 消息结构体
//
// @description 统一的WebSocket消息载体，Type为WebSocket帧类型，BusinessType为业务子类型
// @struct
type Message struct {
	Type         int       `json:"type"`                    // WebSocket帧类型（TextMessage/BinaryMessage/PingMessage/PongMessage/CloseMessage）
	BusinessType int       `json:"business_type,omitempty"` // 业务子类型（1=心跳 2=聊天 3=系统），仅文本/二进制消息使用
	Content      []byte    `json:"content"`                 // 消息内容
	Timestamp    time.Time `json:"timestamp,omitempty"`     // 消息时间戳，发送时由控制器填充
	To           string    `json:"to,omitempty"`            // 目标节点ID，为空或"all"表示广播
	From         string    `json:"from,omitempty"`          // 发送节点ID
}

// Response 响应结构体
//
// @description 控制器向客户端返回的统一响应结构
// @struct
type Response struct {
	Success bool        `json:"success"`        // 是否成功
	Message string      `json:"message"`        // 响应消息
	Data    interface{} `json:"data,omitempty"` // 响应数据
}
