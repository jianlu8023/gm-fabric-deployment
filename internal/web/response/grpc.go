package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/jinzhu/copier"
)

type GrpcSendPingMessageResponse struct {
	Success         bool   `json:"success,omitempty"`
	ResponseCode    int32  `json:"response_code,omitempty"`
	ResponseMessage string `json:"response_message,omitempty"`
	Message         []byte `json:"message,omitempty"`
}

func (resp GrpcSendPingMessageResponse) MarshalJSON() ([]byte, error) {
	type Alias GrpcSendPingMessageResponse
	aux := struct {
		*Alias
		Message string `json:"message,omitempty"`
	}{
		Alias:   (*Alias)(&resp),
		Message: string(resp.Message),
	}
	return json.Marshal(aux)
}

func (resp GrpcSendPingMessageResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

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
