package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/language"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// ExampleI18nHandler 国际化示例处理器结构体
//
// @description 演示如何在处理函数中使用国际化功能
// @struct
type ExampleI18nHandler struct {
	*Handler                       // Handler 基础处理器，提供日志功能
	service  service.SystemService // service 系统服务，处理系统相关的业务逻辑
}

// NewExampleI18nHandler 创建国际化示例处理器
//
// @description 创建并返回一个新的国际化示例处理器实例
// @param handler *Handler 基础处理器
// @param service service.SystemService 系统服务
// @return *ExampleI18nHandler 国际化示例处理器实例
func NewExampleI18nHandler(handler *Handler, service service.SystemService) *ExampleI18nHandler {
	return &ExampleI18nHandler{
		Handler: handler,
		service: service,
	}
}

// ExampleSuccessHandler 示例成功响应处理函数
//
// @description 演示如何返回国际化成功响应
// @method GET
// @url /example/i18n/success
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *ExampleI18nHandler) ExampleSuccessHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "exampleI18nHandler", "exampleSuccessHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received example success handler...")

	// 返回成功响应，消息会根据语言设置自动翻译
	commonhttp.SuccessResponse(ctx, map[string]interface{}{
		"message": "This is example data",
		"lang":    language.GetLanguageFromContext(ctx),
	})
	span.SetStatus(codes.Ok, "success")
}

// ExampleErrorHandler 示例错误响应处理函数
//
// @description 演示如何返回国际化错误响应
// @method GET
// @url /example/i18n/error
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *ExampleI18nHandler) ExampleErrorHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "exampleI18nHandler", "exampleErrorHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received example error handler...")

	// 返回错误响应，消息会根据语言设置自动翻译
	commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
	span.SetStatus(codes.Ok, "success")
}

// ExampleBindingErrorHandler 示例参数绑定错误处理函数
//
// @description 演示如何处理参数绑定错误并返回国际化错误响应
// @method GET
// @url /example/i18n/binding-error
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *ExampleI18nHandler) ExampleBindingErrorHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "exampleI18nHandler", "exampleBindingErrorHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received example binding error handler...")

	// 绑定请求参数
	req := new(request.SystemOverviewRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}

		if len(msg) == 0 {
			errorMessage := commonhttp.GetErrorMessageWithContext(ctx, commonhttp.InvalidParameter)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, errorMessage)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			errorMessage := commonhttp.GetErrorMessageWithContext(ctx, commonhttp.InvalidParameter)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, errorMessage+": "+stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("request is not legal...")
		// 使用国际化错误消息
		errorMessage := commonhttp.GetErrorMessageWithContext(ctx, commonhttp.InvalidParameter)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, errorMessage)
		span.SetStatus(codes.Error, errorMessage)
		return
	}

	// 返回成功响应
	commonhttp.SuccessResponse(ctx, map[string]interface{}{
		"message": "Validation passed",
		"lang":    language.GetLanguageFromContext(ctx),
	})
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取国际化示例相关路由列表
//
// @description 返回国际化示例相关的所有HTTP路由定义
// @return []commonhttp.RouterHandler 国际化示例路由处理器列表
func (h *ExampleI18nHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:        "exampleSuccess",
			Uri:         "/i18n/success",
			Enabled:     true,
			EnableAuth:  false,
			Method:      http.MethodGet,
			Desc:        "国际化成功响应示例",
			HandlerFunc: h.ExampleSuccessHandler,
		},
		&commonhttp.MyRouter{
			Name:        "exampleError",
			Uri:         "/i18n/error",
			Enabled:     true,
			EnableAuth:  false,
			Method:      http.MethodGet,
			Desc:        "国际化错误响应示例",
			HandlerFunc: h.ExampleErrorHandler,
		},
		&commonhttp.MyRouter{
			Name:        "exampleBindingError",
			Uri:         "/i18n/binding-error",
			Enabled:     true,
			EnableAuth:  false,
			Method:      http.MethodGet,
			Desc:        "国际化参数绑定错误示例",
			HandlerFunc: h.ExampleBindingErrorHandler,
		},
	}
}
