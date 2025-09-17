package request

import (
	// validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jianlu8023/golang-example/pkg/json"
	"github.com/jianlu8023/golang-example/pkg/str"
)

type UserRegisterRequest struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" form:"username" binding:"required"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" form:"password" binding:"required"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty" form:"email" binding:"required,email"`
}

// func (u UserRegisterRequest) Valid() error {
// 	err := validation.ValidateStruct(u,
// 		validation.Field(&u.Username, validation.Required, validation.Nil),
// 	)
// 	return err
// }

func (u *UserRegisterRequest) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

func (u *UserRegisterRequest) IsLegal() bool {
	if str.IsBlank(u.Username) ||
		str.IsBlank(u.Password) ||
		str.IsBlank(u.Email) {
		return false
	}
	return true
}

// UserLoginRequest 登录请求结构体
type UserLoginRequest struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" form:"username" binding:"required"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" form:"password" binding:"required"`
}

func (u *UserLoginRequest) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

func (u *UserLoginRequest) IsLegal() bool {
	if str.IsBlank(u.Username) ||
		str.IsBlank(u.Password) {
		return false
	}
	return true
}
