package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// WSConnectRequest 表示WebSocket连接请求的参数
// 用于在建立WebSocket连接前传递必要的参数
type WSConnectRequest struct {
	NodeID string `json:"node_id" form:"nodeId" binding:"required"` // 节点ID
}

func (r WSConnectRequest) String() string {
	str, _ := json.MarshalString(r)
	return str
}

// IsLegal 验证WSConnectRequest参数是否合法
func (r WSConnectRequest) IsLegal() bool {
	if stringer.IsBlank(r.NodeID) {
		return false
	}
	return true
}

type WSDisconnectRequest struct {
	NodeID string `json:"node_id" form:"nodeId" binding:"required"` // 节点ID
}

// IsLegal 验证WSConnectRequest参数是否合法
func (r WSDisconnectRequest) IsLegal() bool {
	if stringer.IsBlank(r.NodeID) {
		return false
	}
	return true
}
func (r WSDisconnectRequest) String() string {
	str, _ := json.MarshalString(r)
	return str
}

// WSMessageRequest 表示WebSocket消息请求的参数
// 用于客户端向服务器发送消息
type WSMessageRequest struct {
	Message string `json:"message" binding:"required"` // 消息内容
	NodeID  string `json:"node_id"`                    // 目标节点ID，可选
}

// IsLegal 验证WSMessageRequest参数是否合法
func (r WSMessageRequest) IsLegal() bool {
	if stringer.IsBlank(r.Message) {
		return false
	}
	return true
}
func (r WSMessageRequest) String() string {
	str, _ := json.MarshalString(r)
	return str
}

// WSConnectionListRequest WebSocket连接列表请求
type WSConnectionListRequest struct {
	commonhttp.PaginationRequest
}

// IsLegal 验证WSConnectionListRequest参数是否合法
func (r WSConnectionListRequest) IsLegal() bool {
	if r.PageNo <= 0 || r.PageSize <= 0 {
		return false
	}
	return true
}
func (r WSConnectionListRequest) String() string {
	str, _ := json.MarshalString(r)
	return str
}
