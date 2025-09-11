package request

import (
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/str"
)

type UserRegisterRequest struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty" form:"username" binding:"required"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" form:"password" binding:"required"`
	Email    string `json:"email,omitempty" yaml:"email,omitempty" form:"email" binding:"required,email"`
}

func (u UserRegisterRequest) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

func (u UserRegisterRequest) IsLegal() bool {
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

func (u UserLoginRequest) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

func (u UserLoginRequest) IsLegal() bool {
	if str.IsBlank(u.Username) ||
		str.IsBlank(u.Password) {
		return false
	}
	return true
}
