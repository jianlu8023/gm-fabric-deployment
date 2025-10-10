package request

import (
	// validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// UserRegisterRequest 用户注册请求结构体
//
// @description 用户注册时提交的请求参数
// @struct
type UserRegisterRequest struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" form:"username" binding:"required"` // 用户名 (必需)
	Password string `json:"password,omitempty" yaml:"password,omitempty" form:"password" binding:"required"` // 密码 (必需)
	Email    string `json:"email,omitempty" yaml:"email,omitempty" form:"email" binding:"required,email"`    // 邮箱 (必需，需符合邮箱格式)
}

// func (u UserRegisterRequest) Valid() error {
//	err := validation.ValidateStruct(u,
//		validation.Field(&u.Username, validation.Required, validation.Nil),
//	)
//	return err
// }

// String 将请求参数转换为字符串表示
//
// @description 将UserRegisterRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req UserRegisterRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证请求参数是否合法
//
// @description 检查用户名、密码和邮箱是否为空
// @return bool 参数是否合法
func (req UserRegisterRequest) IsLegal() bool {
	if stringer.IsBlank(req.Username) ||
		stringer.IsBlank(req.Password) ||
		stringer.IsBlank(req.Email) {
		return false
	}
	return true
}

// UserLoginRequest 登录请求结构体
//
// @description 用户登录时提交的请求参数
// @struct
type UserLoginRequest struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" form:"username" binding:"required"` // 用户名 (必需)
	Password string `json:"password,omitempty" yaml:"password,omitempty" form:"password" binding:"required"` // 密码 (必需)
}

// String 将登录请求参数转换为字符串表示
//
// @description 将UserLoginRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req UserLoginRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证登录请求参数是否合法
//
// @description 检查用户名和密码是否为空
// @return bool 参数是否合法
func (req UserLoginRequest) IsLegal() bool {
	if stringer.IsBlank(req.Username) ||
		stringer.IsBlank(req.Password) {
		return false
	}
	return true
}
