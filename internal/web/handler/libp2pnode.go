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

// Libp2pNodeHandler 节点处理器结构体
//
// @description 处理节点相关的HTTP请求
// @struct
type Libp2pNodeHandler struct {
	*Handler                           // Handler 基础处理器，提供日志功能
	service  service.Libp2pNodeService // service 节点服务，处理节点相关的业务逻辑
}

// NewLibp2pNodeHandler 创建节点处理器
//
// @description 创建并返回一个新的节点处理器实例
// @param handler *Handler 基础处理器
// @param service service.Libp2pNodeService 节点服务
// @return *Libp2pNodeHandler 节点处理器实例
func NewLibp2pNodeHandler(handler *Handler, libp2pNodeService service.Libp2pNodeService) *Libp2pNodeHandler {
	return &Libp2pNodeHandler{
		Handler: handler,
		service: libp2pNodeService,
	}
}

// Libp2pNodeList 获取节点列表的处理函数
//
// @description 处理获取节点列表的HTTP请求，验证参数并调用服务层获取节点列表
// @method GET
// @url libp2p/list
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *Libp2pNodeHandler) Libp2pNodeList(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "libp2pHandler", "libp2pNodeList",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Infof("received libp2p node list handler...")

	// 验证JWT和Session信息
	// userID, exists := ctx.Get("user_id")
	// if !exists {
	// 	h.logger.Errorf("JWT authentication failed: userID not found")
	// 	webhhtp.FailedResponseWithMessage(ctx, webhhtp.Unauthorized, "认证失败，请先登录")
	// 	return
	// }

	// sessionID, exists := ctx.Get("session_id")
	// if !exists {
	// 	h.logger.Errorf("Session validation failed: sessionID not found")
	// 	webhhtp.FailedResponseWithMessage(ctx, webhhtp.SessionExpired, "会话已过期，请重新登录")
	// 	return
	// }

	// h.logger.Debugf("User %v with session %v is accessing node list", userID, sessionID)

	// 验证通过后继续处理请求
	req := new(request.Libp2pNodeListRequest)
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
	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("libp2p node list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.Libp2pNodeList(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Libp2pNodeMyself 获取本机节点信息的处理函数
//
// @description 处理获取本机节点信息的HTTP请求，验证参数并调用服务层获取本机节点信息
// @method GET
// @url libp2p/myself
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *Libp2pNodeHandler) Libp2pNodeMyself(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "libp2pHandler", "libp2pNodeMyself",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()

	h.logger.Infof("received libp2p node myself handler...")

	req := new(request.Libp2pNodeMyselfRequest)
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
	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("libp2p node list request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.Libp2pNodeMyself(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取节点相关路由列表
//
// @description 返回所有节点相关的HTTP路由配置，包括路由名称、URI、请求方法、处理函数、描述等信息
// @return []commonhttp.RouterHandler 节点相关的路由配置列表
func (h *Libp2pNodeHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:        "libp2pNodeList",
			Uri:         "libp2p/list",
			Method:      http.MethodGet,
			HandlerFunc: h.Libp2pNodeList,
			Enabled:     true,
			Desc:        "获取libp2p已发现的节点",
			EnableAuth:  false,
		},
		&commonhttp.MyRouter{
			Name:        "libp2pNodeMyself",
			Uri:         "libp2p/myself",
			Method:      http.MethodGet,
			HandlerFunc: h.Libp2pNodeMyself,
			Enabled:     true,
			Desc:        "获取libp2p本机节点",
			EnableAuth:  false,
		},
	}
}
