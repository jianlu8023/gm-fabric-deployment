package mapper

import (
	"database/sql"
	"errors"

	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"gorm.io/gorm"
)

// UserMapper 用户数据访问层结构体
//
// @description 提供用户相关的数据访问操作
// @struct
type UserMapper struct {
	*Mapper
}

// NewUserMapper 创建一个新的UserMapper实例
//
// @param baseMapper *Mapper 基础Mapper
// @return *UserMapper UserMapper实例
func NewUserMapper(baseMapper *Mapper) *UserMapper {
	return &UserMapper{
		Mapper: baseMapper,
	}
}

// QueryExistUser 查询用户是否存在
//
// @param query model.UserInfo 查询条件
// @return error 错误信息，如果用户存在返回ErrAlreadyExists
func (m *UserMapper) QueryExistUser(query model.UserInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	var exist int64
	if err := m.db.Model(&model.UserInfo{}).
		Where(&query).
		Count(&exist).Error; err != nil {
		return err
	}
	if exist > 0 {
		return datasource.ErrAlreadyExists
	}
	return nil
}

// InsertOneUser 插入一个用户
//
// @param user *model.UserInfo 用户信息
// @return error 错误信息
func (m *UserMapper) InsertOneUser(user *model.UserInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserInfo{}).
			Create(user).Error; err != nil {
			return err
		}
		return nil
	})
}

// QueryUserByUsernameAndPassword 根据用户名和密码查询用户
//
// @param username string 用户名
// @param password string 密码
// @return *model.UserInfo 用户信息
// @return error 错误信息，如果用户不存在或密码错误返回自定义错误
func (m *UserMapper) QueryUserByUsernameAndPassword(
	username,
	password string,
) (*model.UserInfo, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}

	user := &model.UserInfo{}
	if err := m.db.Where(
		&model.UserInfo{
			Username: username,
			Password: password,
			IsDelete: sql.NullBool{
				Bool:  false,
				Valid: true,
			},
		},
	).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在或密码错误
			return nil, errors.New("用户不存在或密码错误")
		}
		// 其他数据库错误
		return nil, err
	}

	return user, nil
}
