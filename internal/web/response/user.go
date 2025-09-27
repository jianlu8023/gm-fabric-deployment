package response

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jinzhu/copier"
)

// RegisterUserResponse 用户注册响应
// @description 用户注册成功后的响应数据结构
// @struct
// @property Username string 用户名（唯一）
// @property Email string 邮箱（唯一）
// @property UserType string 用户类型
// @property MFAEnabled sql.NullBool 是否启用MFA
// @property IsDelete sql.NullBool 是否删除标记
// @property LastLoginTime time.Time 最后登录时间
type RegisterUserResponse struct {
	Username      string       `json:"username" yaml:"username"`               // 用户名（唯一）
	Email         string       `json:"email" yaml:"email"`                     // 邮箱（唯一）
	UserType      string       `json:"user_type" yaml:"user_type"`             // 用户类型
	MFAEnabled    sql.NullBool `json:"mfa_enabled" yaml:"mfa_enabled"`         // 是否启用mfa
	IsDelete      sql.NullBool `json:"is_delete" yaml:"is_delete"`             // 是否删除标记（默认为0）
	LastLoginTime time.Time    `json:"last_login_time" yaml:"last_login_time"` // 最后登录时间（默认为当前时间戳）
}

// String 将RegisterUserResponse转换为字符串
// @description 将RegisterUserResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp RegisterUserResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义JSON序列化方法
// @description 自定义RegisterUserResponse结构体的JSON序列化逻辑，将sql.NullBool类型的IsDelete字段转换为普通bool类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp RegisterUserResponse) MarshalJSON() ([]byte, error) {
	type Alias RegisterUserResponse
	aux := struct {
		*Alias
		IsDelete      bool   `json:"is_delete"`
		LastLoginTime string `json:"last_login_time"`
	}{
		Alias:         (*Alias)(&resp),
		IsDelete:      resp.IsDelete.Bool,
		LastLoginTime: humantime.HumanTimeLower(resp.LastLoginTime, "unknown"),
	}
	return json.Marshal(aux)
}

// NewRegisterUserResponse 创建用户注册响应对象
// @description 创建一个新的用户注册响应对象
// @param user *model.UserInfo 用户信息模型
// @return RegisterUserResponse 用户注册响应
// @return error 错误信息
func NewRegisterUserResponse(user *model.UserInfo) (RegisterUserResponse, error) {
	var convert RegisterUserResponse
	if err := copier.CopyWithOption(&convert, user, copier.Option{
		DeepCopy:    true,
		IgnoreEmpty: true,
	}); err != nil {
		return RegisterUserResponse{}, err
	}
	return convert, nil
}

// LoginUserResponse 用户登录响应
// @description 用户登录成功后的响应数据结构
// @struct
// @property Username string 用户名
// @property UserType string 用户类型
// @property Email string 邮箱
// @property Token string 认证令牌
// @property LastLoginTime time.Time 最后登录时间
type LoginUserResponse struct {
	Username      string    `json:"username" yaml:"username"`               // 用户名
	UserType      string    `json:"user_type" yaml:"user_type"`             // 用户类型
	Email         string    `json:"email" yaml:"email"`                     // 邮箱
	Token         string    `json:"token" yaml:"token"`                     // 认证令牌
	LastLoginTime time.Time `json:"last_login_time" yaml:"last_login_time"` // 最后登录时间（默认为当前时间戳）
}

// String 将LoginUserResponse转换为字符串
// @description 将LoginUserResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp LoginUserResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义JSON序列化方法
// @description 自定义LoginUserResponse结构体的JSON序列化逻辑，处理time.Time类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp LoginUserResponse) MarshalJSON() ([]byte, error) {
	type Alias LoginUserResponse
	aux := struct {
		*Alias
		LastLoginTime string `json:"last_login_time"`
	}{
		Alias:         (*Alias)(&resp),
		LastLoginTime: humantime.HumanTimeLower(resp.LastLoginTime, "unknown"),
	}
	return json.Marshal(aux)
}

// NewLoginUserResponse 创建用户登录响应对象
// @description 创建一个新的用户登录响应对象
// @param user *model.UserInfo 用户信息模型
// @param token string 认证令牌
// @return LoginUserResponse 用户登录响应
func NewLoginUserResponse(user *model.UserInfo, token string) LoginUserResponse {
	return LoginUserResponse{
		Username:      user.Username,
		UserType:      user.UserType,
		Email:         user.Email,
		Token:         token,
		LastLoginTime: user.LastLoginTime,
	}
}
