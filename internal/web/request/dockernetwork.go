package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// DockerNetworkListRequest Docker网络列表请求参数
// @description 用于获取Docker网络列表的请求参数
// @struct
// @property commonhttp.PaginationRequest 分页请求参数
// @property PeerId string 节点ID
// @extends commonhttp.PaginationRequest
type DockerNetworkListRequest struct {
	commonhttp.PaginationRequest
	PeerId string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"-"` // 节点ID
}

// IsLegal 验证Docker网络列表请求参数是否合法
// @description 检查DockerNetworkListRequest结构体中的分页参数是否满足基本要求
// @return bool 参数是否合法
func (req DockerNetworkListRequest) IsLegal() bool {
	if req.PageSize <= 0 || req.PageNo <= 0 {
		return false
	}
	return true
}

// String 将Docker网络列表请求参数转换为字符串表示
// @description 将DockerNetworkListRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req DockerNetworkListRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
