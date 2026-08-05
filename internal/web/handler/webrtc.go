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

// WebRTCHandler WebRTC处理器结构体
//
// @description 处理WebRTC相关的HTTP请求，包括SDP协商、ICE候选交换和连接管理
// @struct
type WebRTCHandler struct {
	*Handler                       // Handler 基础处理器，提供日志功能
	service  service.WebRTCService // service WebRTC服务，处理WebRTC相关的业务逻辑
}

// NewWebRTCHandler 创建WebRTC处理器
//
// @description 创建并返回一个新的WebRTC处理器实例
// @param baseHandler *Handler 基础处理器
// @param webRTCService service.WebRTCService WebRTC服务
// @return *WebRTCHandler WebRTC处理器实例
func NewWebRTCHandler(baseHandler *Handler, webRTCService service.WebRTCService) *WebRTCHandler {
	return &WebRTCHandler{
		Handler: baseHandler,
		service: webRTCService,
	}
}

// sdpOfferHandler 处理SDP Offer请求
//
// @description 处理WebRTC信令交换的SDP Offer请求，绑定参数并调用服务层进行SDP协商
// @method POST
// @url /webrtc/offer
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebRTCHandler) sdpOfferHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCHandler", "sdpOfferHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received sdp offer handler...")

	req := new(request.WebRTCOfferRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
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
		h.logger.Errorf("register user request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.SdpOffer(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// iceCandidateHandler 处理ICE候选请求
//
// @description 接收客户端提交的ICE候选并调用服务层添加到指定连接的PeerConnection
// @method POST
// @url /webrtc/ice
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebRTCHandler) iceCandidateHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCHandler", "iceCandidateHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received ice candidate handler...")

	req := new(request.WebRTCIceCandidateRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
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
		h.logger.Errorf("ice candidate request is illegal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}
	h.service.AddICECandidate(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// getICECandidatesHandler 处理获取ICE候选列表的请求
//
// @description 根据连接ID从URI参数获取指定连接的ICE候选列表
// @method GET
// @url /webrtc/ice/:connectionId
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebRTCHandler) getICECandidatesHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCHandler", "getICECandidatesHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received get ice candidates handler...")

	req := new(request.GetICECandidatesRequest)
	if err := binding.BindURI(ctx, req); err != nil {
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
		h.logger.Errorf("get ice candidates request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.GetICECandidates(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// closeConnectionHandler 处理关闭连接的请求
//
// @description 根据连接ID从URI参数关闭指定的WebRTC连接
// @method POST
// @url /webrtc/close/:connectionId
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *WebRTCHandler) closeConnectionHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCHandler", "closeConnectionHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received close connection handler...")

	// 从URL参数获取连接ID
	req := new(request.CloseConnectionRequest)
	if err := binding.BindURI(ctx, req); err != nil {
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
		h.logger.Errorf("close connection request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	h.service.CloseConnection(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取WebRTC相关路由列表
//
// @description 返回WebRTC相关的所有HTTP路由定义，包括ping、SDP Offer、ICE候选等
// @return []commonhttp.RouterHandler WebRTC路由处理器列表
func (h *WebRTCHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:        "WebRTC Ping",
			Uri:         "webrtc/",
			Method:      http.MethodGet,
			HandlerFunc: h.pingHandler,
			Enabled:     true,
			Desc:        "webrtc ping endpoint",
			EnableAuth:  false,
		},
		&commonhttp.MyRouter{
			Name:        "WebRTC SDP Offer",
			Uri:         "webrtc/offer",
			Method:      http.MethodPost,
			HandlerFunc: h.sdpOfferHandler,
			Enabled:     true,
			Desc:        "webrtc sdp offer",
			EnableAuth:  false,
		},
		&commonhttp.MyRouter{
			Name:        "WebRTC ICE Candidate",
			Uri:         "webrtc/ice",
			Method:      http.MethodPost,
			HandlerFunc: h.iceCandidateHandler,
			Enabled:     true,
			Desc:        "webrtc ice candidate",
			EnableAuth:  false,
		},
		&commonhttp.MyRouter{
			Name:        "WebRTC Get ICE Candidates",
			Uri:         "webrtc/ice/:connectionId",
			Method:      http.MethodGet,
			HandlerFunc: h.getICECandidatesHandler,
			Enabled:     true,
			Desc:        "get webrtc ice candidates",
			EnableAuth:  false,
		},
		&commonhttp.MyRouter{
			Name:        "WebRTC Close Connection",
			Uri:         "webrtc/close/:connectionId",
			Method:      http.MethodPost,
			HandlerFunc: h.closeConnectionHandler,
			Enabled:     true,
			Desc:        "close webrtc connection",
			EnableAuth:  false,
		},
	}
}
