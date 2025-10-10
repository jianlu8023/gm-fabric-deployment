package model

import (
	"database/sql"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

const (
	userInfoTableName = "t_user_info"
)

var columns = defaultColumns()

func defaultColumns() *atomic.Value {
	v := &atomic.Value{}
	taskType := reflect.TypeOf(NewUserInfo())
	m := make(map[string]string)
	for i := 0; i < taskType.NumField(); i++ {
		field := taskType.Field(i)
		gormTag := field.Tag.Get("gorm")
		split := strings.Split(gormTag, ";")
		for j := 0; j < len(split); j++ {
			if strings.Contains(split[i], "column") {
				columnSplit := strings.Split(split[i], ":")
				if len(columnSplit) == 2 {
					m[field.Name] = columnSplit[1]
				}
			}
		}
	}
	v.Store(m)
	return v
}

// UserInfo 用户信息模型
// @description 定义系统用户的数据结构，包含用户的基本信息
// @struct
type UserInfo struct {
	AutoUid           int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                                   // 自增ID（主键）
	UserId            string       `json:"user_id,omitempty" yaml:"user_id,omitempty" gorm:"column:user_id;type:varchar(255);not null;"`                                      // 用户id
	Username          string       `json:"username,omitempty" yaml:"username,omitempty" gorm:"column:username;type:varchar(255);not null;unique"`                             // 用户名（唯一）
	Password          string       `json:"password,omitempty" yaml:"password,omitempty" gorm:"column:password;type:varchar(255);not null"`                                    // 密码（JSON序列化时忽略）
	Email             string       `json:"email,omitempty" yaml:"email,omitempty" gorm:"column:email;type:varchar(255);not null;unique"`                                      // 邮箱（唯一）
	UserType          string       `json:"user_type,omitempty" yaml:"user_type,omitempty" gorm:"column:user_type;type:varchar(255);"`                                         // 用户类型
	Certificate       string       `json:"certificate,omitempty" yaml:"certificate,omitempty" gorm:"column:certificate;type:text;default:''"`                                 // 证书（JSON序列化时忽略）
	MFAEnabled        sql.NullBool `json:"mfa_enabled,omitempty" yaml:"mfa_enabled,omitempty" gorm:"column:mfa_enabled;type:tinyint(1);default:0;"`                           // 是否启用mfa
	MFASecret         string       `json:"mfa_secret,omitempty" yaml:"mfa_secret,omitempty" gorm:"column:mfa_secret;type:varchar(255);default:'';"`                           // mfa的secret
	MFARecoverySecret string       `json:"mfa_recovery_secret,omitempty" yaml:"mfa_recovery_secret,omitempty" gorm:"column:mfa_recovery_secret;type:varchar(255);default:''"` // 恢复secret
	MFAOTPAuthURL     string       `json:"mfa_otp_auth_url,omitempty" yaml:"mfa_otp_auth_url,omitempty" gorm:"column:mfa_otp_auth_url;type:varchar(255);default:'';"`         // mfa的otp auth url
	IsDelete          sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:tinyint(1);default:0"`                                  // 是否删除标记（默认为0）
	LastLoginTime     time.Time    `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty" gorm:"column:last_login_time;type:datetime;default:CURRENT_TIMESTAMP"`  // 最后登录时间（默认为当前时间戳）
}

func (model UserInfo) TableColumns() map[string]string {
	return columns.Load().(map[string]string)
}

// String 将用户信息转换为字符串表示
// @description 将UserInfo结构体转换为JSON格式的字符串
// @return string 用户信息的JSON格式字符串
func (model UserInfo) String() string {
	str, _ := json.MarshalString(model)
	return str
}

// TableName 返回数据库表名
// @description 实现gorm接口，指定UserInfo结构体对应的数据库表名
// @return string 数据库表名
func (UserInfo) TableName() string {
	return userInfoTableName
}

// NewUserInfo 创建新的用户信息实例
// @description 初始化一个空的用户信息结构体指针
// @return *UserInfo 用户信息结构体指针
func NewUserInfo() *UserInfo {
	return &UserInfo{}
}
