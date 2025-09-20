package http

import (
	"github.com/gin-gonic/gin"
)

// RouterHandler 定义HTTP路由处理器接口
// 用于打破pkg和internal之间的循环依赖
// @interface RouterHandler
// @since 1.0.0
// @description 定义HTTP路由的基本行为
// @property Name string 路由名称
// @property Uri string 路由URI
// @property Method string HTTP方法
// @property HandlerFunc gin.HandlerFunc 处理函数
// @property Enabled bool 是否启用
// @property Desc string 路由描述
// @returns RouterHandler 路由处理器接口
// @example
// var router http.RouterHandler
// router = &MyRouter{}
// fmt.Println(router.GetUri())
type RouterHandler interface {
	// GetName 获取路由名称
	GetName() string
	// GetUri 获取路由URI
	GetUri() string
	// GetMethod 获取HTTP方法
	GetMethod() string
	// GetHandlerFunc 获取处理函数
	GetHandlerFunc() gin.HandlerFunc
	// IsEnabled 检查路由是否启用
	IsEnabled() bool
	// GetDesc 获取路由描述
	GetDesc() string
	// GetEnableJWtVerify 获取是否启用jwt验证
	GetEnableJWtVerify() bool
}
