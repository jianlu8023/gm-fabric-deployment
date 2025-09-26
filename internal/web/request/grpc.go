package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

type GrpcSendPingMessageRequest struct{}

func (req GrpcSendPingMessageRequest) IsLegal() bool {
	return true
}

func (req GrpcSendPingMessageRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
