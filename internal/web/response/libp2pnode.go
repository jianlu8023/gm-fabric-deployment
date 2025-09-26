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

type Libp2pNodeListResponse struct {
	NodeId               string       `json:"node_id,omitempty" yaml:"node_id,omitempty" gorm:"column:node_id;type:varchar(256);unique;"`                                      // 节点ID（唯一标识）
	NodeIp               string       `json:"node_ip,omitempty" yaml:"node_ip,omitempty" gorm:"column:node_ip;type:varchar(256);"`                                             // 节点ip
	LastAliveMessageTime time.Time    `json:"last_alive_message_time,omitempty" yaml:"last_alive_message_time,omitempty" gorm:"column:last_alive_message_time;type:datetime;"` // 最后一次收到心跳时间
	IsAlive              sql.NullBool `json:"is_alive,omitempty" yaml:"is_alive,omitempty" gorm:"column:is_alive;type:bool;"`                                                  // 是否存活状态
	IsMySelf             sql.NullBool `json:"is_my_self,omitempty" yaml:"is_my_self,omitempty" gorm:"column:is_my_self;type:bool;default:false"`                               // 是否是本机节点
}

func (resp Libp2pNodeListResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func (resp Libp2pNodeListResponse) MarshalJSON() ([]byte, error) {
	type Alias Libp2pNodeListResponse

	aux := struct {
		*Alias
		IsAlive              bool   `json:"is_alive,omitempty" yaml:"is_alive,omitempty"`
		IsMySelf             bool   `json:"is_my_self,omitempty" yaml:"is_my_self,omitempty"`
		LastAliveMessageTime string `json:"last_alive_message_time,omitempty" yaml:"last_alive_message_time,omitempty"`
	}{
		Alias:                (*Alias)(&resp),
		IsAlive:              resp.IsAlive.Bool,
		IsMySelf:             resp.IsMySelf.Bool,
		LastAliveMessageTime: humantime.HumanTimeLower(resp.LastAliveMessageTime, "unknown"),
	}
	return json.Marshal(aux)
}

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

type Libp2pNodeMyselfResponse struct {
	NodeId               string       `json:"node_id,omitempty" yaml:"node_id,omitempty"`                                 // 节点ID（唯一标识）
	NodeIp               string       `json:"node_ip,omitempty" yaml:"node_ip,omitempty"`                                 // 节点ip
	LastAliveMessageTime time.Time    `json:"last_alive_message_time,omitempty" yaml:"last_alive_message_time,omitempty"` // 最后一次收到心跳时间
	IsAlive              sql.NullBool `json:"is_alive,omitempty" yaml:"is_alive,omitempty"`                               // 是否存活状态
	IsMySelf             sql.NullBool `json:"is_my_self,omitempty" yaml:"is_my_self,omitempty"`                           // 是否是本机节点
}

func (resp Libp2pNodeMyselfResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func (resp Libp2pNodeMyselfResponse) MarshalJSON() ([]byte, error) {
	type Alias Libp2pNodeMyselfResponse

	aux := struct {
		*Alias
		IsAlive              bool   `json:"is_alive,omitempty" yaml:"is_alive,omitempty"`
		IsMySelf             bool   `json:"is_my_self,omitempty" yaml:"is_my_self,omitempty"`
		LastAliveMessageTime string `json:"last_alive_message_time,omitempty" yaml:"last_alive_message_time,omitempty"`
	}{
		Alias:                (*Alias)(&resp),
		IsAlive:              resp.IsAlive.Bool,
		IsMySelf:             resp.IsMySelf.Bool,
		LastAliveMessageTime: humantime.HumanTimeLower(resp.LastAliveMessageTime, "unknown"),
	}
	return json.Marshal(aux)
}

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
