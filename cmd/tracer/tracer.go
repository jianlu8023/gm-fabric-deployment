package main

import (
	"context"

	"github.com/jianlu8023/golang-example/pkg/control/tracer"
)

func main() {
	_, span := tracer.StartSpan(context.Background(), "", "")
	defer span.End()

}
