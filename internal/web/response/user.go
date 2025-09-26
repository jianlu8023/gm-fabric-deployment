package response

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jinzhu/copier"
)

type RegisterUserResponse struct {
	Username      string       `json:"username,omitempty" yaml:"username,omitempty" gorm:"column:username;type:varchar(255);not null;unique"`                            // 用户名（唯一）
	Email         string       `json:"email,omitempty" yaml:"email,omitempty" gorm:"column:email;type:varchar(255);not null;unique"`                                     // 邮箱（唯一）
	UserType      string       `json:"user_type,omitempty" yaml:"user_type,omitempty" gorm:"column:user_type;type:varchar(255);"`                                        // 用户类型
	MFAEnabled    sql.NullBool `json:"mfa_enabled,omitempty" yaml:"mfa_enabled,omitempty" gorm:"column:mfa_enabled;type:tinyint(1);default:0;"`                          // 是否启用mfa
	IsDelete      sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:tinyint(1);default:0"`                                 // 是否删除标记（默认为0）
	LastLoginTime time.Time    `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty" gorm:"column:last_login_time;type:datetime;default:CURRENT_TIMESTAMP"` // 最后登录时间（默认为当前时间戳）
}

func (resp RegisterUserResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义JSON序列化方法
// @description 自定义UserInfo结构体的JSON序列化逻辑，将sql.NullBool类型的IsDelete字段转换为普通bool类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp RegisterUserResponse) MarshalJSON() ([]byte, error) {
	type Alias RegisterUserResponse
	aux := struct {
		*Alias
		IsDelete      bool   `json:"is_delete,omitempty" yaml:"is_delete,omitempty"`
		LastLoginTime string `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty"`
	}{
		Alias:         (*Alias)(&resp),
		IsDelete:      resp.IsDelete.Bool,
		LastLoginTime: humantime.HumanTimeLower(resp.LastLoginTime, "unknown"),
	}
	return json.Marshal(aux)
}

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

type LoginUserResponse struct {
	Username      string    `json:"username,omitempty" yaml:"username,omitempty"`
	UserType      string    `json:"user_type,omitempty" yaml:"user_type,omitempty"` // 用户类型
	Email         string    `json:"email,omitempty" yaml:"email,omitempty"`
	Token         string    `json:"token,omitempty" yaml:"token,omitempty"`
	LastLoginTime time.Time `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty"` // 最后登录时间（默认为当前时间戳）
}

func (resp LoginUserResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func (resp LoginUserResponse) MarshalJSON() ([]byte, error) {
	type Alias LoginUserResponse
	aux := struct {
		*Alias
		LastLoginTime string `json:"last_login_time,omitempty" yaml:"last_login_time,omitempty"`
	}{
		Alias:         (*Alias)(&resp),
		LastLoginTime: humantime.HumanTimeLower(resp.LastLoginTime, "unknown"),
	}
	return json.Marshal(aux)
}

func NewLoginUserResponse(user *model.UserInfo, token string) LoginUserResponse {
	return LoginUserResponse{
		Username:      user.Username,
		UserType:      user.UserType,
		Email:         user.Email,
		Token:         token,
		LastLoginTime: user.LastLoginTime,
	}
}
