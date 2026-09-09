package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// CertsTlsRequest 当前TLS连接信息请求结构体
//
// @description 用于查询当前 HTTPS 连接的 TLS/mTLS 协商信息，无请求参数
// @struct
type CertsTlsRequest struct {
}

// String 将TLS连接信息请求参数转换为字符串表示
//
// @description 将 CertsTlsRequest 结构体转换为 JSON 格式的字符串
// @return string 请求参数的 JSON 格式字符串
func (req CertsTlsRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证TLS连接信息请求参数是否合法
//
// @description 由于无参数，直接返回 true
// @return bool 参数是否合法
func (req CertsTlsRequest) IsLegal() bool {
	return true
}
