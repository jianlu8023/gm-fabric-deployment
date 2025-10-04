package grpc

import (
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
)

type Option func(control *Control)

func WithTracer(tracerControl *tracer.Control) Option {
	return func(control *Control) {
		control.tracerControl = tracerControl
	}
}
