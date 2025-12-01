package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// DockerImageListRequest Docker镜像列表请求参数
//
// @description 用于获取Docker镜像列表的请求参数
// @struct
type DockerImageListRequest struct {
	commonhttp.PaginationRequest
	PeerId string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"-"` // PeerId 节点ID
}

// IsLegal 验证Docker镜像列表请求参数是否合法
//
// @description 检查DockerImageListRequest结构体中的分页参数是否满足基本要求
// @return bool 参数是否合法
func (req DockerImageListRequest) IsLegal() bool {
	if req.PageSize <= 0 || req.PageNo <= 0 {
		return false
	}
	return true
}

// String 将Docker镜像列表请求参数转换为字符串表示
//
// @description 将DockerImageListRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req DockerImageListRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// DockerImagePullRequest Docker镜像拉取请求参数
//
// @description 用于从指定节点拉取Docker镜像的请求参数
// @struct
type DockerImagePullRequest struct {
	ImageName string `json:"image_name,omitempty" yaml:"image_name,omitempty" form:"imageName" binding:"required"` // ImageName 镜像名称（必填）
	PeerId    string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" form:"peerId" binding:"required"`          // PeerId 节点ID（必填）
	Platform  string `json:"platform,omitempty" yaml:"platform,omitempty" form:"platform" binding:"-"`             // Platform 平台
}

// IsLegal 验证Docker镜像拉取请求参数是否合法
//
// @description 检查DockerImagePullRequest结构体中的必要参数是否满足基本要求
// @return bool 参数是否合法
func (req DockerImagePullRequest) IsLegal() bool {
	if stringer.IsBlank(req.ImageName) ||
		stringer.IsBlank(req.PeerId) {
		return false
	}
	return true
}

// String 将Docker镜像拉取请求参数转换为字符串表示
//
// @description 将DockerImagePullRequest结构体转换为JSON格式的字符串
// @return string 请求数据的JSON格式字符串
func (req DockerImagePullRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
