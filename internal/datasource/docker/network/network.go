package network

import (
	"database/sql"
	"errors"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"gorm.io/gorm"
	"time"
)

const (
	networkInfoTableName = "t_docker_network_info"
)

type Info struct {
	AutoUid               int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                        // 自增id
	NetworkName           string       `json:"network_name,omitempty" yaml:"network_name,omitempty" gorm:"column:network_name;type:varchar(255);not null;"`                            // network名称
	NetworkID             string       `json:"network_id,omitempty" yaml:"network_id,omitempty" gorm:"column:network_id;type:varchar(255);not null;"`                                  // network id
	NetworkCreateTime     time.Time    `json:"network_create_time,omitempty" yaml:"network_create_time,omitempty" gorm:"column:network_create_time;type:datetime;not null;"`           // network创建时间
	NetworkScope          string       `json:"network_scope,omitempty" yaml:"network_scope,omitempty" gorm:"column:network_scope;type:varchar(255);not null;"`                         // network作用域
	NetworkDriver         string       `json:"network_driver,omitempty" yaml:"network_driver,omitempty" gorm:"column:network_driver;type:varchar(255);not null;"`                      // network驱动
	NetworkEnableIPv6     sql.NullBool `json:"network_enable_ipv6,omitempty" yaml:"network_enable_ipv6,omitempty" gorm:"column:network_enable_ipv6;type:bool;not null;"`               // network是否启用ipv6
	NetworkIpam           string       `json:"network_ipam,omitempty" yaml:"network_ipam,omitempty" gorm:"column:network_ipam;type:text;not null;"`                                    // network ipam
	NetworkInternal       sql.NullBool `json:"network_internal,omitempty" yaml:"network_internal,omitempty" gorm:"column:network_internal;type:bool;not null;"`                        // network是否内部
	NetworkAttachable     sql.NullBool `json:"network_attachable,omitempty" yaml:"network_attachable,omitempty" gorm:"column:network_attachable;type:bool;not null;"`                  // network是否可附加
	NetworkIngress        sql.NullBool `json:"network_ingress,omitempty" yaml:"network_ingress,omitempty" gorm:"column:network_ingress;type:bool;not null;"`                           // network是否为入口
	NetworkLocationPeerId string       `json:"network_location_peer_id,omitempty" yaml:"network_location_peer_id,omitempty" gorm:"column:network_location_peer_id;type:varchar(255);"` // network位置peer id
	IsDelete              sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:bool;default:false"`                                         // 是否删除
}

func (i Info) TableName() string {
	return networkInfoTableName
}

func (i Info) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

func NewNetworkInfo() *Info {
	return &Info{}
}

type Mapper struct {
	dbConn *gorm.DB
}

func NewNetworkMapper(conn *gorm.DB) *Mapper {
	return &Mapper{dbConn: conn}
}

func (m *Mapper) InsertOrUpdateOne(info *Info) error {
	if m.dbConn == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.dbConn.Transaction(func(tx *gorm.DB) error {
		var existInfo Info
		if err := tx.Model(&Info{}).Where(&Info{
			NetworkID:             info.NetworkID,
			NetworkLocationPeerId: info.NetworkLocationPeerId,
			IsDelete:              sql.NullBool{Bool: false, Valid: true},
		}).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在
				if err := tx.Model(&Info{}).Create(info).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新
			info.AutoUid = existInfo.AutoUid
			if err := tx.Model(&Info{}).Where(&Info{AutoUid: info.AutoUid}).Updates(info).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
