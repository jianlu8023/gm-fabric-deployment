package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// WSConnectRequest WebSocket连接请求结构体
// @description 用于在建立WebSocket连接前传递必要的参数
// @struct
// @property NodeID string 节点ID (必需)
type WSConnectRequest struct {
	NodeID string `json:"node_id,omitempty" yaml:"node_id,omitempty" form:"nodeId" binding:"required"` // 节点ID
}

// String 将WebSocket连接请求参数转换为字符串表示
// @description 将WSConnectRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req WSConnectRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证WebSocket连接请求参数是否合法
// @description 检查NodeID是否为空
// @return bool 参数是否合法
func (req WSConnectRequest) IsLegal() bool {
	if stringer.IsBlank(req.NodeID) {
		return false
	}
	return true
}

// WSDisconnectRequest WebSocket断开连接请求结构体
// @description 用于处理WebSocket连接断开请求
// @struct
// @property NodeID string 节点ID (必需)
type WSDisconnectRequest struct {
	NodeID string `json:"node_id,omitempty" yaml:"node_id,omitempty" form:"nodeId" binding:"required"` // 节点ID
}

// IsLegal 验证WebSocket断开连接请求参数是否合法
// @description 检查NodeID是否为空
// @return bool 参数是否合法
func (req WSDisconnectRequest) IsLegal() bool {
	if stringer.IsBlank(req.NodeID) {
		return false
	}
	return true
}

// String 将WebSocket断开连接请求参数转换为字符串表示
// @description 将WSDisconnectRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req WSDisconnectRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// WSMessageRequest WebSocket消息请求结构体
// @description 用于客户端向服务器发送WebSocket消息
// @struct
// @property Message string 消息内容 (必需)
// @property NodeID string 目标节点ID (可选)
type WSMessageRequest struct {
	Message string `json:"message,omitempty" yaml:"message,omitempty" form:"message" binding:"required"` // 消息内容
	NodeID  string `json:"node_id,omitempty" yaml:"node_id,omitempty" form:"nodeId" binding:"-"`         // 目标节点ID，可选
}

// IsLegal 验证WebSocket消息请求参数是否合法
// @description 检查Message是否为空
// @return bool 参数是否合法
func (req WSMessageRequest) IsLegal() bool {
	if stringer.IsBlank(req.Message) {
		return false
	}
	return true
}

// String 将WebSocket消息请求参数转换为字符串表示
// @description 将WSMessageRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req WSMessageRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// WSConnectionListRequest WebSocket连接列表请求结构体
// @description 用于获取WebSocket连接列表
// @struct
// @property PaginationRequest 分页请求参数
type WSConnectionListRequest struct {
	commonhttp.PaginationRequest
}

// IsLegal 验证WebSocket连接列表请求参数是否合法
// @description 检查分页参数是否合法
// @return bool 参数是否合法
func (req WSConnectionListRequest) IsLegal() bool {
	if req.PageNo <= 0 || req.PageSize <= 0 {
		return false
	}
	return true
}

// String 将WebSocket连接列表请求参数转换为字符串表示
// @description 将WSConnectionListRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req WSConnectionListRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
