package service

import (
	"fmt"
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

type WebRTCService struct {
	*Service
	webrtcControl *webrtc.Control
}

func NewWebRTCService(baseService *Service, webrtcControl *webrtc.Control) *WebRTCService {
	return &WebRTCService{
		Service:       baseService,
		webrtcControl: webrtcControl,
	}
}

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
	peerConnection.OnDataChannel = func(dc *webrtcoffical.DataChannel) {
		s.logger.Infof("Data channel created: %s", dc.Label())
		dc.OnOpen(func() {
			fmt.Printf(
				"Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n",
				dc.Label(), dc.ID(),
			)

			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				message, sendErr := randutil.GenerateCryptoRandomString(15, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
				if sendErr != nil {
					panic(sendErr)
				}

				// Send the message as text
				fmt.Printf("Sending '%s'\n", message)
				if sendErr = dc.SendText(message); sendErr != nil {
					panic(sendErr)
				}
			}
		})

		// Register text message handling
		dc.OnMessage(func(msg webrtcoffical.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", dc.Label(), string(msg.Data))
		})
	}

	// 解析SDP Offer
	offer, err := webrtc.ParseSessionDescription(req.Offer)
	if err != nil {
		s.logger.Errorf("parse sdp offer failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Invalid SDP offer")
		return
	}

	// 设置远程描述
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		s.logger.Errorf("set remote description failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to set remote description")
		return
	}

	// 创建Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		s.logger.Errorf("create answer failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to create answer")
		return
	}

	// 设置本地描述
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		s.logger.Errorf("set local description failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to set local description")
		return
	}

	// 返回Answer和连接ID
	bytes, err := json.Marshal(answer)
	if err != nil {
		s.logger.Errorf("marshal answer failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to marshal answer")
		return
	}
	answerBase64 := base64.ToBase64(bytes)
	resp := &response.WebRTCSDPOfferResponse{
		ConnectionId: connectionId,
		Answer:       answerBase64,
	}
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
	candidate, err := webrtc.ParseICECandidate(req.Candidate)
	if err != nil {
		s.logger.Errorf("parse ice candidate failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Invalid ICE candidate")
		return
	}

	// 添加ICE候选
	if err := peerConnection.AddICECandidate(candidate); err != nil {
		s.logger.Errorf("add ice candidate failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "Failed to add ICE candidate")
		return
	}

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

	// 转换为响应格式
	candidateInfos := make([]response.WebRTCICECandidateInfo, len(candidates))
	for i, candidate := range candidates {
		// 处理可能为nil的指针字段
		var sdpMid string
		if candidate.SDPMid != nil {
			sdpMid = *candidate.SDPMid
		}

		var sdpMLineIndex uint16
		if candidate.SDPMLineIndex != nil {
			sdpMLineIndex = *candidate.SDPMLineIndex
		}

		var usernameFragment string
		if candidate.UsernameFragment != nil {
			usernameFragment = *candidate.UsernameFragment
		}

		candidateInfos[i] = response.WebRTCICECandidateInfo{
			Candidate:        candidate.Candidate,
			SdpMid:           sdpMid,
			SdpMLineIndex:    sdpMLineIndex,
			UsernameFragment: usernameFragment,
		}
	}

	resp := &response.WebRTCICECandidateListResponse{
		Success:    true,
		Candidates: candidateInfos,
	}
	commonhttp.SuccessResponse(ctx, resp)
	span.SetStatus(codes.Ok, "success")
}
