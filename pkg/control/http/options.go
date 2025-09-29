package http

import (
	"context"
	
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
)

// Option 定义用于配置 Control 的函数类型
// @description 用于通过函数式选项模式配置 Control 结构体
// @param control *Control 要配置的 Control 实例

type Option func(control *Control)

// WithTracer 设置 tracer 控制器
// @description 设置 HTTP 服务器的追踪控制器
// @param tracerControl *tracer.Control 追踪控制器实例
// @return Option 配置函数
func WithTracer(tracerControl *tracer.Control) Option {
	return func(control *Control) {
		control.tracerControl = tracerControl
	}
}

// WithContext 设置上下文
// @description 设置 HTTP 服务器的上下文
// @param ctx context.Context 上下文实例
// @return Option 配置函数
func WithContext(ctx context.Context) Option {
	return func(control *Control) {
		control.ctx = ctx
	}
}

// WithSessionManager 设置会话管理器
// @description 设置 HTTP 服务器的会话管理器
// @param sessionManager jwt.SessionManager 会话管理器实例
// @return Option 配置函数
func WithSessionManager(sessionManager jwt.SessionManager) Option {
	return func(control *Control) {
		control.sessionManager = sessionManager
	}
}
