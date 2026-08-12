package webrtc

import (
	"context"
	"errors"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

// DefaultSendChannelBuffer 默认发送通道缓冲大小
const DefaultSendChannelBuffer = 100

// rtpReaderErrorBackoff RTP 读取错误时的重试间隔，避免 CPU 空转
const rtpReaderErrorBackoff = 100 * time.Millisecond

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
	// PeerConnectionStateUnknown 连接状态：未知
	PeerConnectionStateUnknown PeerConnectionState = "unknown"
)

// PeerConnection 封装WebRTC对等连接并整合客户端连接功能
//
// @description 封装WebRTC对等连接并整合客户端连接功能。
// 该结构体同时管理 WebRTC 连接（Connection、MediaTracks、ICE 候选）和客户端通信（Send 通道）两类职责，
// 旨在简化单文件实现；若后续需要更复杂的客户端管理，可将 Send 抽离到独立结构。
// 通过 atomic.Bool 与 sync.Once 保证 Close 的并发幂等性，可在多个 goroutine 中安全调用。
// ctx 字段控制所有 RTP 读取 goroutine 的生命周期，Close 时统一取消，避免 goroutine 泄漏
// @struct
type PeerConnection struct {
	// Connection 底层的WebRTC对等连接对象
	Connection *webrtc.PeerConnection
	// ID 连接的唯一标识符
	ID string

	// MediaTracks 媒体轨道映射，存储所有媒体轨道
	// MediaTracks map[string]*MediaTrack
	MediaTracks concurrent.Map[string, *MediaTrack]

	// MediaTracksMutex 媒体轨道映射的读写锁，保证并发安全
	// MediaTracksMutex sync.RWMutex

	// OnStateChange 连接状态变化时的回调函数
	OnStateChange func(state PeerConnectionState)
	// OnTrack 收到媒体轨道时的回调函数
	OnTrack func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)
	// OnICECandidate 收到ICE候选时的回调函数
	OnICECandidate func(candidate *webrtc.ICECandidate)
	// OnICEConnectionStateChange ICE连接状态变化时的回调函数
	OnICEConnectionStateChange func(state webrtc.ICEConnectionState)
	// OnDataChannel 数据通道创建时的回调函数
	OnDataChannel func(dc *webrtc.DataChannel)

	// iceCandidates 远程ICE候选存储列表（对端发过来的候选）
	// iceCandidates []webrtc.ICECandidateInit
	iceCandidates concurrent.Queue[webrtc.ICECandidateInit]

	// localIceCandidates 本地ICE候选存储列表（pion库通过OnICECandidate回调生成的候选）
	// 与 iceCandidates 区分：iceCandidates 存储远程候选，localIceCandidates 存储本地候选
	localIceCandidates concurrent.Queue[webrtc.ICECandidateInit]

	// iceCandidatesMux ICE候选存储的读写锁，保证并发安全
	// iceCandidatesMux sync.RWMutex

	// remoteDescriptionSet 远程描述是否已设置的标志
	remoteDescriptionSet atomic.Bool
	// closed 连接是否已关闭的标志（atomic保证幂等）
	closed atomic.Bool
	// closeOnce 保证通道仅被关闭一次
	closeOnce sync.Once
	// Send 用于向客户端发送消息的通道
	Send chan []byte
	// ctx 控制所有 RTP 读取 goroutine 的生命周期
	ctx context.Context
	// cancel 取消 ctx，由 Close 调用以停止所有 RTP 读取循环
	cancel context.CancelFunc

	// rtpReadersMux 保护 rtpReaders 映射的并发访问
	// rtpReadersMux sync.Mutex

	// rtpReaders 按轨道 ID 存储停止函数，用于在 Close 时统一取消
	// rtpReaders map[string]func()
	rtpReaders concurrent.Map[string, func()]
}

// CreateOffer 创建SDP Offer
//
// @description 创建SDP Offer，委托给底层webrtc.PeerConnection
// @param options *webrtc.OfferOptions Offer选项
// @return webrtc.SessionDescription 创建的SDP描述
// @return error 如果创建过程中发生错误，则返回错误信息
func (wpc *PeerConnection) CreateOffer(options *webrtc.OfferOptions) (webrtc.SessionDescription, error) {
	return wpc.Connection.CreateOffer(options)
}

