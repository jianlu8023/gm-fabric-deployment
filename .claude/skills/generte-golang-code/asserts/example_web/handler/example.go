package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/request"
	"github.com/jianlu8023/golang-example/.claude/skills/generte-golang-code/asserts/example_web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
)

// ExampleHandler ExampleHandler接口
//
// struct
type ExampleHandler struct {
	*Handler
	exampleService service.ExampleService
}

// exampleHandler example接口
//
// @param ctx *gin.Context: gin上下文
func (h *ExampleHandler) exampleHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "handler", "ping")
	defer span.End()

	h.logger.Debugf("receive /example request from %s", ctx.RemoteIP())

	req := new(request.ExampleRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages,
		)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}
	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("docker image list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.exampleService.ExampleService(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 路由
//
// @return []commonhttp.RouterHandler: 路由定义
func (h *ExampleHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "example",
			Uri:             "/example",
			Method:          http.MethodGet,
			HandlerFunc:     h.exampleHandler,
			Enabled:         true,
			Desc:            "example样例",
			EnableJWtVerify: false,
		},
	}
}
