package mapper

import (
	"database/sql"
	"errors"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/dbpage"
	"gorm.io/gorm"
	"math"
)

type DockerImageMapper struct {
	*Mapper
}

func NewDockerImageMapper(baseMapper *Mapper) *DockerImageMapper {
	return &DockerImageMapper{
		Mapper: baseMapper,
	}
}

// InsertOneWithCheck 插入一条镜像信息，如果存在则返回datasource.ErrAlreadyExists
// @param imageInfo *Info 镜像信息
// @return error 错误信息
func (m *Mapper) InsertOneWithCheck(imageInfo *model.DockerImage) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.DockerImage{}).Where(&model.DockerImage{
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
		if err := tx.Model(&model.DockerImage{}).
			Create(imageInfo).Error; err != nil {
			return err
		}
		return nil
	})
}

// InsertOrUpdateOne 插入或更新Docker镜像信息
// @description 在事务中插入或更新Docker镜像信息，根据名称和位置确定是否存在
// @param info *Info 要插入或更新的镜像信息
// @return error 操作结果错误信息
// InsertOrUpdateOne 插入或更新Docker镜像信息
// @description 在事务中插入或更新Docker镜像信息，根据名称和位置确定是否存在
// @param info *Info 要插入或更新的镜像信息
// @return error 操作结果错误信息
func (m *DockerImageMapper) InsertOrUpdateOne(info *model.DockerImage) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var existInfo model.DockerImage
		if err := tx.Model(&model.DockerImage{}).Where(&model.DockerImage{
			ImageName:           info.ImageName,
			ImageLocationPeerId: info.ImageLocationPeerId,
			IsDelete:            sql.NullBool{Bool: false, Valid: true},
		}).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在
				if err := tx.Model(&model.DockerImage{}).
					Create(info).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新
			info.AutoUid = existInfo.AutoUid
			if err := tx.Model(&model.DockerImage{}).
				Where(&model.DockerImage{AutoUid: info.AutoUid}).
				Updates(info).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *DockerImageMapper) DockerImageList(query model.DockerImage, isPage bool, pageNo int, pageSize int) (dbpage.Info[model.DockerImage], error) {
	page := dbpage.Info[model.DockerImage]{}
	if m.db == nil {
		return page, datasource.ErrNoDataSourceConn
	}
	page.PageNo = int64(pageNo)
	page.PageSize = int64(pageSize)

	// 计算偏移量
	offset := (pageNo - 1) * pageSize

	// 构建查询
	db := m.db.Model(&model.DockerImage{}).Where(&query)

	// 查询总记录数
	if err := db.Count(&page.Count).Error; err != nil {
		return page, err
	}

	// 计算最大页码
	if page.Count > 0 {
		page.MaxPage = int64(math.Ceil(float64(page.Count) / float64(page.PageSize)))
	}

	// 查询记录
	records := make([]model.DockerImage, 0)
	if isPage {
		// 分页查询
		if err := db.Offset(offset).
			Limit(pageSize).
			Find(&records).Error; err != nil {
			return page, err
		}
	} else {
		// 不分页查询
		if err := db.Find(&records).Error; err != nil {
			return page, err
		}
	}

	page.Records = records
	return page, nil

}
