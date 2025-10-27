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

// WebRTCHandler WebRTC处理器
type WebRTCHandler struct {
	*Handler
	service *service.WebRTCService
}

// NewWebRTCHandler 创建新的WebRTC处理器
func NewWebRTCHandler(baseHandler *Handler, webRTCService *service.WebRTCService) *WebRTCHandler {
	return &WebRTCHandler{
		Handler: baseHandler,
		service: webRTCService,
	}
}

// sdpOfferHandler 处理SDP Offer请求
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
func (h *WebRTCHandler) getICECandidatesHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCHandler", "getICECandidatesHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received get ice candidates handler...")

	// 从URL参数获取连接ID
	connectionID := ctx.Param("connectionId")
	if stringer.IsBlank(connectionID) {
		h.logger.Errorf("connection id is required")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "Connection ID is required")
		span.SetStatus(codes.Error, "Connection ID is required")
		return
	}

	h.service.GetICECandidates(ctx, connectionID)
	span.SetStatus(codes.Ok, "success")
}

// closeConnectionHandler 处理关闭连接的请求
func (h *WebRTCHandler) closeConnectionHandler(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCHandler", "closeConnectionHandler",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received close connection handler...")

	// 从URL参数获取连接ID
	connectionID := ctx.Param("connectionId")
	if stringer.IsBlank(connectionID) {
		h.logger.Errorf("connection id is required")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "Connection ID is required")
		span.SetStatus(codes.Error, "Connection ID is required")
		return
	}

	h.service.CloseConnection(ctx, connectionID)
	span.SetStatus(codes.Ok, "success")
}

// Routers 返回路由处理器列表
func (h *WebRTCHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:            "WebRTC Ping",
			Uri:             "webrtc/",
			Method:          http.MethodGet,
			HandlerFunc:     h.pingHandler,
			Enabled:         true,
			Desc:            "webrtc ping endpoint",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "WebRTC SDP Offer",
			Uri:             "webrtc/offer",
			Method:          http.MethodPost,
			HandlerFunc:     h.sdpOfferHandler,
			Enabled:         true,
			Desc:            "webrtc sdp offer",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "WebRTC ICE Candidate",
			Uri:             "webrtc/ice",
			Method:          http.MethodPost,
			HandlerFunc:     h.iceCandidateHandler,
			Enabled:         true,
			Desc:            "webrtc ice candidate",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "WebRTC Get ICE Candidates",
			Uri:             "webrtc/ice/:connectionId",
			Method:          http.MethodGet,
			HandlerFunc:     h.getICECandidatesHandler,
			Enabled:         true,
			Desc:            "get webrtc ice candidates",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:            "WebRTC Close Connection",
			Uri:             "webrtc/close/:connectionId",
			Method:          http.MethodPost,
			HandlerFunc:     h.closeConnectionHandler,
			Enabled:         true,
			Desc:            "close webrtc connection",
			EnableJWtVerify: false,
		},
	}
}
