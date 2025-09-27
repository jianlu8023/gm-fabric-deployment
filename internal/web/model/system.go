package model

import (
	"database/sql"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

const (
	systemInitTableName = "t_system_init_status"
)

// SystemInit 系统初始化状态模型
//
// @description 存储系统的初始化状态信息
// @struct
type SystemInit struct {
	Id     uint         `json:"uid,omitempty" yaml:"uid,omitempty" gorm:"primaryKey;check:id=1"`                          // 主键 确保id 始终是1
	IsInit sql.NullBool `json:"is_init,omitempty" yaml:"is_init,omitempty" gorm:"column:is_init;not null;default:false;"` // 是否已经初始化
}

// String 将SystemInit结构体转换为字符串
//
// @description 将SystemInit结构体转换为JSON格式的字符串
// @return string JSON格式的字符串
func (model SystemInit) String() string {
	str, _ := json.MarshalString(model)
	return str
}

// TableName 获取SystemInit结构体对应的数据库表名
//
// @description 返回SystemInit结构体在数据库中对应的表名
// @return string 数据库表名
func (SystemInit) TableName() string {
	return systemInitTableName
}
