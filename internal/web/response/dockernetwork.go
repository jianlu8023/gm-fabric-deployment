package response

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"github.com/jinzhu/copier"
)

type DockerNetworkListResponse struct {
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

func (i DockerNetworkListResponse) String() string {
	str, _ := json.MarshalString(i)
	return str
}

func (i DockerNetworkListResponse) MarshalJSON() ([]byte, error) {
	type Alias DockerNetworkListResponse
	aux := struct {
		*Alias
		NetworkEnableIPv6 bool `json:"network_enable_ipv6,omitempty" yaml:"network_enable_ipv6,omitempty"`
		NetworkInternal   bool `json:"network_internal,omitempty" yaml:"network_internal,omitempty"`
		NetworkAttachable bool `json:"network_attachable,omitempty" yaml:"network_attachable,omitempty"`
		NetworkIngress    bool `json:"network_ingress,omitempty" yaml:"network_ingress,omitempty"`
		IsDelete          bool `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
	}{
		Alias:             (*Alias)(&i),
		NetworkEnableIPv6: i.NetworkEnableIPv6.Bool,
		NetworkInternal:   i.NetworkInternal.Bool,
		NetworkAttachable: i.NetworkAttachable.Bool,
		NetworkIngress:    i.NetworkIngress.Bool,
		IsDelete:          i.IsDelete.Bool,
	}
	return json.Marshal(aux)
}

func NewDockerNetworkListResponse(page dbpage.Info[model.DockerNetwork]) (*dbpage.Info[DockerNetworkListResponse], error) {
	var convert []DockerNetworkListResponse

	err := copier.CopyWithOption(&convert, page.GetRecords(), copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})
	if err != nil {
		return nil, err
	}
	return dbpage.NewPageInfo[DockerNetworkListResponse](
		page.GetPageNo(),
		page.GetPageSize(),
		page.GetCount(),
		page.GetMaxPage(),
		convert,
	), nil
}
