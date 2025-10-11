package response

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"github.com/jinzhu/copier"
)

// Libp2pNodeListResponse Libp2p节点列表响应结构体
//
// @description 用于返回Libp2p节点列表信息
// @struct
type Libp2pNodeListResponse struct {
	NodeId               string       `json:"node_id" yaml:"node_id"`                                 // 节点ID（唯一标识）
	NodeIp               string       `json:"node_ip" yaml:"node_ip"`                                 // 节点ip
	LastAliveMessageTime time.Time    `json:"last_alive_message_time" yaml:"last_alive_message_time"` // 最后一次收到心跳时间
	IsAlive              sql.NullBool `json:"is_alive" yaml:"is_alive"`                               // 是否存活状态
	IsMySelf             sql.NullBool `json:"is_my_self" yaml:"is_my_self"`                           // 是否是本机节点
}

// String 将Libp2p节点列表响应转换为字符串表示
//
// @description 将Libp2pNodeListResponse结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp Libp2pNodeListResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 自定义Libp2pNodeListResponse结构体的JSON序列化逻辑，处理sql.NullBool和time.Time类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp Libp2pNodeListResponse) MarshalJSON() ([]byte, error) {
	type Alias Libp2pNodeListResponse

	aux := struct {
		*Alias
		IsAlive              bool   `json:"is_alive"`
		IsMySelf             bool   `json:"is_my_self"`
		LastAliveMessageTime string `json:"last_alive_message_time"`
	}{
		Alias:                (*Alias)(&resp),
		IsAlive:              resp.IsAlive.Bool,
		IsMySelf:             resp.IsMySelf.Bool,
		LastAliveMessageTime: humantime.HumanTimeLower(resp.LastAliveMessageTime, "unknown"),
	}
	return json.Marshal(aux)
}

// ToLibp2pNodeListResponse 将模型转换为响应结构体
//
// @description 将model.Libp2pNode的分页结果转换为Libp2pNodeListResponse的分页结果
// @param page dbpage.Info[model.Libp2pNode] 模型的分页结果
// @return *dbpage.Info[Libp2pNodeListResponse] 响应的分页结果
// @return error 转换错误信息
func ToLibp2pNodeListResponse(page dbpage.Info[model.Libp2pNode]) (*dbpage.Info[Libp2pNodeListResponse], error) {
	var convert []Libp2pNodeListResponse

	err := copier.CopyWithOption(&convert, page.GetRecords(), copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})
	if err != nil {
		return nil, err
	}
	return dbpage.NewPageInfo[Libp2pNodeListResponse](
		page.GetPageNo(),
		page.GetPageSize(),
		page.GetCount(),
		page.GetMaxPage(),
		convert,
	), nil
}

// Libp2pNodeMyselfResponse 当前Libp2p节点信息响应结构体
//
// @description 用于返回当前节点自身的Libp2p信息
// @struct
type Libp2pNodeMyselfResponse struct {
	NodeId               string       `json:"node_id" yaml:"node_id"`                                 // 节点ID（唯一标识）
	NodeIp               string       `json:"node_ip" yaml:"node_ip"`                                 // 节点ip
	LastAliveMessageTime time.Time    `json:"last_alive_message_time" yaml:"last_alive_message_time"` // 最后一次收到心跳时间
	IsAlive              sql.NullBool `json:"is_alive" yaml:"is_alive"`                               // 是否存活状态
	IsMySelf             sql.NullBool `json:"is_my_self" yaml:"is_my_self"`                           // 是否是本机节点
}

// String 将当前Libp2p节点信息响应转换为字符串表示
//
// @description 将Libp2pNodeMyselfResponse结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp Libp2pNodeMyselfResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 自定义Libp2pNodeMyselfResponse结构体的JSON序列化逻辑，处理sql.NullBool和time.Time类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp Libp2pNodeMyselfResponse) MarshalJSON() ([]byte, error) {
	type Alias Libp2pNodeMyselfResponse

	aux := struct {
		*Alias
		IsAlive              bool   `json:"is_alive"`
		IsMySelf             bool   `json:"is_my_self"`
		LastAliveMessageTime string `json:"last_alive_message_time"`
	}{
		Alias:                (*Alias)(&resp),
		IsAlive:              resp.IsAlive.Bool,
		IsMySelf:             resp.IsMySelf.Bool,
		LastAliveMessageTime: humantime.HumanTimeLower(resp.LastAliveMessageTime, "unknown"),
	}
	return json.Marshal(aux)
}

// ToLibp2pNodeMyselfResponse 将模型转换为响应结构体
//
// @description 将model.Libp2pNode转换为Libp2pNodeMyselfResponse
// @param node model.Libp2pNode 节点模型
// @return Libp2pNodeMyselfResponse 节点响应
// @return error 转换错误信息
func ToLibp2pNodeMyselfResponse(node model.Libp2pNode) (Libp2pNodeMyselfResponse, error) {
	var convert Libp2pNodeMyselfResponse
	if err := copier.CopyWithOption(&convert, &node, copier.Option{
		DeepCopy:    true,
		IgnoreEmpty: true,
	}); err != nil {
		return Libp2pNodeMyselfResponse{}, err
	}
	return convert, nil
}
