package webrtc

import (
	"sync"

	"github.com/pion/webrtc/v4"
)

// PeerConnectionState 定义对等连接状态
type PeerConnectionState string

const (
	// PeerConnectionStateNew 连接状态：新建
	PeerConnectionStateNew PeerConnectionState = "new"
	// PeerConnectionStateConnecting 连接状态：连接中
	PeerConnectionStateConnecting PeerConnectionState = "connecting"
	// PeerConnectionStateConnected 连接状态：已连接
	PeerConnectionStateConnected PeerConnectionState = "connected"
	// PeerConnectionStateDisconnected 连接状态：已断开
	PeerConnectionStateDisconnected PeerConnectionState = "disconnected"
	// PeerConnectionStateFailed 连接状态：失败
	PeerConnectionStateFailed PeerConnectionState = "failed"
	// PeerConnectionStateClosed 连接状态：已关闭
	PeerConnectionStateClosed PeerConnectionState = "closed"
)

// PeerConnection 封装WebRTC对等连接并整合客户端连接功能
//
// @description 封装WebRTC对等连接并整合客户端连接功能，同时管理WebRTC连接和客户端通信功能
// @struct
type PeerConnection struct {
	Connection                 *webrtc.PeerConnection                                        // Connection 底层的WebRTC对等连接对象
	ID                         string                                                        // ID 连接的唯一标识符
	MediaTracks                map[string]*MediaTrack                                        // MediaTracks 媒体轨道映射，存储所有媒体轨道
	MediaTracksMutex           sync.RWMutex                                                  // MediaTracksMutex 媒体轨道映射的读写锁，保证并发安全
	OnStateChange              func(state PeerConnectionState)                               // OnStateChange 连接状态变化时的回调函数
	OnTrack                    func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) // OnTrack 收到媒体轨道时的回调函数
	OnICECandidate             func(candidate *webrtc.ICECandidate)                          // OnICECandidate 收到ICE候选时的回调函数
	OnICEConnectionStateChange func(state webrtc.ICEConnectionState)                         // OnICEConnectionStateChange ICE连接状态变化时的回调函数
	OnDataChannel              func(dc *webrtc.DataChannel)                                  // OnDataChannel 数据通道创建时的回调函数
	iceCandidates              []webrtc.ICECandidateInit                                     // iceCandidates ICE候选存储列表
	iceCandidatesMux           sync.RWMutex                                                  // iceCandidatesMux ICE候选存储的读写锁，保证并发安全
	remoteDescriptionSet       bool                                                          // remoteDescriptionSet 远程描述是否已设置的标志
	remoteDescriptionMux       sync.RWMutex                                                  // remoteDescriptionMux 远程描述状态的读写锁，保证并发安全
	closed                     bool                                                          // closed 连接是否已关闭的标志
	closeMux                   sync.RWMutex                                                  // closeMux 连接关闭状态的读写锁，保证并发安全
	Send                       chan []byte                                                   // Send 用于向客户端发送消息的通道
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
	err := wpc.Connection.SetRemoteDescription(desc)
	if err == nil {
		// 标记远程描述已设置
		wpc.remoteDescriptionMux.Lock()
		wpc.remoteDescriptionSet = true
		wpc.remoteDescriptionMux.Unlock()

		// 添加存储的ICE候选
		if err := wpc.AddStoredICECandidates(); err != nil {
			return err
		}
	}
	return err
}

// AddICECandidate 添加ICE候选 (接受ICECandidate对象)
func (wpc *PeerConnection) AddICECandidate(candidate *webrtc.ICECandidate) error {
	// 存储ICE候选
	wpc.iceCandidatesMux.Lock()
	wpc.iceCandidates = append(wpc.iceCandidates, candidate.ToJSON())
	wpc.iceCandidatesMux.Unlock()

	// 检查远程描述是否已设置
	wpc.remoteDescriptionMux.RLock()
	remoteSet := wpc.remoteDescriptionSet
	wpc.remoteDescriptionMux.RUnlock()

	// 如果远程描述已设置，则直接添加到PeerConnection
	if remoteSet {
		return wpc.Connection.AddICECandidate(candidate.ToJSON())
	}

	// 否则只存储，不添加
	return nil
}

