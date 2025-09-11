package node

import (
	"database/sql"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"time"
)

const (
	nodeInfoTableName = "t_node_info"
)

// Info 节点信息 指的是libp2p节点信息 以libp2p节点当作机器
type Info struct {
	AutoUid              int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                 // 自增ID
	NodeId               string       `json:"node_id,omitempty" yaml:"node_id,omitempty" gorm:"column:node_id;type:varchar(256);unique;"`                                      // 节点ID
	LastAliveMessageTime time.Time    `json:"last_alive_message_time,omitempty" yaml:"last_alive_message_time,omitempty" gorm:"column:last_alive_message_time;type:datetime;"` // 最后一次收到心跳时间
	IsAlive              sql.NullBool `json:"is_alive,omitempty" yaml:"is_alive,omitempty" gorm:"column:is_alive;type:bool;"`                                                  // 是否存活
	IsMySelf             sql.NullBool `json:"is_my_self,omitempty" yaml:"is_my_self,omitempty" gorm:"column:is_my_self;type:bool;default:false"`                               // 是否是本机
}

// TableName 返回表名
// @return string 表名
func (i Info) TableName() string {
	return nodeInfoTableName
}

// String 返回json格式字符串
// @return string json格式字符串
func (i Info) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

// NewNodeInfo 新建节点信息
// @return *Info 节点信息
func NewNodeInfo() *Info {
	return &Info{}
}