// CreateAnswer 创建SDP Answer
//
// @description 创建SDP Answer，委托给底层webrtc.PeerConnection
// @param options *webrtc.AnswerOptions Answer选项
// @return webrtc.SessionDescription 创建的SDP描述
// @return error 如果创建过程中发生错误，则返回错误信息
func (wpc *PeerConnection) CreateAnswer(options *webrtc.AnswerOptions) (webrtc.SessionDescription, error) {
	return wpc.Connection.CreateAnswer(options)
}

// AddTrack 添加媒体轨道
//
// @description 添加本地媒体轨道到底层对等连接，并注册到本地 MediaTracks 映射中
// @param track *webrtc.TrackLocalStaticRTP 要添加的本地轨道
// @param kind string 媒体类型（audio/video/data）
// @return *MediaTrack 创建的媒体轨道对象
// @return error 如果添加过程中发生错误，则返回错误信息
func (wpc *PeerConnection) AddTrack(track *webrtc.TrackLocalStaticRTP, kind string) (*MediaTrack, error) {
	// 添加轨道到对等连接
	sender, err := wpc.Connection.AddTrack(track)
	if err != nil {
		return nil, err
	}

	// 创建媒体轨道对象
	mediaTrack := &MediaTrack{
		ID:     track.ID(),
		Type:   getMediaType(kind),
		Track:  track,
		Sender: sender,
	}

	// 添加到媒体轨道映射
	// wpc.MediaTracksMutex.Lock()
	// wpc.MediaTracks[track.ID()] = mediaTrack
	// wpc.MediaTracksMutex.Unlock()
	wpc.MediaTracks.Put(track.ID(), mediaTrack)

	return mediaTrack, nil
}

// RemoveTrack 移除媒体轨道
//
// @description 移除指定ID的媒体轨道，同时从底层 WebRTC 连接中移除对应的 RTPSender，
// 确保对端不再接收该轨道的媒体流
// @param id string 媒体轨道ID
// @return error 如果轨道不存在或移除过程中发生错误，则返回错误信息
func (wpc *PeerConnection) RemoveTrack(id string) error {
	// wpc.MediaTracksMutex.Lock()
	// mediaTrack, exists := wpc.MediaTracks[id]
	// if exists {
	// 	delete(wpc.MediaTracks, id)
	// }
	// wpc.MediaTracksMutex.Unlock()

	mediaTrack, exists := wpc.MediaTracks.Get(id)
	if !exists {
		return ErrMediaTrackNotFound
	}

	// 从底层 WebRTC 连接移除 RTPSender，真正停止媒体流
	if mediaTrack.Sender != nil {
		if err := wpc.Connection.RemoveTrack(mediaTrack.Sender); err != nil {
			return err
		}
	}

	return nil
}

// SetLocalDescription 设置本地SDP描述
//
// @description 设置本地SDP描述，委托给底层webrtc.PeerConnection
// @param desc webrtc.SessionDescription 本地SDP描述
// @return error 如果设置过程中发生错误，则返回错误信息
func (wpc *PeerConnection) SetLocalDescription(desc webrtc.SessionDescription) error {
	return wpc.Connection.SetLocalDescription(desc)
}

// SetRemoteDescription 设置远程SDP描述
//
// @description 设置远程SDP描述，设置成功后标记远程描述已设置，并添加此前缓存的ICE候选
// @param desc webrtc.SessionDescription 远程SDP描述
// @return error 如果设置过程中发生错误，则返回错误信息
func (wpc *PeerConnection) SetRemoteDescription(desc webrtc.SessionDescription) error {
	err := wpc.Connection.SetRemoteDescription(desc)
	if err == nil {
		// 标记远程描述已设置
		wpc.remoteDescriptionSet.Store(true)

		// 添加存储的ICE候选
		if err := wpc.AddStoredICECandidates(); err != nil {
			return err
		}
	}
	return err
}

