package network

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"gorm.io/gorm"
)

const (
	networkInfoTableName = "t_docker_network_info"
)

// Info Docker网络信息模型
// @description 定义Docker网络的数据结构，存储网络的详细配置和状态信息
// @struct
type Info struct {
	AutoUid               int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                        // 自增ID（主键）
	NetworkName           string       `json:"network_name,omitempty" yaml:"network_name,omitempty" gorm:"column:network_name;type:varchar(255);not null;"`                            // 网络名称
	NetworkID             string       `json:"network_id,omitempty" yaml:"network_id,omitempty" gorm:"column:network_id;type:varchar(255);not null;"`                                  // 网络ID（Docker生成的唯一标识）
	NetworkCreateTime     time.Time    `json:"network_create_time,omitempty" yaml:"network_create_time,omitempty" gorm:"column:network_create_time;type:datetime;not null;"`           // 网络创建时间
	NetworkScope          string       `json:"network_scope,omitempty" yaml:"network_scope,omitempty" gorm:"column:network_scope;type:varchar(255);not null;"`                         // 网络作用域（local, global, swarm）
	NetworkDriver         string       `json:"network_driver,omitempty" yaml:"network_driver,omitempty" gorm:"column:network_driver;type:varchar(255);not null;"`                      // 网络驱动类型（bridge, overlay, macvlan等）
	NetworkEnableIPv6     sql.NullBool `json:"network_enable_ipv6,omitempty" yaml:"network_enable_ipv6,omitempty" gorm:"column:network_enable_ipv6;type:bool;not null;"`               // 是否启用IPv6
	NetworkIpam           string       `json:"network_ipam,omitempty" yaml:"network_ipam,omitempty" gorm:"column:network_ipam;type:text;not null;"`                                    // 网络IP地址管理配置（JSON格式）
	NetworkInternal       sql.NullBool `json:"network_internal,omitempty" yaml:"network_internal,omitempty" gorm:"column:network_internal;type:bool;not null;"`                        // 是否为内部网络
	NetworkAttachable     sql.NullBool `json:"network_attachable,omitempty" yaml:"network_attachable,omitempty" gorm:"column:network_attachable;type:bool;not null;"`                  // 是否可附加到独立容器
	NetworkIngress        sql.NullBool `json:"network_ingress,omitempty" yaml:"network_ingress,omitempty" gorm:"column:network_ingress;type:bool;not null;"`                           // 是否为入口网络（Swarm模式）
	NetworkLocationPeerId string       `json:"network_location_peer_id,omitempty" yaml:"network_location_peer_id,omitempty" gorm:"column:network_location_peer_id;type:varchar(255);"` // 网络所在节点的PeerID
	IsDelete              sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:bool;default:false"`                                         // 删除标记（默认false）
}

// TableName 返回数据库表名
// @description 实现gorm接口，指定Info结构体对应的数据库表名
// @return string 数据库表名
func (i Info) TableName() string {
	return networkInfoTableName
}

// String 将网络信息转换为JSON字符串
// @description 将Info结构体转换为JSON格式的字符串表示
// @return string 网络信息的JSON格式字符串
func (i Info) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

// NewNetworkInfo 创建新的网络信息实例
// @description 初始化一个空的Docker网络信息结构体指针
// @return *Info 网络信息结构体指针
func NewNetworkInfo() *Info {
	return &Info{}
}

// Mapper Docker网络数据访问层
// @description 提供Docker网络信息的数据库操作方法
// @struct
type Mapper struct {
	dbConn *gorm.DB // 数据库连接
}

// NewNetworkMapper 创建Docker网络数据访问实例
// @description 初始化Docker网络Mapper，设置数据库连接
// @param conn *gorm.DB 数据库连接对象
// @return *Mapper Docker网络数据访问层实例
func NewNetworkMapper(conn *gorm.DB) *Mapper {
	return &Mapper{dbConn: conn}
}

// InsertOrUpdateOne 插入或更新Docker网络信息
// @description 在事务中插入或更新Docker网络信息，根据网络ID和位置确定是否存在
// @param info *Info 要插入或更新的网络信息
// @return error 操作结果错误信息
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
