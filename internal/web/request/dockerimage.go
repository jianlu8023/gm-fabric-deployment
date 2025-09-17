package request

import (
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/json"
	"github.com/jianlu8023/golang-example/pkg/str"
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
func (n *DockerImageListRequest) String() string {
	bytes, _ := json.Marshal(n)
	return string(bytes)
}

type DockerImagePullRequest struct {
	ImageName string `json:"image_name,omitempty" yaml:"image_name,omitempty" form:"imageName" binding:"required"`
	PeerId    string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"required"`
}

func (n *DockerImagePullRequest) IsLegal() bool {
	if str.IsBlank(n.ImageName) ||
		str.IsBlank(n.PeerId) {
		return false
	}
	return true
}

func (n *DockerImagePullRequest) String() string {
	bytes, _ := json.Marshal(n)
	return string(bytes)
}