// AddICECandidate 添加ICE候选 (接受ICECandidate对象)
//
// @description 添加ICE候选，若远程描述尚未设置则缓存至本地列表，待 SetRemoteDescription 后统一添加
// @param candidate *webrtc.ICECandidate ICE候选对象
// @return error 如果添加过程中发生错误，则返回错误信息
func (wpc *PeerConnection) AddICECandidate(candidate *webrtc.ICECandidate) error {
	if candidate == nil {
		return nil
	}
	return wpc.AddICECandidateInit(candidate.ToJSON())
}

// AddICECandidateInit 添加ICE候选 (接受ICECandidateInit对象)
//
// @description 添加ICE候选，若远程描述尚未设置则缓存至本地列表，待 SetRemoteDescription 后统一添加
// @param candidate webrtc.ICECandidateInit ICE候选初始化对象
// @return error 如果添加过程中发生错误，则返回错误信息
func (wpc *PeerConnection) AddICECandidateInit(candidate webrtc.ICECandidateInit) error {
	// 存储ICE候选
	// wpc.iceCandidatesMux.Lock()
	// wpc.iceCandidates = append(wpc.iceCandidates, candidate)
	// wpc.iceCandidatesMux.Unlock()
	wpc.iceCandidates.Enqueue(candidate)

	// 检查远程描述是否已设置
	if wpc.remoteDescriptionSet.Load() {
		return wpc.Connection.AddICECandidate(candidate)
	}

	// 否则只存储，不添加
	return nil
}

// AddStoredICECandidates 添加存储的ICE候选
//
// @description 在远程描述设置完成后，将本地缓存的ICE候选统一添加到底层对等连接
// @return error 如果添加过程中发生错误，则返回错误信息
func (wpc *PeerConnection) AddStoredICECandidates() error {
	// wpc.iceCandidatesMux.Lock()
	// defer wpc.iceCandidatesMux.Unlock()

	// 检查远程描述是否已设置
	if !wpc.remoteDescriptionSet.Load() {
		return nil
	}

	// for _, candidate := range wpc.iceCandidates {
	// 	if err := wpc.Connection.AddICECandidate(candidate); err != nil {
	// 		return err
	// 	}
	// }

	iterator := wpc.iceCandidates.Iterator()
	defer func(iterator concurrent.Iterator[webrtc.ICECandidateInit]) {
		_ = iterator.Close()
	}(iterator)
	for iterator.HasNext() {
		candidate := iterator.Value()
		if err := wpc.Connection.AddICECandidate(candidate); err != nil {
			return err
		}
	}

	return nil
}

// GetICECandidates 获取存储的ICE候选列表
//
// @description 获取本地缓存的ICE候选列表（过滤空候选），返回副本以避免外部修改
// @return []webrtc.ICECandidateInit ICE候选列表副本
func (wpc *PeerConnection) GetICECandidates() []webrtc.ICECandidateInit {
	// wpc.iceCandidatesMux.RLock()
	// defer wpc.iceCandidatesMux.RUnlock()

	// 过滤掉空候选
	// validCandidates := make([]webrtc.ICECandidateInit, 0, len(wpc.iceCandidates))
	// for _, candidate := range wpc.iceCandidates {
	// 	if candidate.Candidate != "" {
	// 		validCandidates = append(validCandidates, candidate)
	// 	}
	// }
	validCandidates := make([]webrtc.ICECandidateInit, 0, wpc.iceCandidates.Len())
	iterator := wpc.iceCandidates.Iterator()
	defer func(iterator concurrent.Iterator[webrtc.ICECandidateInit]) {
		_ = iterator.Close()
	}(iterator)
	for iterator.HasNext() {
		candidate := iterator.Value()
		if candidate.Candidate != "" {
			validCandidates = append(validCandidates, candidate)
		}
	}

	// 返回副本以避免外部修改
	candidates := make([]webrtc.ICECandidateInit, len(validCandidates))
	copy(candidates, validCandidates)
	return candidates
}

