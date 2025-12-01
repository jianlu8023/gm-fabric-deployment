package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// Libp2pNodeListRequest Libp2p节点列表请求结构体
//
// @description 用于获取Libp2p节点列表
// @struct
type Libp2pNodeListRequest struct {
	commonhttp.PaginationRequest
}

// String 将Libp2p节点列表请求参数转换为字符串表示
//
// @description 将Libp2pNodeListRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req Libp2pNodeListRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证Libp2p节点列表请求参数是否合法
//
// @description 检查分页参数是否合法
// @return bool 参数是否合法
func (req Libp2pNodeListRequest) IsLegal() bool {
	if req.PageNo <= 0 || req.PageSize <= 0 {
		return false
	}
	return true
}

// Libp2pNodeMyselfRequest 获取当前Libp2p节点信息请求结构体
//
// @description 用于获取当前节点自身的Libp2p信息
// @struct
type Libp2pNodeMyselfRequest struct{}

// String 将获取当前Libp2p节点信息请求参数转换为字符串表示
//
// @description 将Libp2pNodeMyselfRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req Libp2pNodeMyselfRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证获取当前Libp2p节点信息请求参数是否合法
//
// @description 由于无参数，直接返回true
// @return bool 参数是否合法
func (req Libp2pNodeMyselfRequest) IsLegal() bool {
	return true
}
