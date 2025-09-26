package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

type GrpcSendPingMessageRequest struct{}

func (g GrpcSendPingMessageRequest) IsLegal() bool {
	return true
}

func (g GrpcSendPingMessageRequest) String() string {
	str, _ := json.MarshalString(g)
	return str
}