// GetLocalICECandidates 获取本地生成的ICE候选列表
//
// @description 获取本地通过 OnICECandidate 回调生成的 ICE 候选列表（过滤空候选），
// 返回副本以避免外部修改。用于向对端返回本端的 ICE 候选
// @return []webrtc.ICECandidateInit 本地ICE候选列表副本
func (wpc *PeerConnection) GetLocalICECandidates() []webrtc.ICECandidateInit {
	validCandidates := make([]webrtc.ICECandidateInit, 0, wpc.localIceCandidates.Len())
	iterator := wpc.localIceCandidates.Iterator()
	defer func(iterator concurrent.Iterator[webrtc.ICECandidateInit]) {
		_ = iterator.Close()
	}(iterator)
	for iterator.HasNext() {
		candidate := iterator.Value()
		if candidate.Candidate != "" {
			validCandidates = append(validCandidates, candidate)
		}
	}

	// 返回副本以避免外部修改
	candidates := make([]webrtc.ICECandidateInit, len(validCandidates))
	copy(candidates, validCandidates)
	return candidates
}

// AddLocalICECandidate 存储本地生成的ICE候选
//
// @description 将 pion 库通过 OnICECandidate 回调生成的本地 ICE 候选存储到 localIceCandidates 队列，
// 供 GetLocalICECandidates 查询使用
// @param candidate webrtc.ICECandidateInit 本地ICE候选
func (wpc *PeerConnection) AddLocalICECandidate(candidate webrtc.ICECandidateInit) {
	wpc.localIceCandidates.Enqueue(candidate)
}

// CreateDataChannel 创建数据通道
//
// @description 创建数据通道，委托给底层webrtc.PeerConnection
// @param label string 数据通道标签
// @param options *webrtc.DataChannelInit 数据通道初始化选项
// @return *webrtc.DataChannel 创建的数据通道
// @return error 如果创建过程中发生错误，则返回错误信息
func (wpc *PeerConnection) CreateDataChannel(label string, options *webrtc.DataChannelInit) (*webrtc.DataChannel, error) {
	return wpc.Connection.CreateDataChannel(label, options)
}

// GetTrack 获取指定ID的媒体轨道
//
// @description 从本地 MediaTracks 映射中获取指定ID的媒体轨道
// @param id string 媒体轨道ID
// @return *MediaTrack 如果找到则返回该媒体轨道实例，否则返回nil
func (wpc *PeerConnection) GetTrack(id string) *MediaTrack {
	// wpc.MediaTracksMutex.RLock()
	// defer wpc.MediaTracksMutex.RUnlock()
	// return wpc.MediaTracks[id]
	track, exists := wpc.MediaTracks.Get(id)
	if !exists {
		return nil
	}
	return track
}

// GetAllTracks 获取所有媒体轨道
//
// @description 获取本地所有媒体轨道列表的副本
// @return []*MediaTrack 所有媒体轨道列表
func (wpc *PeerConnection) GetAllTracks() []*MediaTrack {
	// wpc.MediaTracksMutex.RLock()
	// defer wpc.MediaTracksMutex.RUnlock()
	// tracks := make([]*MediaTrack, 0, len(wpc.MediaTracks))
	// for _, track := range wpc.MediaTracks {
	// 	tracks = append(tracks, track)
	// }
	// return tracks

	mediaTracks := wpc.MediaTracks.Values()
	return mediaTracks
}

