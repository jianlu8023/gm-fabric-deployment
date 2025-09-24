package websocket

import (
	"encoding/json"
	"fmt"
	"time"
)

// DefaultTextMessageHandler 默认文本消息处理器
type DefaultTextMessageHandler struct {
	logger interface {
		Debugf(template string, args ...interface{})
		Infof(template string, args ...interface{})
		Warnf(template string, args ...interface{})
		Errorf(template string, args ...interface{})
	}
}

// NewDefaultTextMessageHandler 创建默认文本消息处理器
func NewDefaultTextMessageHandler(logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}) *DefaultTextMessageHandler {
	return &DefaultTextMessageHandler{
		logger: logger,
	}
}

// HandleMessage 处理文本消息
func (h *DefaultTextMessageHandler) HandleMessage(conn *Connection, messageType int, message []byte) error {
	h.logger.Debugf("[handler] processing text message from connection: %s", conn.ID)

	// 尝试解析JSON消息
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		// 如果不是JSON格式，作为普通文本处理
		h.logger.Debugf("[handler] received plain text: %s", string(message))

		// 可以在这里添加自定义的文本处理逻辑
		response := Response{
			Success: true,
			Message: "text message received",
			Data:    string(message),
		}

		if responseData, err := json.Marshal(response); err == nil {
			// 这里需要通过connection.Send发送响应
			select {
			case conn.Send <- responseData:
				h.logger.Debugf("[handler] sent response to connection: %s", conn.ID)
			default:
				h.logger.Warnf("[handler] failed to send response, channel full for connection: %s", conn.ID)
			}
		}
		return nil
	}

	// 处理结构化消息
	h.logger.Debugf("[handler] received structured message: type=%d, from=%s", msg.Type, msg.From)
	return h.handleStructuredMessage(conn, &msg)
}

// handleStructuredMessage 处理结构化消息
func (h *DefaultTextMessageHandler) handleStructuredMessage(conn *Connection, msg *Message) error {
	switch msg.Type {
	case 1: // 心跳消息
		return h.handleHeartbeat(conn, msg)
	case 2: // 聊天消息
		return h.handleChat(conn, msg)
	case 3: // 系统消息
		return h.handleSystem(conn, msg)
	default:
		h.logger.Warnf("[handler] unknown message type: %d", msg.Type)
		return fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

// handleHeartbeat 处理心跳消息
func (h *DefaultTextMessageHandler) handleHeartbeat(conn *Connection, msg *Message) error {
	h.logger.Debugf("[handler] handling heartbeat from connection: %s", conn.ID)

	response := Response{
		Success: true,
		Message: "heartbeat",
		Data:    time.Now().Unix(),
	}

	if responseData, err := json.Marshal(response); err == nil {
		select {
		case conn.Send <- responseData:
			h.logger.Debugf("[handler] sent heartbeat response to connection: %s", conn.ID)
		default:
			h.logger.Warnf("[handler] failed to send heartbeat response, channel full for connection: %s", conn.ID)
		}
	}
	return nil
}

// handleChat 处理聊天消息
func (h *DefaultTextMessageHandler) handleChat(conn *Connection, msg *Message) error {
	h.logger.Debugf("[handler] handling chat message from %s to %s", msg.From, msg.To)

	// 这里可以添加聊天消息的处理逻辑
	// 例如：消息过滤、存储、转发等

	response := Response{
		Success: true,
		Message: "chat message received",
		Data: map[string]interface{}{
			"messageId": fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			"timestamp": time.Now().Unix(),
		},
	}

	if responseData, err := json.Marshal(response); err == nil {
		select {
		case conn.Send <- responseData:
			h.logger.Debugf("[handler] sent chat response to connection: %s", conn.ID)
		default:
			h.logger.Warnf("[handler] failed to send chat response, channel full for connection: %s", conn.ID)
		}
	}
	return nil
}

// handleSystem 处理系统消息
func (h *DefaultTextMessageHandler) handleSystem(conn *Connection, msg *Message) error {
	h.logger.Debugf("[handler] handling system message from connection: %s", conn.ID)

	// 这里可以添加系统消息的处理逻辑
	// 例如：配置更新、状态同步等

	response := Response{
		Success: true,
		Message: "system message processed",
		Data:    time.Now().Unix(),
	}

	if responseData, err := json.Marshal(response); err == nil {
		select {
		case conn.Send <- responseData:
			h.logger.Debugf("[handler] sent system response to connection: %s", conn.ID)
		default:
			h.logger.Warnf("[handler] failed to send system response, channel full for connection: %s", conn.ID)
		}
	}
	return nil
}

// DefaultBinaryMessageHandler 默认二进制消息处理器
type DefaultBinaryMessageHandler struct {
	logger interface {
		Debugf(template string, args ...interface{})
		Infof(template string, args ...interface{})
		Warnf(template string, args ...interface{})
		Errorf(template string, args ...interface{})
	}
}

// NewDefaultBinaryMessageHandler 创建默认二进制消息处理器
func NewDefaultBinaryMessageHandler(logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}) *DefaultBinaryMessageHandler {
	return &DefaultBinaryMessageHandler{
		logger: logger,
	}
}

// HandleMessage 处理二进制消息
func (h *DefaultBinaryMessageHandler) HandleMessage(conn *Connection, messageType int, message []byte) error {
	h.logger.Debugf("[handler] processing binary message from connection: %s, size: %d bytes", conn.ID, len(message))

	// 这里可以添加二进制数据的处理逻辑
	// 例如：文件上传、图片处理、音频/视频数据等

	// 简单的示例：回显二进制数据的长度
	response := Response{
		Success: true,
		Message: "binary data received",
		Data: map[string]interface{}{
			"size":      len(message),
			"timestamp": time.Now().Unix(),
		},
	}

	if responseData, err := json.Marshal(response); err == nil {
		select {
		case conn.Send <- responseData:
			h.logger.Debugf("[handler] sent binary response to connection: %s", conn.ID)
		default:
			h.logger.Warnf("[handler] failed to send binary response, channel full for connection: %s", conn.ID)
		}
	}

	return nil
}

// DefaultPingMessageHandler 默认ping消息处理器
type DefaultPingMessageHandler struct {
	logger interface {
		Debugf(template string, args ...interface{})
	}
}

// NewDefaultPingMessageHandler 创建默认ping消息处理器
func NewDefaultPingMessageHandler(logger interface {
	Debugf(template string, args ...interface{})
}) *DefaultPingMessageHandler {
	return &DefaultPingMessageHandler{
		logger: logger,
	}
}

// HandleMessage 处理ping消息
func (h *DefaultPingMessageHandler) HandleMessage(conn *Connection, messageType int, message []byte) error {
	h.logger.Debugf("[handler] processing ping message from connection: %s", conn.ID)

	// ping消息通常由WebSocket框架自动处理
	// 这里只是记录日志，实际的pong响应由control.go中的ping handler处理

	return nil
}
