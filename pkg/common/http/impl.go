package http

import (
	"github.com/gin-gonic/gin"
)

type MyRouter struct {
	Name            string                 `json:"name,omitempty" yaml:"router_name,omitempty"`
	Uri             string                 `json:"uri,omitempty" yaml:"router_uri,omitempty"`
	Method          string                 `json:"method,omitempty" yaml:"router_method,omitempty"`
	HandlerFunc     func(ctx *gin.Context) `json:"-" yaml:"-"`
	Enabled         bool                   `json:"enabled,omitempty" yaml:"router_enabled,omitempty"`
	Desc            string                 `json:"desc,omitempty" yaml:"router_desc,omitempty"`
	EnableJWtVerify bool                   `json:"enable_jwt_verify,omitempty" yaml:"router_enable_jwt_verify,omitempty"`
}

// GetName 获取路由名称
func (r *MyRouter) GetName() string {
	return r.Name
}

// GetUri 获取路由URI
func (r *MyRouter) GetUri() string {
	return r.Uri
}

// GetMethod 获取HTTP方法
func (r *MyRouter) GetMethod() string {
	return r.Method
}

// GetHandlerFunc 获取处理函数
func (r *MyRouter) GetHandlerFunc() gin.HandlerFunc {
	return r.HandlerFunc
}

// GetEnableJWtVerify 获取是否启用JWT验证
func (r *MyRouter) GetEnableJWtVerify() bool {
	return r.EnableJWtVerify
}

// IsEnabled 检查路由是否启用
func (r *MyRouter) IsEnabled() bool {
	return r.Enabled
}

// GetDesc 获取路由描述
func (r *MyRouter) GetDesc() string {
	return r.Desc
}
