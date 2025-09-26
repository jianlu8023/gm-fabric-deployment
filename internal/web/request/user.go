package request

import (
	// validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
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

func (req UserRegisterRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

func (req UserRegisterRequest) IsLegal() bool {
	if stringer.IsBlank(req.Username) ||
		stringer.IsBlank(req.Password) ||
		stringer.IsBlank(req.Email) {
		return false
	}
	return true
}

// UserLoginRequest 登录请求结构体
type UserLoginRequest struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" form:"username" binding:"required"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" form:"password" binding:"required"`
}

func (req UserLoginRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

func (req UserLoginRequest) IsLegal() bool {
	if stringer.IsBlank(req.Username) ||
		stringer.IsBlank(req.Password) {
		return false
	}
	return true
}
