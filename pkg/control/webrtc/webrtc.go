package webrtc

import (
	"fmt"

	"github.com/jianlu8023/go-tools/v2/pkg/encoding/base64"
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/pion/webrtc/v4"
)

// ParseICECandidate 解析ICE候选字符串
//
// @description 解析ICE候选字符串，该函数首先尝试Base64解码，如果失败则直接解析JSON
// @param candidate string ICE候选字符串，可能是Base64编码或JSON格式
// @return webrtc.ICECandidateInit 解析后的ICE候选对象
// @return error 如果解析失败，则返回错误信息
func ParseICECandidate(candidate string) (webrtc.ICECandidateInit, error) {
	// 首先尝试Base64解码
	bytes, err := base64.ToByte(candidate)
	if err != nil {
		// 如果Base64解码失败，尝试直接解析JSON
		var iceCandidate webrtc.ICECandidateInit
		err = json.Unmarshal([]byte(candidate), &iceCandidate)
		if err != nil {
			return webrtc.ICECandidateInit{}, fmt.Errorf("无法解析ICE候选: %v", err)
		}
		return iceCandidate, nil
	}

	var iceCandidate webrtc.ICECandidateInit
	err = json.Unmarshal(bytes, &iceCandidate)
	if err != nil {
		return webrtc.ICECandidateInit{}, fmt.Errorf("JSON解析失败: %v", err)
	}
	return iceCandidate, nil
}

// ParseSessionDescription 解析SDP描述
//
// @description 解析SDP描述，该函数首先尝试Base64解码，如果失败则直接解析JSON，如果还失败则假设是纯SDP字符串
// @param sdp string SDP描述字符串，可能是Base64编码、JSON格式或纯SDP字符串
// @return webrtc.SessionDescription 解析后的SDP描述对象
// @return error 如果解析失败，则返回错误信息
func ParseSessionDescription(sdp string) (webrtc.SessionDescription, error) {
	// 首先尝试Base64解码
	bytes, err := base64.ToByte(sdp)
	if err != nil {
		// 如果Base64解码失败，尝试直接解析JSON
		var desc webrtc.SessionDescription
		err = json.Unmarshal([]byte(sdp), &desc)
		if err != nil {
			// 如果JSON解析也失败，假设是纯SDP字符串
			desc = webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  sdp,
			}
			return desc, nil
		}
		return desc, nil
	}
	// Base64解码成功，尝试解析JSON
	var desc webrtc.SessionDescription
	err = json.Unmarshal(bytes, &desc)
	if err != nil {
		return webrtc.SessionDescription{}, fmt.Errorf("JSON解析失败: %v", err)
	}
	return desc, nil
}

// GatheringCompletePromise 返回一个在ICE收集完成时关闭的通道
//
// @description 返回一个在ICE收集完成时关闭的通道
// @param peerConnection *PeerConnection WebRTC对等连接对象
// @return <-chan struct{} ICE收集完成时关闭的通道
func GatheringCompletePromise(peerConnection *PeerConnection) <-chan struct{} {
	return webrtc.GatheringCompletePromise(peerConnection.Connection)
}
