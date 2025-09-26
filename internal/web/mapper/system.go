package mapper

import (
	"database/sql"

	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"gorm.io/gorm"
)

// SystemMapper 系统数据访问层结构体
//
// @description 提供系统相关的数据访问操作
// @struct
type SystemMapper struct {
	*Mapper
}

func (m *SystemMapper) GetSystemInit() (bool, error) {
	if m.db == nil {
		return false, datasource.ErrNoDataSourceConn
	}
	m.init()
	var systemInit model.SystemInit
	if err := m.db.Model(&model.SystemInit{}).Where(&model.SystemInit{
		Id: 1,
	}).First(&systemInit).Error; err != nil {
		return false, err
	}

	return systemInit.IsInit.Bool, nil
}

func (m *SystemMapper) init() {
	if m.db == nil {
		return
	}
	if err := m.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.SystemInit{}).Where(&model.SystemInit{
			Id: 1,
		}).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := tx.Model(&model.SystemInit{}).Create(&model.SystemInit{
				Id: 1,
				IsInit: sql.NullBool{
					Bool:  false,
					Valid: true,
				},
			}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		m.logger.Errorf("init system status failed: %s", err)
		return
	}
}

// NewSystemMapper 创建一个新的SystemMapper实例
//
// @param mapper *Mapper 基础Mapper
// @return *SystemMapper SystemMapper实例
func NewSystemMapper(mapper *Mapper) *SystemMapper {

	m := &SystemMapper{Mapper: mapper}

	// 调用init方法
	// m.init()

	return m
}
