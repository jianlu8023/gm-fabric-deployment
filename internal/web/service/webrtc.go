package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/encoding/base64"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/pkg/control/webrtc"
	"github.com/pion/randutil"
	webrtcoffical "github.com/pion/webrtc/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// WebRTCService WebRTC服务
type WebRTCService struct {
	*Service
	webrtcControl *webrtc.Control
}

// NewWebRTCService 创建新的WebRTC服务
func NewWebRTCService(baseService *Service, webrtcControl *webrtc.Control) *WebRTCService {
	return &WebRTCService{
		Service:       baseService,
		webrtcControl: webrtcControl,
	}
}

// SdpOffer 处理SDP Offer请求
func (s *WebRTCService) SdpOffer(ctx *gin.Context, req *request.WebRTCOfferRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCService", "SdpOffer",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received sdp offer request: %v", req)

	connectionId, peerConnection, err := s.webrtcControl.CreatePeerConnection(nil)
	if err != nil {
		s.logger.Errorf("create peer connection failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, err.Error())
		return
	}

	// 设置数据通道回调
	peerConnection.OnDataChannel = func(dc *webrtcoffical.DataChannel) {
		s.logger.Infof("Data channel created: %s", dc.Label())
		ticker := time.NewTicker(5 * time.Second)
		dc.OnClose(func() {
			s.logger.Infof("Data channel '%s'-'%d' has been closed", dc.Label(), dc.ID())
			ticker.Stop()
		})
		dc.OnOpen(func() {
			s.logger.Infof("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds", dc.Label(), dc.ID())

			defer ticker.Stop()
			for range ticker.C {
				message, sendErr := randutil.GenerateCryptoRandomString(15, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
				if sendErr != nil {
					s.logger.Errorf("Failed to generate random string: %v", sendErr)
					continue
				}

				// Send the message as text
				s.logger.Infof("Sending '%s'", message)
				if sendErr = dc.SendText(message); sendErr != nil {
					s.logger.Errorf("Failed to send message: %v", sendErr)
				}
			}
		})

		// Register text message handling
		dc.OnMessage(func(msg webrtcoffical.DataChannelMessage) {
			s.logger.Infof("Message from DataChannel '%s': '%s'", dc.Label(), string(msg.Data))
		})
	}

	// 解析SDP Offer
	s.logger.Debugf("Parsing SDP offer...")
	offer, err := webrtc.ParseSessionDescription(req.Offer)
	if err != nil {
		s.logger.Errorf("parse sdp offer failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Invalid SDP offer")
		return
	}
	s.logger.Debugf("SDP offer parsed successfully, type: %s", offer.Type)

	// 设置远程描述
	s.logger.Debugf("Setting remote description...")
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		s.logger.Errorf("set remote description failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to set remote description")
		return
	}
	s.logger.Debugf("Remote description set successfully")

	// 处理随Offer一起发送的ICE候选
	s.logger.Debugf("Processing ICE candidates from offer, count: %d", len(req.Candidates))
	for i, candidateStr := range req.Candidates {
		if candidateStr != "" {
			s.logger.Debugf("Processing ICE candidate %d: %s", i, candidateStr)
			candidateInit, err := webrtc.ParseICECandidate(candidateStr)
			if err != nil {
				s.logger.Warnf("parse ice candidate failed: %v", err)
				continue
			}

			// 添加ICE候选
			if err := peerConnection.AddICECandidateInit(candidateInit); err != nil {
				s.logger.Warnf("add ice candidate failed: %v", err)
			}
		}
	}

	// 添加存储的ICE候选
	s.logger.Debugf("Adding stored ICE candidates...")
	if err := peerConnection.AddStoredICECandidates(); err != nil {
		s.logger.Errorf("add stored ice candidates failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to add stored ICE candidates")
		return
	}

	// 创建Answer
	s.logger.Debugf("Creating answer...")
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		s.logger.Errorf("create answer failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to create answer")
		return
	}
	s.logger.Debugf("Answer created successfully")

	// 设置本地描述
	s.logger.Debugf("Setting local description...")
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		s.logger.Errorf("set local description failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to set local description")
		return
	}
	s.logger.Debugf("Local description set successfully")

	// 返回Answer和连接ID
	s.logger.Debugf("Marshaling answer...")
	bytes, err := json.Marshal(answer)
	if err != nil {
		s.logger.Errorf("marshal answer failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to marshal answer")
		return
	}
	answerBase64 := base64.ToBase64(bytes)
	s.logger.Debugf("Answer marshaled successfully, base64 length: %d", len(answerBase64))

	resp := &response.WebRTCSDPOfferResponse{
		ConnectionId: connectionId,
		Answer:       answerBase64,
	}
	s.logger.Debugf("Sending response with connection ID: %s", connectionId)
	commonhttp.SuccessResponse(ctx, resp)
	span.SetStatus(codes.Ok, "success")
}

