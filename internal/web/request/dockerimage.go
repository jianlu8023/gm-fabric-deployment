package request

import (
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
)

type DockerImageListRequest struct {
	commonhttp.PaginationRequest
	PeerId string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"-"`
}

func (n *DockerImageListRequest) IsLegal() bool {
	if n.PageSize <= 0 || n.PageNo <= 0 {
		return false
	}
	return true
}
