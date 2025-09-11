package model

import (
	"time"
)

// ConnectionInfo 表示WebSocket连接的基本信息
// 对应数据库表: websocket_connections
// 用于存储WebSocket连接的相关信息

type ConnectionInfo struct {
	ID         string    `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`
	UserID     string    `gorm:"column:user_id;type:varchar(64)" json:"user_id"`
	NodeID     string    `gorm:"column:node_id;type:varchar(64)" json:"node_id"`
	Status     string    `gorm:"column:status;type:varchar(20);default:'connected'" json:"status"` // connected, disconnected
	LastActive time.Time `gorm:"column:last_active;type:datetime" json:"last_active"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime" json:"updated_at"`
}

// TableName 指定表名
func (c *ConnectionInfo) TableName() string {
	return "websocket_connections"
}

// NewConnectionInfo 创建一个新的WebSocket连接信息
func NewConnectionInfo(id string, userID string, nodeID string) *ConnectionInfo {
	now := time.Now()
	return &ConnectionInfo{
		ID:         id,
		UserID:     userID,
		NodeID:     nodeID,
		Status:     "connected",
		LastActive: now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
