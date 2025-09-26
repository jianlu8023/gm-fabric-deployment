package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

type DockerNetworkListRequest struct {
	commonhttp.PaginationRequest
	PeerId string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"-"`
}

func (req DockerNetworkListRequest) IsLegal() bool {
	if req.PageSize <= 0 || req.PageNo <= 0 {
		return false
	}
	return true
}

func (req DockerNetworkListRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
