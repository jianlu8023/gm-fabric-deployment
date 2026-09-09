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
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// CertsHandler 证书处理器结构体
//
// @description 处理证书只读查询相关的HTTP请求（证书详情、证书目录列表、当前TLS连接信息）
// @struct
type CertsHandler struct {
	*Handler                      // Handler 基础处理器，提供日志功能
	service  service.CertsService // service 证书服务，处理证书相关的业务逻辑
}

// NewCertsHandler 创建证书处理器
//
// @description 创建并返回一个新的证书处理器实例
// @param handler *Handler 基础处理器
// @param service service.CertsService 证书服务
// @return *CertsHandler 证书处理器实例
func NewCertsHandler(handler *Handler, service service.CertsService) *CertsHandler {
	return &CertsHandler{
		Handler: handler,
		service: service,
	}
}

// GetCertsTls 获取当前TLS连接信息处理函数
//
// @description 返回当前 HTTPS 连接的 TLS/mTLS 协商信息与对端证书信息
// @method GET
// @url /certs/tls
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *CertsHandler) GetCertsTls(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "certsHandler", "getCertsTls",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Debugf("received certs tls handler...")
	req := new(request.CertsTlsRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		h.bindCertsError(ctx, span, err, "certs tls")
		return
	}
	if !req.IsLegal() {
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.GetCertsTls(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// bindCertsError 统一处理证书接口参数绑定错误
//
// @description 解析参数绑定/校验错误并返回统一的失败响应，同时记录链路状态
// @param ctx *gin.Context Gin上下文
// @param span trace.Span 当前链路span
// @param err error 绑定错误
// @param scene string 场景描述，用于日志
func (h *CertsHandler) bindCertsError(ctx *gin.Context, span trace.Span, err error, scene string) {
	messages := binding.GetValidationErrorMessages(err)
	h.logger.Errorf("binding %s request params failed: %v message: %v", scene, err, messages)
	msg := make([]string, 0, len(messages))
	for _, message := range messages {
		msg = append(msg, message)
	}
	if len(msg) == 0 {
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	joined := stringer.Join(msg, ",")
	commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, joined)
	span.RecordError(err)
	span.SetStatus(codes.Error, joined)
}

// Routers 获取证书相关路由列表
//
// @description 返回证书只读查询相关的所有HTTP路由定义
// @return []commonhttp.RouterHandler 证书路由处理器列表
func (h *CertsHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{

		&commonhttp.MyRouter{
			Name:        "certsTls",
			Uri:         "certs/tls",
			Enabled:     true,
			EnableAuth:  false,
			Method:      http.MethodGet,
			Desc:        "获取当前HTTPS连接的TLS/mTLS协商信息与对端证书",
			HandlerFunc: h.GetCertsTls,
		},
	}
}
