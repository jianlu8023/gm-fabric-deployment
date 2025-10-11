package response

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"github.com/jinzhu/copier"
)

// DockerNetworkListResponse Docker网络列表响应结构体
//
// @description 用于返回Docker网络列表的响应数据
// @struct
type DockerNetworkListResponse struct {
	NetworkName           string       `json:"network_name" yaml:"network_name"`                         // 网络名称
	NetworkID             string       `json:"network_id" yaml:"network_id"`                             // 网络ID（Docker生成的唯一标识）
	NetworkCreateTime     time.Time    `json:"network_create_time" yaml:"network_create_time"`           // 网络创建时间
	NetworkScope          string       `json:"network_scope" yaml:"network_scope"`                       // 网络作用域（local, global, swarm）
	NetworkDriver         string       `json:"network_driver" yaml:"network_driver"`                     // 网络驱动类型（bridge, overlay, macvlan等）
	NetworkEnableIPv6     sql.NullBool `json:"network_enable_ipv6" yaml:"network_enable_ipv6"`           // 是否启用IPv6
	NetworkIpam           string       `json:"network_ipam" yaml:"network_ipam"`                         // 网络IP地址管理配置（JSON格式）
	NetworkInternal       sql.NullBool `json:"network_internal" yaml:"network_internal"`                 // 是否为内部网络
	NetworkAttachable     sql.NullBool `json:"network_attachable" yaml:"network_attachable"`             // 是否可附加到独立容器
	NetworkIngress        sql.NullBool `json:"network_ingress" yaml:"network_ingress"`                   // 是否为入口网络（Swarm模式）
	NetworkLocationPeerId string       `json:"network_location_peer_id" yaml:"network_location_peer_id"` // 网络所在节点的PeerID
	IsDelete              sql.NullBool `json:"is_delete" yaml:"is_delete"`                               // 删除标记（默认false）
}

// String 将Docker网络列表响应转换为字符串表示
//
// @description 将DockerNetworkListResponse结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp DockerNetworkListResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 自定义DockerNetworkListResponse结构体的JSON序列化逻辑，处理sql.NullBool类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp DockerNetworkListResponse) MarshalJSON() ([]byte, error) {
	type Alias DockerNetworkListResponse
	aux := struct {
		*Alias
		NetworkEnableIPv6 bool `json:"network_enable_ipv6"`
		NetworkInternal   bool `json:"network_internal"`
		NetworkAttachable bool `json:"network_attachable"`
		NetworkIngress    bool `json:"network_ingress"`
		IsDelete          bool `json:"is_delete"`
	}{
		Alias:             (*Alias)(&resp),
		NetworkEnableIPv6: resp.NetworkEnableIPv6.Bool,
		NetworkInternal:   resp.NetworkInternal.Bool,
		NetworkAttachable: resp.NetworkAttachable.Bool,
		NetworkIngress:    resp.NetworkIngress.Bool,
		IsDelete:          resp.IsDelete.Bool,
	}
	return json.Marshal(aux)
}

// NewDockerNetworkListResponse 创建Docker网络列表响应对象
//
// @description 将数据库模型转换为API响应对象，并创建分页信息
// @param page dbpage.Info[model.DockerNetwork] 数据库查询得到的分页网络数据
// @return *dbpage.Info[DockerNetworkListResponse] 转换后的分页响应数据
// @return error 转换过程中的错误信息
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

