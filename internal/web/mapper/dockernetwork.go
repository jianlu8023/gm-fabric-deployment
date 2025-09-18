package mapper

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"gorm.io/gorm"
	"math"
)

type DockerNetworkMapper struct {
	*Mapper
}

// NewDockerNetworkMapper 新建docker network mapper
func NewDockerNetworkMapper(baseMapper *Mapper) *DockerNetworkMapper {
	return &DockerNetworkMapper{
		Mapper: baseMapper,
	}
}

// InsertOrUpdateOne 插入或更新Docker网络信息，支持恢复已逻辑删除的记录
// @description 在事务中插入或更新Docker网络信息，根据网络ID和位置确定是否存在
// @param info *Info 要插入或更新的网络信息
// @return error 操作结果错误信息
func (m *DockerNetworkMapper) InsertOrUpdateOne(info *model.DockerNetwork) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		var existInfo model.DockerNetwork
		if err := tx.Model(&model.DockerNetwork{}).
			Where(
				&model.DockerNetwork{
					NetworkID:             info.NetworkID,
					NetworkLocationPeerId: info.NetworkLocationPeerId,
					// 不限制IsDelete，允许找到已逻辑删除的记录
				},
			).First(&existInfo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 记录不存在，创建新记录
				// 确保新记录的IsDelete字段为false
				info.IsDelete = sql.NullBool{Bool: false, Valid: true}
				if err := tx.Model(&model.DockerNetwork{}).
					Create(info).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 记录存在，更新记录
			// 恢复已逻辑删除的记录（确保IsDelete为false）
			info.IsDelete = sql.NullBool{Bool: false, Valid: true}
			info.AutoUid = existInfo.AutoUid
			if err := tx.Model(&model.DockerNetwork{}).
				Where(
					&model.DockerNetwork{
						AutoUid: info.AutoUid,
					},
				).Updates(info).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchLogicalDelete 批量逻辑删除Docker网络信息
// @description 根据查询条件批量将Docker网络标记为已删除
// @param query model.DockerNetwork 查询条件
// @return error 操作结果错误信息
func (m *DockerNetworkMapper) BatchLogicalDelete(query model.DockerNetwork) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}

	// 开始事务
	return m.db.Transaction(func(tx *gorm.DB) error {
		// 构建查询
		db := tx.Model(&model.DockerNetwork{}).Where(&query)

		var count int64
		if err := db.Count(&count).Error; err != nil {
			return fmt.Errorf("查询Docker镜像记录数失败: %w", err)
		}

		// 如果记录数为0，则返回nil
		if count == 0 {
			return nil
		}

		// 执行逻辑删除，设置IsDelete为true
		result := db.Updates(&model.DockerNetwork{
			IsDelete: sql.NullBool{Bool: true, Valid: true},
		})
		if result.Error != nil {
			return fmt.Errorf("批量逻辑删除Docker网络失败: %w", result.Error)
		}

		// 检查是否有记录被删除
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func (m *DockerNetworkMapper) DockerNetworkList(query model.DockerNetwork,
	isPage bool, pageNo int, pageSize int,
) (dbpage.Info[model.DockerNetwork], error) {
	page := dbpage.Info[model.DockerNetwork]{}
	if m.db == nil {
		return page, datasource.ErrNoDataSourceConn
	}
	page.PageNo = int64(pageNo)
	page.PageSize = int64(pageSize)

	// 计算偏移量
	offset := (pageNo - 1) * pageSize

	// 构建查询
	db := m.db.Model(&model.DockerNetwork{}).Where(&query)

	// 查询总记录数
	if err := db.Count(&page.Count).Error; err != nil {
		return page, err
	}

	// 计算最大页码
	if page.Count > 0 {
		page.MaxPage = int64(math.Ceil(float64(page.Count) /
			float64(page.PageSize)))
	}

	// 查询记录
	records := make([]model.DockerNetwork, 0)
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
