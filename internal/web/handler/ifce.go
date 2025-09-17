package handler

import (
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// RoutersInterface 路由接口
//
// @description 定义所有处理器需要实现的路由接口
// @interface
type RoutersInterface interface {
	// Routers 获取路由列表
	//
	// @return []commonhttp.RouterHandler 路由处理器列表
	Routers() []commonhttp.RouterHandler
}