// AddICECandidate 处理ICE候选
func (s *WebRTCService) AddICECandidate(ctx *gin.Context, req *request.WebRTCIceCandidateRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCService", "AddICECandidate",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received ice candidate request: %v", req)

	// 获取PeerConnection
	peerConnection := s.webrtcControl.GetPeerConnection(req.ConnectionID)
	if peerConnection == nil {
		s.logger.Errorf("peer connection not found: %s", req.ConnectionID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Connection not found")
		return
	}

	// 解析ICE候选
	s.logger.Debugf("parsing ice candidate: %s", req.Candidate)
	candidateInit, err := webrtc.ParseICECandidate(req.Candidate)
	if err != nil {
		s.logger.Errorf("parse ice candidate failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Invalid ICE candidate")
		return
	}
	s.logger.Debugf("ice candidate parsed successfully: %s", candidateInit.Candidate)

	// 添加ICE候选
	if err := peerConnection.AddICECandidateInit(candidateInit); err != nil {
		s.logger.Errorf("add ice candidate failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to add ICE candidate")
		return
	}

	s.logger.Debugf("ice candidate added successfully: %s", candidateInit.Candidate)
	resp := &response.WebRTCICECandidateResponse{
		Success: true,
		Message: "ICE candidate added successfully",
	}
	commonhttp.SuccessResponse(ctx, resp)
	span.SetStatus(codes.Ok, "success")
}

// GetICECandidates 获取指定连接的ICE候选列表
func (s *WebRTCService) GetICECandidates(ctx *gin.Context, connectionID string) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCService", "GetICECandidates",
		attribute.String("connectionId", connectionID),
	)
	defer span.End()
	s.logger.Debugf("received get ice candidates request for connection: %s", connectionID)

	// 获取PeerConnection
	peerConnection := s.webrtcControl.GetPeerConnection(connectionID)
	if peerConnection == nil {
		s.logger.Errorf("peer connection not found: %s", connectionID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Connection not found")
		return
	}

	// 获取存储的ICE候选
	candidates := peerConnection.GetICECandidates()
	s.logger.Debugf("found %d ice candidates for connection: %s", len(candidates), connectionID)

	// 如果没有候选，返回空响应
	if len(candidates) == 0 {
		s.logger.Debugf("no ice candidates found for connection: %s", connectionID)
		resp := &response.WebRTCSingleICECandidateResponse{
			Success: true,
			Message: "No ICE candidates available",
		}
		commonhttp.SuccessResponse(ctx, resp)
		span.SetStatus(codes.Ok, "success")
		return
	}

	// 只返回第一个有效的候选
	var firstCandidate webrtcoffical.ICECandidateInit
	for _, candidate := range candidates {
		if candidate.Candidate != "" {
			firstCandidate = candidate
			s.logger.Debugf("found valid ice candidate: %s", candidate.Candidate)
			break
		}
	}

	// 如果没有有效的候选
	if firstCandidate.Candidate == "" {
		s.logger.Debugf("no valid ice candidates found for connection: %s", connectionID)
		resp := &response.WebRTCSingleICECandidateResponse{
			Success: true,
			Message: "No valid ICE candidates available",
		}
		commonhttp.SuccessResponse(ctx, resp)
		span.SetStatus(codes.Ok, "success")
		return
	}

	// 序列化候选信息为JSON字符串
	candidateBytes, err := json.Marshal(firstCandidate)
	if err != nil {
		s.logger.Errorf("marshal ice candidate failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to marshal ICE candidate")
		return
	}

	s.logger.Debugf("sending ice candidate: %s", string(candidateBytes))
	resp := &response.WebRTCSingleICECandidateResponse{
		Success:   true,
		Candidate: string(candidateBytes), // 确保这里返回的是JSON字符串
		Message:   "ICE candidate sent successfully",
	}
	commonhttp.SuccessResponse(ctx, resp)
	span.SetStatus(codes.Ok, "success")
}

// CloseConnection 关闭指定的WebRTC连接
func (s *WebRTCService) CloseConnection(ctx *gin.Context, connectionID string) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "webRTCService", "CloseConnection",
		attribute.String("connectionId", connectionID),
	)
	defer span.End()
	s.logger.Debugf("received close connection request for connection: %s", connectionID)

	// 关闭PeerConnection
	if err := s.webrtcControl.ClosePeerConnection(connectionID); err != nil {
		s.logger.Errorf("close peer connection failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to close connection")
		return
	}

	s.logger.Debugf("connection closed successfully: %s", connectionID)
	resp := &response.WebRTCICECandidateResponse{
		Success: true,
		Message: "Connection closed successfully",
	}
	commonhttp.SuccessResponse(ctx, resp)
	span.SetStatus(codes.Ok, "success")
}