// AddICECandidateInit 添加ICE候选 (接受ICECandidateInit对象)
func (wpc *PeerConnection) AddICECandidateInit(candidate webrtc.ICECandidateInit) error {

	// 存储ICE候选
	wpc.iceCandidatesMux.Lock()
	wpc.iceCandidates = append(wpc.iceCandidates, candidate)
	wpc.iceCandidatesMux.Unlock()

	// 检查远程描述是否已设置
	wpc.remoteDescriptionMux.RLock()
	remoteSet := wpc.remoteDescriptionSet
	wpc.remoteDescriptionMux.RUnlock()

	// 如果远程描述已设置，则直接添加到PeerConnection
	if remoteSet {
		return wpc.Connection.AddICECandidate(candidate)
	}

	// 否则只存储，不添加
	return nil
}

// AddStoredICECandidates 添加存储的ICE候选
func (wpc *PeerConnection) AddStoredICECandidates() error {
	wpc.iceCandidatesMux.Lock()
	defer wpc.iceCandidatesMux.Unlock()

	// 检查远程描述是否已设置
	wpc.remoteDescriptionMux.RLock()
	remoteSet := wpc.remoteDescriptionSet
	wpc.remoteDescriptionMux.RUnlock()

	// 如果远程描述未设置，则不添加
	if !remoteSet {
		return nil
	}

	for _, candidate := range wpc.iceCandidates {
		if err := wpc.Connection.AddICECandidate(candidate); err != nil {
			return err
		}
	}

	return nil
}

// GetICECandidates 获取存储的ICE候选列表
func (wpc *PeerConnection) GetICECandidates() []webrtc.ICECandidateInit {
	wpc.iceCandidatesMux.RLock()
	defer wpc.iceCandidatesMux.RUnlock()

	// 过滤掉空候选
	validCandidates := make([]webrtc.ICECandidateInit, 0, len(wpc.iceCandidates))
	for _, candidate := range wpc.iceCandidates {
		if candidate.Candidate != "" {
			validCandidates = append(validCandidates, candidate)
		}
	}

	// 返回副本以避免外部修改
	candidates := make([]webrtc.ICECandidateInit, len(validCandidates))
	copy(candidates, validCandidates)
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
	// 检查是否已经关闭
	wpc.closeMux.RLock()
	if wpc.closed {
		wpc.closeMux.RUnlock()
		return nil
	}
	wpc.closeMux.RUnlock()

	// 设置关闭标志
	wpc.closeMux.Lock()
	wpc.closed = true
	wpc.closeMux.Unlock()

	// 关闭所有媒体轨道
	wpc.MediaTracksMutex.Lock()
	tracks := make([]*MediaTrack, 0, len(wpc.MediaTracks))
	for _, track := range wpc.MediaTracks {
		tracks = append(tracks, track)
	}
	// 清空媒体轨道映射
	wpc.MediaTracks = make(map[string]*MediaTrack)
	wpc.MediaTracksMutex.Unlock()

	for _, track := range tracks {
		_ = wpc.RemoveTrack(track.ID)
	}

	// 清空ICE候选
	wpc.iceCandidatesMux.Lock()
	wpc.iceCandidates = make([]webrtc.ICECandidateInit, 0)
	wpc.iceCandidatesMux.Unlock()

	// 重置远程描述状态
	wpc.remoteDescriptionMux.Lock()
	wpc.remoteDescriptionSet = false
	wpc.remoteDescriptionMux.Unlock()

	// 关闭对等连接
	if wpc.Connection != nil {
		// 检查连接状态，避免重复关闭
		if wpc.Connection.ConnectionState() != webrtc.PeerConnectionStateClosed {
			return wpc.Connection.Close()
		}
	}

	return nil
}

// GatheringCompletePromise 返回一个在ICE收集完成时关闭的通道
func (wpc *PeerConnection) GatheringCompletePromise() <-chan struct{} {
	return webrtc.GatheringCompletePromise(wpc.Connection)
}
