package response

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"github.com/jinzhu/copier"
)

type DockerImageListResponse struct {
	ImageName           string       `json:"image_name,omitempty" yaml:"image_name,omitempty" gorm:"column:image_name;type:varchar(255);"`                                     // 镜像名称
	ImageCreated        time.Time    `json:"image_created,omitempty" yaml:"image_created,omitempty" gorm:"column:image_created;type:datetime;"`                                // 镜像创建时间
	ImageLabels         string       `json:"image_labels,omitempty" yaml:"image_labels,omitempty" gorm:"column:image_labels;type:varchar(255);"`                               // 镜像标签
	ImageId             string       `json:"image_id,omitempty" yaml:"image_id,omitempty" gorm:"column:image_id;type:varchar(255);"`                                           // 镜像id
	ImageLocationPeerId string       `json:"image_location_peer_id,omitempty" yaml:"image_location_peer_id,omitempty" gorm:"column:image_location_peer_id;type:varchar(255);"` // 镜像所在peer
	IsDelete            sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:bool;default:false"`                                   // 是否删除
}

// MarshalJSON 自定义json返回
func (resp DockerImageListResponse) MarshalJSON() ([]byte, error) {
	type Alias DockerImageListResponse
	aux := struct {
		*Alias
		IsDelete     bool   `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
		ImageCreated string `json:"image_created,omitempty" yaml:"image_created,omitempty"`
	}{
		Alias:        (*Alias)(&resp),
		IsDelete:     resp.IsDelete.Bool,
		ImageCreated: humantime.HumanTime(resp.ImageCreated, "unknown"),
	}
	return json.Marshal(aux)
}

// String 返回json字符串
// @return string json字符串
func (resp DockerImageListResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func NewDockerImageListResponse(page dbpage.Info[model.DockerImage]) (*dbpage.Info[DockerImageListResponse], error) {
	var convert []DockerImageListResponse

	err := copier.CopyWithOption(&convert, page.GetRecords(), copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})
	if err != nil {
		return nil, err
	}
	return dbpage.NewPageInfo[DockerImageListResponse](
		page.GetPageNo(),
		page.GetPageSize(),
		page.GetCount(),
		page.GetMaxPage(),
		convert,
	), nil
}

type DockerImagePullResponse struct {
	Msg string `json:"msg,omitempty" yaml:"msg,omitempty"`
}

func (resp DockerImagePullResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(resp)
}

func (resp DockerImagePullResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func NewDockerImagePullResponse(msg string) *DockerImagePullResponse {
	return &DockerImagePullResponse{
		Msg: msg,
	}
}
