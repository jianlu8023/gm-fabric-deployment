// Package model 提供Web模块的数据模型定义
//
// @Description 包括用户、节点、Docker镜像、Docker网络、文件、WebSocket连接等业务模型的定义，
// 所有模型均基于基础Model结构体，提供统一的主键和时间字段管理。
package model

import (
	"time"

	"github.com/jianlu8023/golang-example/pkg/snowflake"
	"gorm.io/gorm"
)

// Model 基础模型结构体
//
// @description 所有业务模型的基础结构体，封装统一的行ID、创建时间和更新时间字段
// @struct
type Model struct {
	RawUid         int64     `json:"raw_uid" yaml:"raw_uid" gorm:"column:raw_uid;type:int;comment:'行id';"`                                                        // 行id（雪花算法生成）
	CreateDateTime time.Time `json:"create_date_time" yaml:"create_date_time" gorm:"column:create_date_time;type:datetime;comment:'记录创建时间';autoCreateTime:true;"` // 记录创建时间（自动填充）
	UpdateDateTime time.Time `json:"update_date_time" yaml:"update_date_time" gorm:"column:update_date_time;type:datetime;comment:'记录更新时间';autoUpdateTime:true;"` // 记录更新时间（自动填充）
}

// BeforeSave 保存前的钩子函数
//
// @description 在记录保存前调用，当RawUid为0时通过雪花算法自动生成行ID
// @param db *gorm.DB GORM数据库操作实例
// @return error 雪花ID生成失败时返回错误，成功返回nil
func (m *Model) BeforeSave(db *gorm.DB) error {
	if m.RawUid == 0 {
		id, err := snowflake.GenID()
		if err != nil {
			return err
		}
		m.RawUid = int64(id)
	}

	// if m.CreateDateTime.IsZero() {
	// 	m.CreateDateTime = time.Now()
	// }
	// if m.UpdateDateTime.IsZero() {
	// 	m.UpdateDateTime = time.Now()
	// }
	return nil
}

// func (m *Model) BeforeUpdate(db *gorm.DB) error {
// 	m.UpdateDateTime = time.Now()
// 	return nil
// }

// BeforeCreate 创建前的钩子函数
//
// @description 在记录插入数据库前调用，当RawUid为0时通过雪花算法自动生成行ID
// @param db *gorm.DB GORM数据库操作实例
// @return error 雪花ID生成失败时返回错误，成功返回nil
func (m *Model) BeforeCreate(db *gorm.DB) error {
	if m.RawUid == 0 {
		id, err := snowflake.GenID()
		if err != nil {
			return err
		}
		m.RawUid = int64(id)
	}
	// if m.CreateDateTime.IsZero() {
	// 	m.CreateDateTime = time.Now()
	// }
	// if m.UpdateDateTime.IsZero() {
	// 	m.UpdateDateTime = time.Now()
	// }
	return nil
}

// func (m *Model) MarshalJSON() ([]byte, error) {
// 	type Alias Model
// 	aux := struct {
// 		*Alias
// 	}{
// 		Alias: (*Alias)(m),
// 	}
// 	return json.Marshal(aux)
// }
//
// func (m *Model) GoString() string {
// 	return m.String()
// }
//
// func (m *Model) String() string {
// 	marshalString, _ := json.MarshalString(m)
// 	return marshalString
// }
