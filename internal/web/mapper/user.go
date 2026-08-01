package mapper

import (
	"errors"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/sqlnull"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"gorm.io/gorm"
)

// UserMapper 用户数据访问接口
//
// @description 定义用户相关的数据访问方法，包括用户的增删改查及登录信息维护
// @interface
type UserMapper interface {
	// UpdateUser 更新用户信息
	//
	// @description 根据UserId更新未删除用户的信息
	// @param user *model.UserInfo 包含要更新的用户信息
	// @return error 错误信息
	UpdateUser(user *model.UserInfo) error
	// QueryExistUser 查询用户是否存在
	//
	// @description 根据查询条件检查未删除的用户是否存在
	// @param query model.UserInfo 查询条件
	// @return bool 用户是否存在
	// @return error 错误信息
	QueryExistUser(query model.UserInfo) (bool, error)
	// InsertOneUser 插入一个用户
	//
	// @description 插入一条用户信息到数据库
	// @param user *model.UserInfo 用户信息
	// @return error 错误信息
	InsertOneUser(user *model.UserInfo) error
	// QueryUserByUsernameAndPassword 根据用户名和密码查询用户
	//
	// @description 根据用户名和密码查询未删除的用户信息
	// @param username string 用户名
	// @param password string 密码
	// @return *model.UserInfo 用户信息
	// @return error 错误信息，如果用户不存在或密码错误返回自定义错误
	QueryUserByUsernameAndPassword(username, password string) (*model.UserInfo, error)
	// UpdateLastLoginTime 更新用户最后登录时间
	//
	// @description 更新指定用户的最后登录时间为当前时间
	// @param user *model.UserInfo 用户信息
	// @return error 错误信息
	UpdateLastLoginTime(user *model.UserInfo) error
	// QueryUserByQuery 根据查询条件查询用户信息
	//
	// @description 根据传入的查询条件，查询未被删除的用户信息
	// @param query model.UserInfo 查询条件
	// @return model.UserInfo 用户信息
	// @return error 错误信息
	QueryUserByQuery(query model.UserInfo) (model.UserInfo, error)
}

// userMapperImpl 用户数据访问层结构体
//
// @description 提供用户相关的数据访问操作
// @struct
type userMapperImpl struct {
	*Mapper
}

// NewUserMapper 创建一个新的UserMapper实例
//
// @description 创建并返回一个新的UserMapper实例，用于用户相关的数据访问操作
// @param baseMapper *Mapper 基础Mapper
// @return UserMapper UserMapper实例
func NewUserMapper(baseMapper *Mapper) UserMapper {
	return &userMapperImpl{
		Mapper: baseMapper,
	}
}

// QueryExistUser 查询用户是否存在
//
// @description 查询指定条件的用户是否存在
// @param query model.UserInfo 查询条件
// @return bool 用户是否存在
// @return error 错误信息
func (m *userMapperImpl) QueryExistUser(query model.UserInfo) (bool, error) {
	if m.db == nil {
		return false, datasource.ErrNoDataSourceConn
	}
	var exist int64
	if err := m.db.Model(&model.UserInfo{}).
		Where(&query).
		Where(&model.UserInfo{
			IsDelete: sqlnull.FalseToNull(),
		}).
		Count(&exist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	// 用户存在返回true
	if exist > 0 {
		return true, nil
	}
	// 用户不存在返回false
	return false, nil
}

// InsertOneUser 插入一个用户
//
// @description 插入一条用户信息到数据库
// @param user *model.UserInfo 用户信息
// @return error 错误信息
func (m *userMapperImpl) InsertOneUser(user *model.UserInfo) error {
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
// @description 根据用户名和密码查询用户信息
// @param username string 用户名
// @param password string 密码
// @return *model.UserInfo 用户信息
// @return error 错误信息，如果用户不存在或密码错误返回自定义错误
func (m *userMapperImpl) QueryUserByUsernameAndPassword(
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
			IsDelete: sqlnull.FalseToNull(),
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

// UpdateLastLoginTime 更新用户最后登录时间
//
// @description 更新指定用户的最后登录时间为当前时间
// @param user *model.UserInfo 用户信息
// @return error 错误信息
func (m *userMapperImpl) UpdateLastLoginTime(user *model.UserInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserInfo{}).
			Where(&model.UserInfo{
				AutoUid:  user.AutoUid,
				Username: user.Username,
				Email:    user.Email,
			}).Updates(&model.UserInfo{
			LastLoginTime: time.Now(),
		}).Error; err != nil {
			return err
		}
		return nil
	})

}

// QueryUserByQuery 根据查询条件查询用户信息
//
// @description 根据传入的查询条件，查询未被删除的用户信息
// @param query model.UserInfo 查询条件
// @return model.UserInfo 用户信息
// @return error 错误信息
func (m *userMapperImpl) QueryUserByQuery(query model.UserInfo) (model.UserInfo, error) {
	if m.db == nil {
		return model.UserInfo{}, datasource.ErrNoDataSourceConn
	}
	var user model.UserInfo
	if err := m.db.Model(&model.UserInfo{}).
		Where(&query).Where(&model.UserInfo{
		IsDelete: sqlnull.FalseToNull(),
	}).First(&user).Error; err != nil {
		return model.UserInfo{}, err
	}
	return user, nil
}

// UpdateUser 更新用户信息
//
// @description 更新指定用户的信息，根据UserId字段确定要更新的用户
// @param user *model.UserInfo 包含要更新的用户信息
// @return error 错误信息
func (m *userMapperImpl) UpdateUser(user *model.UserInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}
	return m.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.UserInfo{}).Where(&model.UserInfo{
			IsDelete: sqlnull.FalseToNull(),
			UserId:   user.UserId,
		}).Updates(user).Error; err != nil {
			return err
		}
		return nil
	})
}
