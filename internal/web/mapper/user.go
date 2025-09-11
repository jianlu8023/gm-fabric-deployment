package mapper

import (
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"gorm.io/gorm"
)

type UserMapper struct {
	*Mapper
}

func (m *UserMapper) QueryExistUser(query model.UserInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	var exist int64
	if err := m.db.Model(&model.UserInfo{}).Where(&query).Count(&exist).Error; err != nil {
		return err
	}
	if exist > 0 {
		return datasource.ErrAlreadyExists
	}
	return nil
}

func (m *UserMapper) InsertOneUser(user *model.UserInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserInfo{}).Create(user).Error; err != nil {
			return err
		}
		return nil
	})
}

func NewUserMapper(baseMapper *Mapper) *UserMapper {
	return &UserMapper{
		Mapper: baseMapper,
	}
}
