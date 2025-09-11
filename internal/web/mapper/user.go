package mapper

import (
	"errors"

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

// QueryUserByUsernameAndPassword 根据用户名和密码查询用户
func (m *UserMapper) QueryUserByUsernameAndPassword(username, password string) (*model.UserInfo, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}

	user := &model.UserInfo{}
	if err := m.db.Where("username = ? AND password = ? AND is_delete = 0", username, password).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在或密码错误
			return nil, errors.New("用户不存在或密码错误")
		}
		// 其他数据库错误
		return nil, err
	}

	return user, nil
}

func NewUserMapper(baseMapper *Mapper) *UserMapper {
	return &UserMapper{
		Mapper: baseMapper,
	}
}