// StartRTPReader 启动 RTP 包读取循环
//
// @description 启动一个 goroutine 持续从指定 TrackRemote 读取 RTP 包，保证 pion 内部
// 正常发送 RTCP RR（Receiver Report），避免对端因收不到反馈而触发拥塞控制导致视频冻结。
// 读取到的包通过 onPacket 回调上抛；若仅需保活可传 nil。
// 读取循环受 PeerConnection.ctx 控制，在 Close 时统一取消，避免 goroutine 泄漏。
// 同一 track 重复调用会返回已注册的停止函数，不会重复启动
// @param track *webrtc.TrackRemote 远端媒体轨道
// @param onPacket func(*rtp.Packet) 收到 RTP 包时的回调，可为 nil
// @return func() 停止函数，调用以单独终止该轨道的读取循环
func (wpc *PeerConnection) StartRTPReader(track *webrtc.TrackRemote, onPacket func(*rtp.Packet)) func() {
	trackID := track.ID()

	// 同一轨道已有读取循环，直接返回已有停止函数
	// wpc.rtpReadersMux.Lock()
	// if stop, ok := wpc.rtpReaders[trackID]; ok {
	// 	wpc.rtpReadersMux.Unlock()
	// 	return stop
	// }

	if stop, ok := wpc.rtpReaders.Get(trackID); ok {
		return stop
	}

	readerCtx, readerCancel := context.WithCancel(wpc.ctx)
	stopped := make(chan struct{})

	// 注册停止函数
	stop := func() {
		readerCancel()
		<-stopped
	}
	// wpc.rtpReaders[trackID] = stop
	// wpc.rtpReadersMux.Unlock()

	wpc.rtpReaders.Put(trackID, stop)

	go func() {
		defer close(stopped)
		defer func() {
			// 从映射中清理
			// wpc.rtpReadersMux.Lock()
			// delete(wpc.rtpReaders, trackID)
			// wpc.rtpReadersMux.Unlock()
			wpc.rtpReaders.Del(trackID)
		}()

		pkt := &rtp.Packet{}
		for {
			select {
			case <-readerCtx.Done():
				return
			default:
			}

			// pion/webrtc v4 的 ReadRTP 不接受参数，返回新分配的 *rtp.Packet
			// 这里复用 pkt 变量以避免每包分配，但需通过返回值获取数据
			readPkt, _, err := track.ReadRTP()
			if err != nil {
				// EOF 或连接关闭，正常退出
				if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
					return
				}
				// ctx 取消导致的错误，正常退出
				if readerCtx.Err() != nil {
					return
				}
				// 其他错误：短暂等待后重试，避免 CPU 空转
				select {
				case <-readerCtx.Done():
					return
				case <-time.After(rtpReaderErrorBackoff):
				}
				continue
			}
			pkt = readPkt

			// 回调上抛包
			if onPacket != nil {
				onPacket(pkt)
			}
		}
	}()

	return stop
}

// Close 关闭对等连接
//
// @description 关闭对等连接，使用 atomic.Bool 保证并发幂等性，避免 TOCTOU 竞态。
// 依次取消所有 RTP 读取 goroutine、清理媒体轨道映射、ICE候选缓存与远程描述状态，
// 关闭底层 WebRTC 连接，并通过 sync.Once 安全关闭 Send 通道
// @return error 如果关闭底层连接过程中发生错误，则返回错误信息
func (wpc *PeerConnection) Close() error {
	// 使用 CAS 保证仅第一个调用者执行关闭逻辑，避免 TOCTOU 竞态
	if !wpc.closed.CompareAndSwap(false, true) {
		return nil
	}

	// 取消 ctx，停止所有 RTP 读取 goroutine
	if wpc.cancel != nil {
		wpc.cancel()
	}

	// 安全关闭 Send 通道（仅一次）
	wpc.closeOnce.Do(func() {
		close(wpc.Send)
	})

	// 关闭所有媒体轨道
	// wpc.MediaTracksMutex.Lock()
	// tracks := make([]*MediaTrack, 0, len(wpc.MediaTracks))
	// for _, track := range wpc.MediaTracks {
	// 	tracks = append(tracks, track)
	// }
	tracks := wpc.MediaTracks.Values()

	// 清空媒体轨道映射
	// wpc.MediaTracks = make(map[string]*MediaTrack)
	// wpc.MediaTracksMutex.Unlock()
	wpc.MediaTracks.Clear()

	// 从底层连接移除所有 RTPSender
	for _, track := range tracks {
		if track.Sender != nil {
			_ = wpc.Connection.RemoveTrack(track.Sender)
		}
	}

	// 清空ICE候选
	// wpc.iceCandidatesMux.Lock()
	// wpc.iceCandidates = make([]webrtc.ICECandidateInit, 0)
	// wpc.iceCandidatesMux.Unlock()
	wpc.iceCandidates.Clear()

	// 清空本地ICE候选
	wpc.localIceCandidates.Clear()

	// 重置远程描述状态
	wpc.remoteDescriptionSet.Store(false)

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
//
// @description 返回一个在ICE收集完成或连接关闭时被关闭的通道，用于等待ICE候选收集完成
// @return <-chan struct{} ICE收集完成时关闭的通道
func (wpc *PeerConnection) GatheringCompletePromise() <-chan struct{} {
	return webrtc.GatheringCompletePromise(wpc.Connection)
}
