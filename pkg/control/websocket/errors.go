package websocket

import "errors"

// 定义WebSocket相关错误
//
// @description 集中定义WebSocket控制器对外暴露的错误，供调用方做错误判定与分类处理
var (
	// ErrConnectionNotFound 表示找不到指定的连接
	ErrConnectionNotFound = errors.New("websocket: connection not found")

	// ErrSendBufferFull 表示发送缓冲区已满
	ErrSendBufferFull = errors.New("websocket: send buffer full")

	// ErrConnectionClosed 表示连接已关闭
	ErrConnectionClosed = errors.New("websocket: connection closed")

	// ErrInvalidMessageType 表示消息类型无效
	ErrInvalidMessageType = errors.New("websocket: invalid message type")

	// ErrReadMessageFailed 表示读取消息失败
	ErrReadMessageFailed = errors.New("websocket: read message failed")

	// ErrWriteMessageFailed 表示写入消息失败
	ErrWriteMessageFailed = errors.New("websocket: write message failed")

	// ErrControlClosed 表示控制器已关闭
	ErrControlClosed = errors.New("websocket: control closed")

	// ErrInvalidConnectionID 表示连接ID无效
	ErrInvalidConnectionID = errors.New("websocket: invalid connection ID")

	// ErrConnectionAlreadyExists 表示连接已存在
	ErrConnectionAlreadyExists = errors.New("websocket: connection already exists")

	// ErrHandlerNotSet 表示处理器未设置
	ErrHandlerNotSet = errors.New("websocket: handler not set")
)
