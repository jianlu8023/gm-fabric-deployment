package handler

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// WebSocketHandler WebSocket处理器结构体
//
// @description 处理WebSocket相关的HTTP请求和连接管理
// @struct
type WebSocketHandler struct {
	*Handler                          // Handler 基础处理器，提供日志功能
	service  service.WebSocketService // service WebSocket服务，处理WebSocket相关的业务逻辑
}

// NewWebSocketHandler 创建WebSocket处理器
//
// @description 创建并返回一个新的WebSocket处理器实例
// @param handler *Handler 基础处理器
// @param service service.WebSocketService WebSocket服务
// @return *WebSocketHandler WebSocket处理器实例
func NewWebSocketHandler(handler *Handler, service service.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		Handler: handler,
		service: service,
	}
}

// UpgradeHandler WebSocket连接升级处理函数
//
// @description 将HTTP连接升级为WebSocket连接，是WebSocket通信的核心入口
// @method GET
// @url /ws
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebSocketHandler) UpgradeHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "websocketHandler", "upgradeHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Infof("received websocket upgrade handler...")

	// 验证JWT和Session信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("JWT authentication failed: userID not found")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "认证失败，请先登录")
		span.SetStatus(codes.Error, "认证失败，请先登录")
		return
	}

	sessionID, exists := ctx.Get("session_id")
	if !exists {
		h.logger.Errorf("Session validation failed: sessionID not found")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "会话已过期，请重新登录")
		span.SetStatus(codes.Error, "会话已过期，请重新登录")
		return
	}

	h.logger.Debugf("User %v with session %v is requesting websocket upgrade", userID, sessionID)

	// 绑定请求参数
	req := new(request.WSConnectRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
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

	// 验证请求参数
	if !req.IsLegal() {
		h.logger.Debugf("websocket connect request is illegal: %v", req)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "输入参数不合法")
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用service层处理业务逻辑
	h.service.ConnectService(ctx, req, userID)
	span.SetStatus(codes.Ok, "success")
}

// DisconnectHandler WebSocket断开连接处理函数
//
// @description 处理WebSocket连接断开请求
// @method POST
// @url /ws/disconnect
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebSocketHandler) DisconnectHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "websocketHandler", "disconnectHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Infof("received websocket disconnect request...")

	// 验证JWT和Session信息
	_, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("JWT authentication failed: userID not found")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "认证失败，请先登录")
		span.SetStatus(codes.Error, "认证失败，请先登录")
		return
	}

	_, exists = ctx.Get("session_id")
	if !exists {
		h.logger.Errorf("Session validation failed: sessionID not found")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "会话已过期，请重新登录")
		span.SetStatus(codes.Error, "会话已过期，请重新登录")
		return
	}

	// 绑定请求参数
	req := new(request.WSDisconnectRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
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

	// 验证请求参数
	if !req.IsLegal() {
		h.logger.Debugf("websocket disconnect request is illegal: %v", req)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "输入参数不合法")
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用service层处理业务逻辑
	h.service.DisconnectService(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// SendMessageHandler 发送WebSocket消息处理函数
//
// @description 处理通过WebSocket发送消息的请求
// @method POST
// @url /ws/message
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebSocketHandler) SendMessageHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "websocketHandler", "sendMessageHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Infof("received websocket send message request...")

	// 验证JWT和Session信息
	_, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("JWT authentication failed: userID not found")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "认证失败，请先登录")
		span.SetStatus(codes.Error, "认证失败，请先登录")
		return
	}

	_, exists = ctx.Get("session_id")
	if !exists {
		h.logger.Errorf("Session validation failed: sessionID not found")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "会话已过期，请重新登录")
		span.SetStatus(codes.Error, "会话已过期，请重新登录")
		return
	}

	// 绑定请求参数
	req := new(request.WSMessageRequest)
	if err := binding.BindJSON(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
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

	// 验证请求参数
	if !req.IsLegal() {
		h.logger.Debugf("websocket message request is illegal: %v", req)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "输入参数不合法")
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用service层处理业务逻辑
	h.service.SendMessageService(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// GetConnectionListHandler 获取WebSocket连接列表处理函数
//
// @description 获取当前系统中的WebSocket连接列表
// @method GET
// @url /ws/connections
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebSocketHandler) GetConnectionListHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "websocketHandler", "getConnectionListHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Infof("received get connection list request...")

	// 验证JWT和Session信息
	userID, exists := ctx.Get("user_id")
	if !exists {
		h.logger.Errorf("JWT authentication failed: userID not found")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.Unauthorized, "认证失败，请先登录")
		span.SetStatus(codes.Error, "认证失败，请先登录")
		return
	}

	userIDStr := userID.(string)

	// 绑定分页参数
	req := new(request.WSConnectionListRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
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

	// 验证请求参数
	if !req.IsLegal() {
		h.logger.Debugf("websocket connection list request is illegal: %v", req)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "输入参数不合法")
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用service层处理业务逻辑
	h.service.GetConnectionListService(ctx, req, userIDStr)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取WebSocket相关路由列表
//
// @description 返回WebSocket相关的所有HTTP路由定义
// @return []commonhttp.RouterHandler WebSocket路由处理器列表
func (h *WebSocketHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		// WebSocket连接升级路由
		&commonhttp.MyRouter{
			Name:        "websocketUpgrade",
			Uri:         "ws",
			Method:      http.MethodGet,
			HandlerFunc: h.UpgradeHandler,
			Enabled:     true,
			EnableAuth:  true,
			Desc:        "WebSocket连接升级",
		},
		// WebSocket断开连接路由
		&commonhttp.MyRouter{
			Name:        "websocketDisconnect",
			Uri:         "ws/disconnect",
			Method:      http.MethodPost,
			HandlerFunc: h.DisconnectHandler,
			Enabled:     true,
			EnableAuth:  true,
			Desc:        "WebSocket断开连接请求",
		},
		// WebSocket发送消息路由
		&commonhttp.MyRouter{
			Name:        "websocketSendMessage",
			Uri:         "ws/message",
			Method:      http.MethodPost,
			HandlerFunc: h.SendMessageHandler,
			Enabled:     true,
			EnableAuth:  true,
			Desc:        "发送WebSocket消息",
		},
		// 获取WebSocket连接列表路由
		&commonhttp.MyRouter{
			Name:        "websocketConnectionList",
			Uri:         "ws/connections",
			Method:      http.MethodGet,
			HandlerFunc: h.GetConnectionListHandler,
			Enabled:     true,
			EnableAuth:  true,
			Desc:        "获取WebSocket连接列表",
		},
	}
}
