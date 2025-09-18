package model

import (
	"database/sql"
	humantime "github.com/jianlu8023/golang-example/pkg/human/time"
	"github.com/jianlu8023/golang-example/pkg/json"
	"time"
)

const (
	nodeInfoTableName = "t_node_info"
)

type Libp2pNode struct {
	AutoUid              int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                 // 自增ID
	NodeId               string       `json:"node_id,omitempty" yaml:"node_id,omitempty" gorm:"column:node_id;type:varchar(256);unique;"`                                      // 节点ID（唯一标识）
	NodeIp               string       `json:"node_ip,omitempty" yaml:"node_ip,omitempty" gorm:"column:node_ip;type:varchar(256);"`                                             // 节点ip
	LastAliveMessageTime time.Time    `json:"last_alive_message_time,omitempty" yaml:"last_alive_message_time,omitempty" gorm:"column:last_alive_message_time;type:datetime;"` // 最后一次收到心跳时间
	IsAlive              sql.NullBool `json:"is_alive,omitempty" yaml:"is_alive,omitempty" gorm:"column:is_alive;type:bool;"`                                                  // 是否存活状态
	IsMySelf             sql.NullBool `json:"is_my_self,omitempty" yaml:"is_my_self,omitempty" gorm:"column:is_my_self;type:bool;default:false"`                               // 是否是本机节点
}

func (i *Libp2pNode) MarshalJSON() ([]byte, error) {
	type Alias Libp2pNode

	aux := struct {
		*Alias
		IsAlive              bool   `json:"is_alive,omitempty" yaml:"is_alive,omitempty"`
		IsMySelf             bool   `json:"is_my_self,omitempty" yaml:"is_my_self,omitempty"`
		LastAliveMessageTime string `yaml:"last_alive_message_time,omitempty" yaml:"last_alive_message_time,omitempty"`
	}{
		Alias:                (*Alias)(i),
		IsAlive:              i.IsAlive.Bool,
		IsMySelf:             i.IsMySelf.Bool,
		LastAliveMessageTime: humantime.HumanTimeLower(i.LastAliveMessageTime, "unknown"),
	}
	return json.Marshal(aux)
}

// TableName 返回表名
// @description 实现gorm接口，指定结构体对应的数据库表名
// @return string 数据库表名
func (i *Libp2pNode) TableName() string {
	return nodeInfoTableName
}

// String 返回json格式字符串
// @description 将节点信息转换为JSON格式字符串
// @return string 节点信息的JSON格式字符串
func (i *Libp2pNode) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

// NewLibp2pNode 创建新的节点信息实例
// @description 初始化一个空的节点信息结构体指针
// @return *Info 节点信息结构体指针
func NewLibp2pNode() *Libp2pNode {
	return &Libp2pNode{}
}
