package image

import (
	"database/sql"
	"errors"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"gorm.io/gorm"
)

const (
	imageInfoTableName = "t_docker_image_info"
)

// Info docker镜像表结构
type Info struct {
	AutoUid             int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                  // 自增id
	ImageName           string       `json:"image_name,omitempty" yaml:"image_name,omitempty" gorm:"column:image_name;type:varchar(255);"`                                     // 镜像名称
	ImageCreated        int64        `json:"image_created,omitempty" yaml:"image_created,omitempty" gorm:"column:image_created;type:bigint(20);"`                              // 镜像创建时间
	ImageLabels         string       `json:"image_labels,omitempty" yaml:"image_labels,omitempty" gorm:"column:image_labels;type:varchar(255);"`                               // 镜像标签
	ImageId             string       `json:"image_id,omitempty" yaml:"image_id,omitempty" gorm:"column:image_id;type:varchar(255);"`                                           // 镜像id
	ImageLocationPeerId string       `json:"image_location_peer_id,omitempty" yaml:"image_location_peer_id,omitempty" gorm:"column:image_location_peer_id;type:varchar(255);"` // 镜像所在peer
	IsDelete            sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:bool;default:false"`                                   // 是否删除
	// ImageFrom string `json:"image_from,omitempty" yaml:"image_from,omitempty" gorm:"column:image_from;type:varchar(255);"`
}

// TableName 返回表名
// @return string 表名
func (i Info) TableName() string {
	return imageInfoTableName
}

// String 返回json字符串
// @return string json字符串
func (i Info) String() string {
	bytes, _ := json.Marshal(i)
	return string(bytes)
}

// NewImageInfo 返回一个新的镜像信息
// @return *Info 镜像信息
func NewImageInfo() *Info {
	return &Info{}
}

type Mapper struct {
	dbConn *gorm.DB
}

// NewImageMapper 返回一个新的镜像信息映射器
func NewImageMapper(conn *gorm.DB) *Mapper {
	return &Mapper{
		dbConn: conn,
	}
}

// InsertOneWithCheck 插入一条镜像信息，如果存在则返回datasource.ErrAlreadyExists
// @param imageInfo *Info 镜像信息
// @return error 错误信息
func (m *Mapper) InsertOneWithCheck(imageInfo *Info) error {
	if m.dbConn == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.dbConn.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&Info{}).Where(&Info{
			ImageName:           imageInfo.ImageName,
			ImageLocationPeerId: imageInfo.ImageLocationPeerId,
			IsDelete:            sql.NullBool{Bool: false, Valid: true},
		}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			// 镜像已经存在
			return datasource.ErrAlreadyExists
		}
		// 镜像不存在，插入
		if err := tx.Model(&Info{}).Create(imageInfo).Error; err != nil {
			return err
		}
		return nil
	})
}

func (m *Mapper) InsertOrUpdateOne(info *Info) error {
	if m.dbConn == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.dbConn.Transaction(func(tx *gorm.DB) error {
		var existInfo Info
		if err := tx.Model(&Info{}).Where(&Info{
			ImageName:           info.ImageName,
			ImageLocationPeerId: info.ImageLocationPeerId,
			IsDelete:            sql.NullBool{Bool: false, Valid: true},
		}).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在
				if err := tx.Model(&Info{}).Create(info).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新
			info.AutoUid = existInfo.AutoUid
			if err := tx.Model(&Info{}).Where(&Info{AutoUid: info.AutoUid}).Updates(info).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
