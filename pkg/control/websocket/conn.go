package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

// Connection 表示一个WebSocket连接
type Connection struct {
	Conn         *websocket.Conn
	NodeID       string // 节点ID
	Send         chan Message
	CloseChan    chan struct{}
	LastPingTime time.Time // 最后ping时间
	CreatedAt    time.Time // 创建时间
}
