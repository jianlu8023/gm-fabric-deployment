package model

import (
	"database/sql"
	"time"

	humantime "github.com/jianlu8023/golang-example/pkg/human/time"
	"github.com/jianlu8023/golang-example/pkg/json"
)

const (
	userInfoTableName = "t_user_info"
)

// UserInfo 用户信息模型
// @description 定义系统用户的数据结构，包含用户的基本信息
// @struct
type UserInfo struct {
	AutoUid       int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                  // 自增ID（主键）
	Username      string       `json:"username,omitempty" yaml:"username,omitempty" gorm:"column:username;type:varchar(255);not null;unique"`                            // 用户名（唯一）
	Password      string       `json:"-" yaml:"password,omitempty" gorm:"column:password;type:varchar(255);not null"`                                                    // 密码（JSON序列化时忽略）
	Email         string       `json:"email,omitempty" yaml:"email,omitempty" gorm:"column:email;type:varchar(255);not null;unique"`                                     // 邮箱（唯一）
	UserType      string       `json:"user_type,omitempty" yaml:"user_type,omitempty" gorm:"column:user_type;type:varchar(255);"`                                        // 用户类型
	Certificate   string       `json:"-" yaml:"certificate,omitempty" gorm:"column:certificate;type:text"`                                                               // 证书（JSON序列化时忽略）
	IsDelete      sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:tinyint(1);default:0"`                                 // 是否删除标记（默认为0）
	LastLoginTime time.Time    `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty" gorm:"column:last_login_time;type:datetime;default:CURRENT_TIMESTAMP"` // 最后登录时间（默认为当前时间戳）
}

// MarshalJSON 自定义JSON序列化方法
// @description 自定义UserInfo结构体的JSON序列化逻辑，将sql.NullBool类型的IsDelete字段转换为普通bool类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (u *UserInfo) MarshalJSON() ([]byte, error) {
	type Alias UserInfo
	aux := struct {
		*Alias
		IsDelete      bool   `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
		LastLoginTime string `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty"`
	}{
		Alias:         (*Alias)(u),
		IsDelete:      u.IsDelete.Bool,
		LastLoginTime: humantime.HumanTimeLower(u.LastLoginTime, "unknown"),
	}
	return json.Marshal(aux)
}

// String 将用户信息转换为字符串表示
// @description 将UserInfo结构体转换为JSON格式的字符串
// @return string 用户信息的JSON格式字符串
func (u *UserInfo) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

// TableName 返回数据库表名
// @description 实现gorm接口，指定UserInfo结构体对应的数据库表名
// @return string 数据库表名
func (u *UserInfo) TableName() string {
	return userInfoTableName
}

// NewUserInfo 创建新的用户信息实例
// @description 初始化一个空的用户信息结构体指针
// @return *UserInfo 用户信息结构体指针
func NewUserInfo() *UserInfo {
	return &UserInfo{}
}
