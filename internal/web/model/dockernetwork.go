package model

import (
	"database/sql"
	"github.com/jianlu8023/golang-example/pkg/json"
	"time"
)

const (
	networkInfoTableName = "t_docker_network_info"
)

type DockerNetwork struct {
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

func (i *DockerNetwork) MarshalJSON() ([]byte, error) {
	type Alias DockerNetwork
	aux := struct {
		*Alias
		NetworkEnableIPv6 bool `json:"network_enable_ipv6,omitempty" yaml:"network_enable_ipv6,omitempty"`
		NetworkInternal   bool `json:"network_internal,omitempty" yaml:"network_internal,omitempty"`
		NetworkAttachable bool `json:"network_attachable,omitempty" yaml:"network_attachable,omitempty"`
		NetworkIngress    bool `json:"network_ingress,omitempty" yaml:"network_ingress,omitempty"`
		IsDelete          bool `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
	}{
		Alias:             (*Alias)(i),
		NetworkEnableIPv6: i.NetworkEnableIPv6.Bool,
		NetworkInternal:   i.NetworkInternal.Bool,
		NetworkAttachable: i.NetworkAttachable.Bool,
		NetworkIngress:    i.NetworkIngress.Bool,
		IsDelete:          i.IsDelete.Bool,
	}
	return json.Marshal(aux)
}

// TableName 返回数据库表名
// @description 实现gorm接口，指定Info结构体对应的数据库表名
// @return string 数据库表名
func (i *DockerNetwork) TableName() string {
	return networkInfoTableName
}

// String 将网络信息转换为JSON字符串
// @description 将Info结构体转换为JSON格式的字符串表示
// @return string 网络信息的JSON格式字符串
func (i *DockerNetwork) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

// NewDockerNetwork 创建新的网络信息实例
// @description 初始化一个空的Docker网络信息结构体指针
// @return *Info 网络信息结构体指针
func NewDockerNetwork() *DockerNetwork {
	return &DockerNetwork{}
}
