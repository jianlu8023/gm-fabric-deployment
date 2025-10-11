package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/jinzhu/copier"
)

// GrpcSendPingMessageResponse gRPC发送Ping消息响应
//
// @description gRPC发送Ping消息后的响应数据结构
// @struct
type GrpcSendPingMessageResponse struct {
	Success         bool   `json:"success" yaml:"success"`                   // 操作是否成功
	ResponseCode    int32  `json:"response_code" yaml:"response_code"`       // 响应代码
	ResponseMessage string `json:"response_message" yaml:"response_message"` // 响应消息
	Message         []byte `json:"message" yaml:"message"`                   // 消息内容
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 将GrpcSendPingMessageResponse结构体转换为JSON格式，特别处理Message字段为字符串
// @return []byte JSON字节数组
// @return error 错误信息
func (resp GrpcSendPingMessageResponse) MarshalJSON() ([]byte, error) {
	type Alias GrpcSendPingMessageResponse
	aux := struct {
		*Alias
		Message string `json:"message"`
	}{
		Alias:   (*Alias)(&resp),
		Message: string(resp.Message),
	}
	return json.Marshal(aux)
}

// String 将GrpcSendPingMessageResponse转换为字符串
//
// @description 将GrpcSendPingMessageResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp GrpcSendPingMessageResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewGrpcSendPingMessageResponse 创建gRPC发送Ping消息响应对象
//
// @description 创建一个新的gRPC发送Ping消息响应对象
// @param response *pb.BaseResponse gRPC基础响应对象
// @return *GrpcSendPingMessageResponse gRPC发送Ping消息响应对象
// @return error 错误信息
func NewGrpcSendPingMessageResponse(response *pb.BaseResponse) (*GrpcSendPingMessageResponse, error) {
	var convert GrpcSendPingMessageResponse
	if err := copier.CopyWithOption(&convert, response, copier.Option{
		DeepCopy:    true,
		IgnoreEmpty: true,
	}); err != nil {
		return nil, err
	}
	return &convert, nil
}

