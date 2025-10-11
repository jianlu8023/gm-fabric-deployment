package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// GrpcSendPingMessageRequest GRPC发送Ping消息请求参数
//
// @description 用于通过GRPC协议发送Ping消息的请求参数
// @struct
type GrpcSendPingMessageRequest struct{}

// IsLegal 验证GRPC发送Ping消息请求参数是否合法
//
// @description 检查GrpcSendPingMessageRequest结构体中的参数是否满足基本要求
// @return bool 参数是否合法
func (req GrpcSendPingMessageRequest) IsLegal() bool {
	// 当前版本不接收任何参数，直接返回合法
	return true
}

// String 将GRPC发送Ping消息请求参数转换为字符串表示
//
// @description 将GrpcSendPingMessageRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req GrpcSendPingMessageRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
