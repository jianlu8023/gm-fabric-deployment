package model

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
)

const (
	userInfoTableName = "t_user_info"
)

type UserInfo struct {
	AutoUid       int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                  // 自增ID
	Username      string       `json:"username,omitempty" yaml:"username,omitempty" gorm:"column:username;type:varchar(255);not null;unique"`                            // 用户名
	Password      string       `json:"-" yaml:"password,omitempty" gorm:"column:password;type:varchar(255);not null"`                                                    // 密码
	Email         string       `json:"email,omitempty" yaml:"email,omitempty" gorm:"column:email;type:varchar(255);not null;unique"`                                     // 邮箱
	IsDelete      sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:tinyint(1);default:0"`                                 // 是否删除
	LastLoginTime time.Time    `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty" gorm:"column:last_login_time;type:datetime;default:CURRENT_TIMESTAMP"` // 最后登录时间
}

// MarshalJSON 自定义json序列化
func (u UserInfo) MarshalJSON() ([]byte, error) {
	type Alias UserInfo
	aux := struct {
		*Alias
		IsDelete bool `json:"is_delete" yaml:"is_delete"`
	}{
		Alias:    (*Alias)(&u),
		IsDelete: u.IsDelete.Bool,
	}
	return json.Marshal(aux)
}

func (u UserInfo) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

func (u UserInfo) TableName() string {
	return userInfoTableName
}

func NewUserInfo() *UserInfo {
	return &UserInfo{}
}
