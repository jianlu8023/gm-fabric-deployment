package node

import (
	"database/sql"
	"errors"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"gorm.io/gorm"
	"time"
)

const (
	nodeInfoTableName = "t_node_info"
)

// Info 节点信息 指的是libp2p节点信息 以libp2p节点当作机器
type Info struct {
	AutoUid              int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                 // 自增ID
	NodeId               string       `json:"node_id,omitempty" yaml:"node_id,omitempty" gorm:"column:node_id;type:varchar(32);unique;"`                                       // 节点ID
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

type Mapper struct {
	dbConn *gorm.DB
}

func NewNodeMapper(conn *gorm.DB) *Mapper {
	return &Mapper{dbConn: conn}
}

func (m *Mapper) InsertOneWithCheck(node *Info) error {
	if m.dbConn == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.dbConn.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Info{}).Where(&Info{
			NodeId: node.NodeId,
		}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			// 节点已存在
			return datasource.ErrAlreadyExists
		}
		// 节点不存在，插入
		if err := tx.Model(&Info{}).Create(node).Error; err != nil {
			return err
		}
		return nil
	})
}

func (m *Mapper) InsertOrUpdate(node *Info) error {
	if m.dbConn == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.dbConn.Transaction(func(tx *gorm.DB) error {
		var existInfo Info
		if err := tx.Model(&Info{}).Where(&Info{
			NodeId: node.NodeId,
		}).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在
				if err := tx.Model(&Info{}).Create(node).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新
			node.AutoUid = existInfo.AutoUid
			if err := tx.Model(&Info{}).Where(&Info{AutoUid: node.AutoUid}).Updates(node).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
