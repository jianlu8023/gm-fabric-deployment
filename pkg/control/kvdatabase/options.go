package kvdatabase

import (
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
)

// Option 配置选项
type Option func(*Control)

// WithTracerControl 设置追踪控制器
func WithTracerControl(tracerControl *tracer.Control) Option {
	return func(c *Control) {
		c.tracerControl = tracerControl
	}
}
