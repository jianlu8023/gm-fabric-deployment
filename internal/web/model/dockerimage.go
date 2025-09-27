package model

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

const (
	imageInfoTableName = "t_docker_image_info"
)

// DockerImage Docker镜像信息模型
//
// @description 定义Docker镜像的数据结构，包含镜像的基本信息
// @struct
type DockerImage struct {
	AutoUid             int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                  // 自增id
	ImageName           string       `json:"image_name,omitempty" yaml:"image_name,omitempty" gorm:"column:image_name;type:varchar(255);"`                                     // 镜像名称
	ImageCreated        time.Time    `json:"image_created,omitempty" yaml:"image_created,omitempty" gorm:"column:image_created;type:datetime;"`                                // 镜像创建时间
	ImageLabels         string       `json:"image_labels,omitempty" yaml:"image_labels,omitempty" gorm:"column:image_labels;type:varchar(255);"`                               // 镜像标签
	ImageId             string       `json:"image_id,omitempty" yaml:"image_id,omitempty" gorm:"column:image_id;type:varchar(255);"`                                           // 镜像id
	ImageLocationPeerId string       `json:"image_location_peer_id,omitempty" yaml:"image_location_peer_id,omitempty" gorm:"column:image_location_peer_id;type:varchar(255);"` // 镜像所在peer
	IsDelete            sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:bool;default:false"`                                   // 是否删除
	// ImageFrom string `json:"image_from,omitempty" yaml:"image_from,omitempty" gorm:"column:image_from;type:varchar(255);"`
}

// TableName 返回数据库表名
//
// @description 实现gorm接口，指定DockerImage结构体对应的数据库表名
// @return string 数据库表名
func (DockerImage) TableName() string {
	return imageInfoTableName
}

// String 将Docker镜像信息转换为字符串表示
//
// @description 将DockerImage结构体转换为JSON格式的字符串
// @return string 镜像信息的JSON格式字符串
func (model DockerImage) String() string {
	str, _ := json.MarshalString(model)
	return str
}

// NewDockerImage 创建新的Docker镜像信息实例
//
// @description 初始化一个空的Docker镜像信息结构体指针
// @return *DockerImage Docker镜像信息结构体指针
func NewDockerImage() *DockerImage {
	return &DockerImage{}
}
