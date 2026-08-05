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
)

// GrpcHandler gRPC处理器结构体
//
// @description 处理gRPC相关的HTTP请求
// @struct
type GrpcHandler struct {
	*Handler                     // Handler 基础处理器，提供日志功能
	service  service.GrpcService // service gRPC服务，处理gRPC相关的业务逻辑
}

// NewGrpcHandler 创建gRPC处理器
//
// @description 创建并返回一个新的gRPC处理器实例
// @param baseHandler *Handler 基础处理器
// @param grpcService service.GrpcService gRPC服务
// @return *GrpcHandler gRPC处理器实例
func NewGrpcHandler(baseHandler *Handler, grpcService service.GrpcService) *GrpcHandler {
	return &GrpcHandler{
		Handler: baseHandler,
		service: grpcService,
	}
}

// SendGrpcPingMessage 发送gRPC Ping消息处理函数
//
// @description 处理发送gRPC Ping消息的HTTP请求
// @method GET
// @url /grpc/send/message/ping
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *GrpcHandler) SendGrpcPingMessage(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "grpcHandler", "sendGrpcPingMessage",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received grpc send ping message handler...")

	// 绑定请求参数
	req := new(request.GrpcSendPingMessageRequest)
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

	// 验证参数合法性
	if !req.IsLegal() {
		h.logger.Errorf("grpc send ping message request is illegal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.GrpcPingMessage(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取gRPC相关路由列表
//
// @description 返回gRPC相关的所有HTTP路由定义
// @return []commonhttp.RouterHandler gRPC路由处理器列表
func (h *GrpcHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:        "SendGrpcPingMessage",
			Uri:         "/grpc/send/message/ping",
			Method:      http.MethodGet,
			HandlerFunc: h.SendGrpcPingMessage,
			Enabled:     true,
			EnableAuth:  false,
			Desc:        "发送Grpc的pingMessage",
		},
	}
}
