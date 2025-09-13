package http

import (
	"github.com/gin-gonic/gin"
)

type MyRouter struct {
	Name            string                 `yaml:"name" json:"router_name"`
	Uri             string                 `yaml:"uri" json:"router_uri"`
	Method          string                 `yaml:"method" json:"router_method"`
	HandlerFunc     func(ctx *gin.Context) `yaml:"-" json:"-"`
	Enabled         bool                   `yaml:"enabled" json:"router_enabled"`
	Desc            string                 `yaml:"desc" json:"router_desc"`
	EnableJWtVerify bool                   `yaml:"enable_jwt_verify" json:"router_enable_jwt_verify"`
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
