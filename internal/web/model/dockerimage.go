package model

import (
	"database/sql"
	"github.com/jianlu8023/golang-example/pkg/json"
)

const (
	imageInfoTableName = "t_docker_image_info"
)

type DockerImage struct {
	AutoUid             int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                  // 自增id
	ImageName           string       `json:"image_name,omitempty" yaml:"image_name,omitempty" gorm:"column:image_name;type:varchar(255);"`                                     // 镜像名称
	ImageCreated        int64        `json:"image_created,omitempty" yaml:"image_created,omitempty" gorm:"column:image_created;type:bigint(20);"`                              // 镜像创建时间
	ImageLabels         string       `json:"image_labels,omitempty" yaml:"image_labels,omitempty" gorm:"column:image_labels;type:varchar(255);"`                               // 镜像标签
	ImageId             string       `json:"image_id,omitempty" yaml:"image_id,omitempty" gorm:"column:image_id;type:varchar(255);"`                                           // 镜像id
	ImageLocationPeerId string       `json:"image_location_peer_id,omitempty" yaml:"image_location_peer_id,omitempty" gorm:"column:image_location_peer_id;type:varchar(255);"` // 镜像所在peer
	IsDelete            sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:bool;default:false"`                                   // 是否删除
	// ImageFrom string `json:"image_from,omitempty" yaml:"image_from,omitempty" gorm:"column:image_from;type:varchar(255);"`
}

// MarshalJSON 自定义json返回
func (i *DockerImage) MarshalJSON() ([]byte, error) {
	type Alias DockerImage
	aux := struct {
		*Alias
		IsDelete bool `json:"is_delete"`
	}{
		Alias:    (*Alias)(i),
		IsDelete: i.IsDelete.Bool,
	}
	return json.Marshal(aux)
}

// TableName 返回表名
// @return string 表名
func (i *DockerImage) TableName() string {
	return imageInfoTableName
}

// String 返回json字符串
// @return string json字符串
func (i *DockerImage) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

// NewDockerImage 返回一个新的镜像信息
// @return *Info 镜像信息
func NewDockerImage() *DockerImage {
	return &DockerImage{}
}
