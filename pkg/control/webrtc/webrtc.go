package webrtc

import (
	"encoding/json"
	"sync"

	"github.com/jianlu8023/go-tools/v2/pkg/encoding/base64"
	"github.com/pion/webrtc/v4"
)

// MediaType 定义媒体类型
type MediaType string

const (
	MediaTypeAudio MediaType = "audio"
	MediaTypeVideo MediaType = "video"
	MediaTypeData  MediaType = "data"
)

// MediaTrack 表示一个媒体轨道
type MediaTrack struct {
	ID       string
	Type     MediaType
	Label    string
	Kind     string
	Track    *webrtc.TrackLocalStaticRTP
	Receiver *webrtc.RTPReceiver
}

// PeerConnectionState 定义对等连接状态
type PeerConnectionState string

const (
	PeerConnectionStateNew          PeerConnectionState = "new"
	PeerConnectionStateConnecting   PeerConnectionState = "connecting"
	PeerConnectionStateConnected    PeerConnectionState = "connected"
	PeerConnectionStateDisconnected PeerConnectionState = "disconnected"
	PeerConnectionStateFailed       PeerConnectionState = "failed"
	PeerConnectionStateClosed       PeerConnectionState = "closed"
)

// PeerConnection 封装WebRTC对等连接并整合客户端连接功能
// 这个结构体现在同时管理WebRTC连接和客户端通信功能
type PeerConnection struct {
	Connection                 *webrtc.PeerConnection
	ID                         string
	PeerID                     string // 添加PeerID字段，替代原Connection中的PeerID
	MediaTracks                map[string]*MediaTrack
	MediaTracksMutex           sync.RWMutex
	OnStateChange              func(state PeerConnectionState)
	OnTrack                    func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)
	OnICECandidate             func(candidate *webrtc.ICECandidate)
	OnICEConnectionStateChange func(state webrtc.ICEConnectionState)
	OnDataChannel              func(dc *webrtc.DataChannel)

	// ICE候选存储
	iceCandidates    []webrtc.ICECandidateInit
	iceCandidatesMux sync.RWMutex

	// 整合客户端连接功能
	Send     chan []byte // 用于向客户端发送消息的通道
	closeMux sync.Mutex  // 保护关闭操作的互斥锁
}

// CreateOffer 创建SDP Offer
func (wpc *PeerConnection) CreateOffer(options *webrtc.OfferOptions) (webrtc.SessionDescription, error) {
	return wpc.Connection.CreateOffer(options)
}

// CreateAnswer 创建SDP Answer
func (wpc *PeerConnection) CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error) {
	return wpc.Connection.CreateAnswer(options)
}

// AddTrack 添加媒体轨道
func (wpc *PeerConnection) AddTrack(track *webrtc.TrackLocalStaticRTP, kind string) (*MediaTrack, error) {
	// 添加轨道到对等连接
	_, err := wpc.Connection.AddTrack(track)
	if err != nil {
		return nil, err
	}

	// 创建媒体轨道对象
	mediaTrack := &MediaTrack{
		ID:    track.ID(),
		Type:  getMediaType(kind),
		Track: track,
	}

	// 添加到媒体轨道映射
	wpc.MediaTracksMutex.Lock()
	wpc.MediaTracks[track.ID()] = mediaTrack
	wpc.MediaTracksMutex.Unlock()

	return mediaTrack, nil
}

// RemoveTrack 移除媒体轨道
func (wpc *PeerConnection) RemoveTrack(id string) error {
	wpc.MediaTracksMutex.Lock()
	_, exists := wpc.MediaTracks[id]
	if exists {
		delete(wpc.MediaTracks, id)
	}
	wpc.MediaTracksMutex.Unlock()

	if !exists {
		return ErrMediaTrackNotFound
	}

	// 不需要显式调用RemoveTrack，因为在WebRTC v4中，移除本地轨道会自动处理
	return nil
}

// SetLocalDescription 设置本地SDP描述
func (wpc *PeerConnection) SetLocalDescription(desc webrtc.SessionDescription) error {
	return wpc.Connection.SetLocalDescription(desc)
}

// SetRemoteDescription 设置远程SDP描述
func (wpc *PeerConnection) SetRemoteDescription(desc webrtc.SessionDescription) error {
	return wpc.Connection.SetRemoteDescription(desc)
}

