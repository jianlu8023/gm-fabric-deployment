package model

import (
	"time"
)

// ConnectionInfo WebSocket连接信息模型
// @description 表示WebSocket连接的基本信息，用于存储和管理WebSocket连接状态
// @struct
type ConnectionInfo struct {
	ID         string    `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`                  // 连接唯一ID（主键）
	UserID     string    `gorm:"column:user_id;type:varchar(64)" json:"user_id"`                   // 用户ID，关联到用户表
	NodeID     string    `gorm:"column:node_id;type:varchar(64)" json:"node_id"`                   // 节点ID，关联到节点表
	Status     string    `gorm:"column:status;type:varchar(20);default:'connected'" json:"status"` // 连接状态（connected:已连接, disconnected:已断开）
	LastActive time.Time `gorm:"column:last_active;type:datetime" json:"last_active"`              // 最后活动时间
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime" json:"created_at"`                // 创建时间
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime" json:"updated_at"`                // 更新时间
}

// TableName 返回数据库表名
// @description 实现gorm接口，指定ConnectionInfo结构体对应的数据库表名
// @return string 数据库表名
func (ConnectionInfo) TableName() string {
	return "websocket_connections"
}

// NewConnectionInfo 创建一个新的WebSocket连接信息实例
// @description 初始化WebSocket连接信息，设置初始状态为connected
// @param id string 连接唯一标识符
// @param userID string 用户ID
// @param nodeID string 节点ID
// @return *ConnectionInfo WebSocket连接信息结构体指针
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
