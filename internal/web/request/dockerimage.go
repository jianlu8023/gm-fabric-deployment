package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

type DockerImageListRequest struct {
	commonhttp.PaginationRequest
	PeerId string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"-"`
}

func (req DockerImageListRequest) IsLegal() bool {
	if req.PageSize <= 0 || req.PageNo <= 0 {
		return false
	}
	return true
}
func (req DockerImageListRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

type DockerImagePullRequest struct {
	ImageName string `json:"image_name,omitempty" yaml:"image_name,omitempty" form:"imageName" binding:"required"`
	PeerId    string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"required"`
}

func (req DockerImagePullRequest) IsLegal() bool {
	if stringer.IsBlank(req.ImageName) ||
		stringer.IsBlank(req.PeerId) {
		return false
	}
	return true
}

func (req DockerImagePullRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