// AddICECandidate 添加ICE候选
func (wpc *PeerConnection) AddICECandidate(candidate webrtc.ICECandidateInit) error {
	// 存储ICE候选
	wpc.iceCandidatesMux.Lock()
	wpc.iceCandidates = append(wpc.iceCandidates, candidate)
	wpc.iceCandidatesMux.Unlock()

	return wpc.Connection.AddICECandidate(candidate)
}

// GetICECandidates 获取存储的ICE候选列表
func (wpc *PeerConnection) GetICECandidates() []webrtc.ICECandidateInit {
	wpc.iceCandidatesMux.RLock()
	defer wpc.iceCandidatesMux.RUnlock()

	// 返回副本以避免外部修改
	candidates := make([]webrtc.ICECandidateInit, len(wpc.iceCandidates))
	copy(candidates, wpc.iceCandidates)
	return candidates
}

// CreateDataChannel 创建数据通道
func (wpc *PeerConnection) CreateDataChannel(label string, options *webrtc.DataChannelInit) (*webrtc.DataChannel, error) {
	return wpc.Connection.CreateDataChannel(label, options)
}

// GetTrack 获取指定ID的媒体轨道
func (wpc *PeerConnection) GetTrack(id string) *MediaTrack {
	wpc.MediaTracksMutex.RLock()
	defer wpc.MediaTracksMutex.RUnlock()
	return wpc.MediaTracks[id]
}

// GetAllTracks 获取所有媒体轨道
func (wpc *PeerConnection) GetAllTracks() []*MediaTrack {
	wpc.MediaTracksMutex.RLock()
	tracks := make([]*MediaTrack, 0, len(wpc.MediaTracks))
	for _, track := range wpc.MediaTracks {
		tracks = append(tracks, track)
	}
	wpc.MediaTracksMutex.RUnlock()
	return tracks
}

// Close 关闭对等连接
func (wpc *PeerConnection) Close() error {
	// 关闭所有媒体轨道
	wpc.MediaTracksMutex.RLock()
	tracks := make([]*MediaTrack, 0, len(wpc.MediaTracks))
	for _, track := range wpc.MediaTracks {
		tracks = append(tracks, track)
	}
	wpc.MediaTracksMutex.RUnlock()

	for _, track := range tracks {
		wpc.RemoveTrack(track.ID)
	}

	// 关闭对等连接
	return wpc.Connection.Close()
}

// 内部方法

func (wpc *PeerConnection) onConnectionStateChange(state webrtc.PeerConnectionState) {
	// 转换状态
	peerState := PeerConnectionState(state.String())

	// 调用状态变化回调
	if wpc.OnStateChange != nil {
		wpc.OnStateChange(peerState)
	}
}

func (wpc *PeerConnection) onTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	// 调用轨道回调
	if wpc.OnTrack != nil {
		wpc.OnTrack(track, receiver)
	}
}

// getMediaType 根据类型字符串获取MediaType枚举值
func getMediaType(kind string) MediaType {
	switch kind {
	case "audio":
		return MediaTypeAudio
	case "video":
		return MediaTypeVideo
	case "data":
		return MediaTypeData
	default:
		return MediaTypeData
	}
}

// ParseICECandidate 解析ICE候选字符串
func ParseICECandidate(candidate string) (webrtc.ICECandidateInit, error) {
	bytes, err := base64.ToByte(candidate)
	if err != nil {
		return webrtc.ICECandidateInit{}, err
	}

	var iceCandidate webrtc.ICECandidateInit
	err = json.Unmarshal(bytes, &iceCandidate)
	return iceCandidate, err
}

// ParseSessionDescription 解析SDP描述
func ParseSessionDescription(sdp string) (webrtc.SessionDescription, error) {
	bytes, err := base64.ToByte(sdp)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}
	// 首先尝试解析JSON格式
	var desc webrtc.SessionDescription
	err = json.Unmarshal(bytes, &desc)
	if err == nil {
		return desc, nil
	}

	// 如果JSON解析失败，假设是纯SDP字符串
	desc = webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}
	return desc, nil
}

// CreateConfiguration 创建WebRTC配置
func CreateConfiguration(iceServers []webrtc.ICEServer) *webrtc.Configuration {
	return &webrtc.Configuration{
		ICEServers: iceServers,
	}
}
