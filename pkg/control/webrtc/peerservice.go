package webrtc

import (
	"context"
	"fmt"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	concurrentqueue "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/queue"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/pion/ice/v4"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

// DefaultSendMessageTimeout 默认发送消息超时时长
const DefaultSendMessageTimeout = 1 * time.Second

type PeerService interface {
	CreatePeerConnection(connConfig *webrtc.Configuration, id string) (*PeerConnection, error)
	GetPeerConnection(id string) *PeerConnection
	GetAllPeerConnections() []*PeerConnection
	ClosePeerConnection(id string) error
	SendTo(id string, message []byte) error
	OnPeerConnectionCreated(callback func(id string, conn *PeerConnection))
	OnPeerConnectionClosed(callback func(id string))
	OnPeerConnectionFailed(callback func(id string, err error))
	OnPeerConnectionStateChange(callback func(id string, state PeerConnectionState))
	OnTrack(callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver))
	OnICECandidate(callback func(id string, candidate *webrtc.ICECandidate))
	OnICEConnectionStateChange(callback func(id string, state webrtc.ICEConnectionState))
	OnDataChannel(callback func(id string, dc *webrtc.DataChannel))
}

// peerServiceImpl 管理WebRTC对等连接的服务
//
// @description 管理所有WebRTC对等连接的创建、维护和关闭，并管理各种回调函数。
// 所有 onXXX 回调字段通过 atomic.Pointer 存储，可在运行期安全切换；
// 对等连接映射使用 concurrent.Map 保证并发安全
// @struct
type peerServiceImpl struct {
	// ctx 上下文，用于控制生命周期
	ctx context.Context
	// logger 日志记录器，用于记录WebRTC相关日志
	logger *zap.SugaredLogger
	// config WebRTC配置信息
	config *config.WebRTCConfig
	// api webrtc.API 复用实例
	api *webrtc.API
	// peerConfig WebRTC连接配置
	peerConfig *webrtc.Configuration
	// peerConnections 对等连接映射，存储所有活跃的WebRTC连接
	peerConnections concurrent.Map[string, *PeerConnection]
	// callbacks 回调函数集合（使用atomic.Pointer保证并发安全切换）
	callbacks atomic.Pointer[peerServiceCallbacks]
}

// peerServiceCallbacks 封装 peerServiceImpl 的所有回调函数
//
// @description 封装 peerServiceImpl 的所有回调函数，整体通过 atomic.Pointer 存储与替换，
// 避免单个字段并发读写造成的数据竞争
// @struct
type peerServiceCallbacks struct {
	// onPeerConnectionCreated 对等连接创建时的回调函数
	onPeerConnectionCreated func(id string, conn *PeerConnection)
	// onPeerConnectionClosed 对等连接关闭时的回调函数
	onPeerConnectionClosed func(id string)
	// onPeerConnectionFailed 对等连接失败时的回调函数
	onPeerConnectionFailed func(id string, err error)
	// onPeerConnectionStateChange 对等连接状态变化时的回调函数
	onPeerConnectionStateChange func(id string, state PeerConnectionState)
	// onTrack 收到媒体轨道时的回调函数
	onTrack func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)
	// onICECandidate 收到ICE候选时的回调函数
	onICECandidate func(id string, candidate *webrtc.ICECandidate)
	// onICEConnectionStateChange ICE连接状态变化时的回调函数
	onICEConnectionStateChange func(id string, state webrtc.ICEConnectionState)
	// onDataChannel 数据通道创建时的回调函数
	onDataChannel func(id string, dc *webrtc.DataChannel)
}

// newPeerService 创建一个新的WebRTC对等连接服务
//
// @description 创建并初始化一个新的WebRTC对等连接服务实例
// @param ctx context.Context 上下文，用于控制生命周期
// @param config *config.WebRTCConfig WebRTC配置信息
// @param loggerControl *logger.Control 日志控制器
// @return *peerServiceImpl 创建的对等连接服务实例
// @return error 如果创建过程中发生错误，则返回错误信息
func newPeerService(ctx context.Context,
	config *config.WebRTCConfig,
	loggerControl *logger.Control,
) (PeerService, error) {
	// log := loggerControl.GenLogger(logger.ModuleWebRTC)
	log := loggerControl.GenLogger("")
	log.Debugf("[webrtc/peerservice] starting generate peer service...")

	ps := &peerServiceImpl{
		ctx:             ctx,
		config:          config,
		logger:          log,
		peerConnections: concurrentmap.NewRWMap[string, *PeerConnection](),
		// peerConnections: make(map[string]*PeerConnection),
	}

	// 生成 webrtc.API 实例
	if err := ps.initWebRTCAPI(loggerControl.GetConfig()); err != nil {
		ps.logger.Errorf("[webrtc/peerservice] generate webrtc api failed: %v", err)
		return nil, err
	}

	// 设置默认的回调函数
	ps.setupDefaultCallbacks()

	log.Infof("[webrtc/peerservice] created WebRTC peer service")
	return ps, nil
}

// setupDefaultCallbacks 设置默认的回调函数
//
// @description 设置默认的回调函数，这些回调函数提供基本的日志记录和错误处理；
// 整体替换 callbacks 指针，保证并发安全
func (ps *peerServiceImpl) setupDefaultCallbacks() {
	// 默认的连接创建回调
	cb := &peerServiceCallbacks{
		onPeerConnectionCreated: func(id string, conn *PeerConnection) {
			ps.logger.Debugf("[webrtc/peerservice] peer connection created, id: %s", id)
		},
		onPeerConnectionClosed: func(id string) {
			ps.logger.Debugf("[webrtc/peerservice] peer connection closed, id: %s", id)
		},
		onPeerConnectionFailed: func(id string, err error) {
			ps.logger.Errorf("[webrtc/peerservice] peer connection failed, id: %s, error: %v", id, err)
		},
		onPeerConnectionStateChange: func(id string, state PeerConnectionState) {
			ps.logger.Debugf("[webrtc/peerservice] peer connection state changed, id: %s, state: %s", id, state)
		},
		onTrack: func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			ps.logger.Debugf("[webrtc/peerservice] received track, peer connection id: %s, track id: %s", id, track.ID())
		},
		onICECandidate: func(id string, candidate *webrtc.ICECandidate) {
			ps.logger.Debugf("[webrtc/peerservice] received ICE candidate, peer connection id: %s", id)
		},
		onICEConnectionStateChange: func(id string, state webrtc.ICEConnectionState) {
			ps.logger.Debugf("[webrtc/peerservice] ICE connection state changed, peer connection id: %s, state: %s", id, state.String())
		},
		onDataChannel: func(id string, dc *webrtc.DataChannel) {
			ps.logger.Debugf("[webrtc/peerservice] data channel created, peer connection id: %s, channel label: %s", id, dc.Label())
		},
	}
	ps.callbacks.Store(cb)
}

// getCallbacks 厷取当前回调函数集合的快照
//
// @description 原子读取当前回调函数集合，调用方在调用具体回调前需做 nil 判断
// @return *peerServiceCallbacks 当前回调函数集合
func (ps *peerServiceImpl) getCallbacks() *peerServiceCallbacks {
	return ps.callbacks.Load()
}

// createWebRTCConfiguration 统一创建WebRTC配置
//
// @description 统一创建WebRTC配置，整合之前分离的配置创建逻辑，避免重复代码并使用正确的WebRTCConfig字段名称
func (ps *peerServiceImpl) createWebRTCConfiguration() {
	peerConfiguration := webrtc.Configuration{}
	var iceServers []webrtc.ICEServer

	// 使用新的ICEServers结构，每个服务器可以有自己的用户名和密码
	for _, server := range ps.config.ICEServers {
		iceServer := webrtc.ICEServer{
			URLs: []string{server.URL},
		}

		// 如果有用户名和密码，则设置
		if server.Username != "" {
			iceServer.Username = server.Username
			iceServer.Credential = server.Password
			iceServer.CredentialType = webrtc.ICECredentialTypePassword
		}

		iceServers = append(iceServers, iceServer)
	}

	peerConfiguration.ICEServers = iceServers
	peerConfiguration.ICETransportPolicy = webrtc.ICETransportPolicyAll
	peerConfiguration.BundlePolicy = webrtc.BundlePolicyMaxBundle
	peerConfiguration.RTCPMuxPolicy = webrtc.RTCPMuxPolicyRequire
	ps.peerConfig = &peerConfiguration
}

// initWebRTCAPI 初始化WebRTC API
//
// @description 初始化WebRTC API，包括设置日志记录器、端口范围、媒体引擎和拦截器等
// @param loggerConfig *config.LoggerConfig 日志配置信息
// @return error 如果初始化过程中发生错误，则返回错误信息
func (ps *peerServiceImpl) initWebRTCAPI(loggerConfig *config.LoggerConfig) error {
	// 使用统一的配置创建方法
	ps.createWebRTCConfiguration()

	settingEngine := webrtc.SettingEngine{
		LoggerFactory: newWebRTCLogger(loggerConfig, ps.config.LogInConsole),
	}

	if err := settingEngine.SetEphemeralUDPPortRange(uint16(ps.config.MinPort), uint16(ps.config.MaxPort)); err != nil {
		ps.logger.Errorf("[webrtc/peerservice] set ephemeral udp port range failed: %v", err)
		return err
	}

	// 禁用mDNS：服务端WebRTC不需要mDNS隐私保护（mDNS主要用于浏览器隐藏本地IP），
	// 且pion/mdns库在Windows上有已知的SetControlMessage和wsarecvfrom缓冲区警告
	if ps.config.DisableMulticastDNS {
		settingEngine.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
		ps.logger.Debugf("[webrtc/peerservice] multicast DNS disabled")
	}

	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		ps.logger.Errorf("[webrtc/peerservice] register default codecs failed: %v", err)
		return err
	}

	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	interceptorRegistry := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		ps.logger.Errorf("[webrtc/peerservice] register default interceptors failed: %v", err)
		return err
	}

	// Register a intervalpli factory
	// This interceptor sends a PLI every 3 seconds. A PLI causes a video keyframe to be generated by the sender.
	// This makes our video seekable and more error resilent, but at a cost of lower picture quality and higher bitrates
	// A real world application should process incoming RTCP packets from viewers and forward them to senders
	intervalPliFactory, err := intervalpli.NewReceiverInterceptor()
	if err != nil {
		return err
	}
	interceptorRegistry.Add(intervalPliFactory)

	ps.api = webrtc.NewAPI(
		webrtc.WithSettingEngine(settingEngine),
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry),
	)

	return nil
}

// CreatePeerConnection 创建一个新的WebRTC对等连接
//
// @description 创建一个新的WebRTC对等连接
// @param connConfig *webrtc.Configuration WebRTC连接配置，如果为nil则使用默认配置
// @param id string 连接ID
// @return *PeerConnection 创建的对等连接实例
// @return error 如果创建过程中发生错误，则返回错误信息
func (ps *peerServiceImpl) CreatePeerConnection(connConfig *webrtc.Configuration, id string) (*PeerConnection, error) {
	ps.logger.Debugf("[webrtc/peerservice] creating new peer connection, id: %s", id)

	// 记录当前活跃的连接数
	// ps.peerConnectionsMux.RLock()
	// activeConnections := len(ps.peerConnections)
	// ps.peerConnectionsMux.RUnlock()
	activeConnections := ps.peerConnections.Len()
	ps.logger.Debugf("[webrtc/peerservice] current active connections: %d", activeConnections)

	// 如果没有提供配置，使用默认配置
	if connConfig == nil {
		connConfig = ps.peerConfig
	}

	// 创建PeerConnection实例，确保所有字段都被正确初始化
	// 派生子 ctx：当 PeerConnection.Close 时仅停止该连接的 RTP 读取 goroutine，
	// 而不影响 peerServiceImpl 主 ctx 控制的其他连接
	pcCtx, pcCancel := context.WithCancel(ps.ctx)
	wpc := &PeerConnection{
		ID:                 id,
		MediaTracks:        concurrentmap.NewRWMap[string, *MediaTrack](),
		Send:               make(chan []byte, DefaultSendChannelBuffer),
		iceCandidates:      concurrentqueue.NewRWQueue[webrtc.ICECandidateInit](),
		localIceCandidates: concurrentqueue.NewRWQueue[webrtc.ICECandidateInit](),
		ctx:                pcCtx,
		cancel:             pcCancel,
		rtpReaders:         concurrentmap.NewRWMap[string, func()](),
		// MediaTracks:   make(map[string]*MediaTrack),
		// iceCandidates: make([]webrtc.ICECandidateInit, 0),
		// rtpReaders:    make(map[string]func()),
	}

	var err error
	wpc.Connection, err = ps.api.NewPeerConnection(*connConfig)
	if err != nil {
		ps.logger.Errorf("[webrtc/peerservice] failed to create peer connection: %v", err)
		return nil, err
	}

	// 设置各种回调函数
	wpc.OnStateChange = func(state PeerConnectionState) {
		ps.handleConnectionStateChange(wpc, state)
	}

	wpc.OnTrack = func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		ps.handleTrack(wpc, track, receiver)
	}

	wpc.OnICECandidate = func(candidate *webrtc.ICECandidate) {
		ps.handleICECandidate(wpc, candidate)
	}

	wpc.OnICEConnectionStateChange = func(state webrtc.ICEConnectionState) {
		ps.handleICEConnectionStateChange(wpc, state)
	}

	wpc.OnDataChannel = func(dc *webrtc.DataChannel) {
		ps.handleDataChannel(wpc, dc)
	}

	// 设置WebRTC库的回调函数
	wpc.Connection.OnDataChannel(func(dc *webrtc.DataChannel) {
		ps.logger.Debugf("[webrtc/peerservice] WebRTC library OnDataChannel called, peer connection id: %s, channel label: %s", wpc.ID, dc.Label())

		// 调用我们自己的OnDataChannel回调
		if wpc.OnDataChannel != nil {
			wpc.OnDataChannel(dc)
		}
	})

	wpc.Connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		ps.logger.Debugf("[webrtc/peerservice] WebRTC library OnConnectionStateChange called, peer connection id: %s, state: %s", wpc.ID, state.String())

		// 转换状态并调用我们自己的回调
		if wpc.OnStateChange != nil {
			peerState := convertPeerConnectionState(state)
			wpc.OnStateChange(peerState)
		}
	})

	wpc.Connection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		ps.logger.Debugf("[webrtc/peerservice] WebRTC library OnTrack called, peer connection id: %s, track id: %s", wpc.ID, track.ID())

		// 调用我们自己的OnTrack回调
		if wpc.OnTrack != nil {
			wpc.OnTrack(track, receiver)
		}
	})

	wpc.Connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		ps.logger.Debugf("[webrtc/peerservice] WebRTC library OnICECandidate called, peer connection id: %s", wpc.ID)

		// 调用我们自己的OnICECandidate回调
		if wpc.OnICECandidate != nil {
			wpc.OnICECandidate(candidate)
		}
	})

	wpc.Connection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		ps.logger.Debugf("[webrtc/peerservice] WebRTC library OnICEConnectionStateChange called, peer connection id: %s, state: %s", wpc.ID, state.String())

		// 调用我们自己的OnICEConnectionStateChange回调
		if wpc.OnICEConnectionStateChange != nil {
			wpc.OnICEConnectionStateChange(state)
		}
	})

	// 添加到对等连接映射
	// ps.peerConnectionsMux.Lock()
	// ps.peerConnections[id] = wpc
	// ps.peerConnectionsMux.Unlock()
	ps.peerConnections.Put(id, wpc)

	// 调用创建回调（受 ctx 控制，避免在服务关闭后回调泄漏）
	cb := ps.getCallbacks()
	if cb != nil && cb.onPeerConnectionCreated != nil {
		go ps.runCallback(func() {
			cb.onPeerConnectionCreated(id, wpc)
		})
	}

	ps.logger.Debugf("[webrtc/peerservice] peer connection created successfully, id: %s", id)
	return wpc, nil
}

// runCallback 在受 ctx 控制的 goroutine 中执行回调
//
// @description 在 goroutine 中执行回调，若 ctx 已取消则直接返回，避免服务关闭后回调泄漏
// @param fn func() 要执行的回调函数
func (ps *peerServiceImpl) runCallback(fn func()) {
	go func() {
		select {
		case <-ps.ctx.Done():
			return
		default:
			fn()
		}
	}()
}

// GetPeerConnection 获取指定ID的对等连接
//
// @description 获取指定ID的对等连接
// @param id string 连接ID
// @return *PeerConnection 如果找到对应的连接，则返回该连接实例；否则返回nil
func (ps *peerServiceImpl) GetPeerConnection(id string) *PeerConnection {
	// ps.peerConnectionsMux.RLock()
	// defer ps.peerConnectionsMux.RUnlock()
	// conn, exists := ps.peerConnections[id]
	conn, exists := ps.peerConnections.Get(id)
	if exists {
		ps.logger.Debugf("[webrtc/peerservice] peer connection found, id: %s, state: %s", id, conn.Connection.ConnectionState().String())
		return conn
	}
	ps.logger.Debugf("[webrtc/peerservice] peer connection not found, id: %s", id)
	return nil
}

// GetAllPeerConnections 获取所有对等连接
//
// @description 获取所有对等连接
// @return []*PeerConnection 所有对等连接的列表
func (ps *peerServiceImpl) GetAllPeerConnections() []*PeerConnection {
	// ps.peerConnectionsMux.RLock()
	// defer ps.peerConnectionsMux.RUnlock()

	ps.logger.Debugf("[webrtc/peerservice] getting all peer connections, count: %d", ps.peerConnections.Len())
	return ps.peerConnections.Values()
}

// ClosePeerConnection 关闭指定ID的对等连接
//
// @description 关闭指定ID的对等连接。方法具备幂等性，可容忍重复/并发调用：
//  1. 先 Get 取到 conn 引用，再 Del 从 map 中移除，避免顺序写反导致"先删后查拿不到引用"
//  2. pion 库在收到 DTLS CloseNotify 时会主动触发 OnConnectionStateChange(closed)，
//     回调中也会调用本方法，随后浏览器又会调用 /webrtc/close，形成"双重关闭"
//  3. conn.Close() 内部使用 atomic.Bool 保证真正的关闭只执行一次
//  4. 只有在 Get 阶段就找不到连接，且未被本方法关闭过（通过状态回调判断），才返回 ErrConnectionNotFound
//
// @param id string 连接ID
// @return error 如果连接从未存在则返回 ErrConnectionNotFound，关闭过程中发生错误则返回该错误
func (ps *peerServiceImpl) ClosePeerConnection(id string) error {
	ps.logger.Debugf("[webrtc/peerservice] closing peer connection, id: %s", id)

	// 1. 先从 map 中取 conn 引用（必须先取再删，否则永远拿不到）
	conn, exists := ps.peerConnections.Get(id)
	if !exists {
		// 若不存在，说明两种情况之一：
		//   a) 重复调用：pion 的 closed 回调已先触发了本方法，已调用过 Del + Close
		//   b) 真正不存在的 id
		// 这里只记录 debug 日志，避免返回错误影响上层（对浏览器来说重复调用 /close 是正常的）
		ps.logger.Debugf("[webrtc/peerservice] peer connection already removed from map or not found, id: %s", id)
		return nil
	}

	// 2. 取到引用后立即从 map 移除（防止并发再进来拿到同一个 conn）
	ps.peerConnections.Del(id)

	// 3. 检查底层 PeerConnection 的连接状态
	if conn.Connection != nil {
		state := conn.Connection.ConnectionState()
		ps.logger.Debugf("[webrtc/peerservice] peer connection state before closing: %s, id: %s", state.String(), id)

		// 如果连接已经被 pion 关闭（收到 DTLS CloseNotify 后先触发了 closed 回调），
		// 则 conn.Close() 内部的 atomic.Bool 会直接返回成功，这里提前短路即可
		if state == webrtc.PeerConnectionStateClosed {
			ps.logger.Debugf("[webrtc/peerservice] peer connection already closed by pion library, id: %s", id)
			ps.invokeClosedCallback(id)
			return nil
		}
	}

	// 4. 真正关闭对等连接（Close 内部 atomic.Bool 保证幂等，只执行一次）
	ps.logger.Debugf("[webrtc/peerservice] closing peer connection object, id: %s", id)
	if err := conn.Close(); err != nil {
		ps.logger.Errorf("[webrtc/peerservice] failed to close peer connection: %v", err)
		ps.invokeFailedCallback(id, err)
		// 即使关闭连接出错，也继续执行回调，避免资源泄露
	}

	// 5. 调用关闭回调
	ps.invokeClosedCallback(id)

	ps.logger.Debugf("[webrtc/peerservice] peer connection closed successfully, id: %s", id)
	return nil
}

// invokeClosedCallback 调用连接关闭回调
//
// @description 在 goroutine 中安全调用连接关闭回调，受 ctx 控制
// @param id string 连接ID
func (ps *peerServiceImpl) invokeClosedCallback(id string) {
	cb := ps.getCallbacks()
	if cb != nil && cb.onPeerConnectionClosed != nil {
		ps.runCallback(func() {
			cb.onPeerConnectionClosed(id)
		})
	}
}

// invokeFailedCallback 调用连接失败回调
//
// @description 在 goroutine 中安全调用连接失败回调，受 ctx 控制
// @param id string 连接ID
// @param err error 失败原因
func (ps *peerServiceImpl) invokeFailedCallback(id string, err error) {
	cb := ps.getCallbacks()
	if cb != nil && cb.onPeerConnectionFailed != nil {
		ps.runCallback(func() {
			cb.onPeerConnectionFailed(id, err)
		})
	}
}

// CloseAllPeerConnections 关闭所有对等连接
//
// @description 关闭所有对等连接
func (ps *peerServiceImpl) CloseAllPeerConnections() {
	ps.logger.Infof("[webrtc/peerservice] closing all peer connections")

	// 获取所有连接ID
	// ids := make([]string, 0)
	// ps.peerConnectionsMux.RLock()
	// for id := range ps.peerConnections {
	// 	ids = append(ids, id)
	// }
	// ps.peerConnectionsMux.RUnlock()
	ids := ps.peerConnections.Keys()
	// 关闭每个连接
	for _, id := range ids {
		_ = ps.ClosePeerConnection(id)
	}
}

// SendTo 向指定ID的客户端发送消息
//
// @description 向指定ID的客户端发送消息，会校验连接是否存在与是否已关闭
// @param id string 连接ID
// @param message []byte 要发送的消息内容
// @return error 如果连接不存在、已关闭或发送超时，则返回对应错误（使用哨兵错误支持 errors.Is）
func (ps *peerServiceImpl) SendTo(id string, message []byte) error {
	// ps.peerConnectionsMux.RLock()
	// conn, exists := ps.peerConnections[id]
	// ps.peerConnectionsMux.RUnlock()
	conn, exists := ps.peerConnections.Get(id)
	if !exists {
		return fmt.Errorf("%w: %s", ErrConnectionNotFound, id)
	}

	// 检查连接是否已关闭（同时防止 Connection 为空导致 panic）
	if conn.Connection == nil {
		return fmt.Errorf("%w: %s", ErrConnectionClosed, id)
	}

	// 检查通道是否已关闭
	select {
	case <-ps.ctx.Done():
		return fmt.Errorf("%w: %s", ErrConnectionClosed, id)
	default:
		// 检查连接是否已关闭
		if conn.Connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
			return fmt.Errorf("%w: %s", ErrConnectionClosed, id)
		}
	}

	// 发送消息到连接的Send通道
	// 使用select处理可能的阻塞情况
	select {
	case conn.Send <- message:
		// 消息成功发送
		return nil
	case <-time.After(DefaultSendMessageTimeout):
		// 发送超时
		return fmt.Errorf("%w: %s", ErrConnectionBufferFull, id)
	case <-ps.ctx.Done():
		// 上下文已取消
		return fmt.Errorf("%w: %s", ErrConnectionClosed, id)
	}
}

// OnPeerConnectionCreated 设置对等连接创建时的回调
//
// @description 设置对等连接创建时的回调，原子替换整个回调集合
// @param callback func(id string, conn *PeerConnection) 当新的对等连接创建时被调用的回调函数
func (ps *peerServiceImpl) OnPeerConnectionCreated(callback func(id string, conn *PeerConnection)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onPeerConnectionCreated = callback
		return &newCb
	})
}

// OnPeerConnectionClosed 设置对等连接关闭时的回调
//
// @description 设置对等连接关闭时的回调，原子替换整个回调集合
// @param callback func(id string) 当对等连接关闭时被调用的回调函数
func (ps *peerServiceImpl) OnPeerConnectionClosed(callback func(id string)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onPeerConnectionClosed = callback
		return &newCb
	})
}

// OnPeerConnectionFailed 设置对等连接失败时的回调
//
// @description 设置对等连接失败时的回调，原子替换整个回调集合
// @param callback func(id string, err error) 当对等连接失败时被调用的回调函数
func (ps *peerServiceImpl) OnPeerConnectionFailed(callback func(id string, err error)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onPeerConnectionFailed = callback
		return &newCb
	})
}

// OnPeerConnectionStateChange 设置对等连接状态变化时的回调
//
// @description 设置对等连接状态变化时的回调，原子替换整个回调集合
// @param callback func(id string, state PeerConnectionState) 当对等连接状态发生变化时被调用的回调函数
func (ps *peerServiceImpl) OnPeerConnectionStateChange(callback func(id string, state PeerConnectionState)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onPeerConnectionStateChange = callback
		return &newCb
	})
}

// OnTrack 设置收到媒体轨道时的回调
//
// @description 设置收到媒体轨道时的回调，原子替换整个回调集合
// @param callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) 当收到新的媒体轨道时被调用的回调函数
func (ps *peerServiceImpl) OnTrack(callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onTrack = callback
		return &newCb
	})
}

// OnICECandidate 设置收到ICE候选时的回调
//
// @description 设置收到ICE候选时的回调，原子替换整个回调集合
// @param callback func(id string, candidate *webrtc.ICECandidate) 当收到新的ICE候选时被调用的回调函数
func (ps *peerServiceImpl) OnICECandidate(callback func(id string, candidate *webrtc.ICECandidate)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onICECandidate = callback
		return &newCb
	})
}

// OnICEConnectionStateChange 设置ICE连接状态变化时的回调
//
// @description 设置ICE连接状态变化时的回调，原子替换整个回调集合
// @param callback func(id string, state webrtc.ICEConnectionState) 当ICE连接状态发生变化时被调用的回调函数
func (ps *peerServiceImpl) OnICEConnectionStateChange(callback func(id string, state webrtc.ICEConnectionState)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onICEConnectionStateChange = callback
		return &newCb
	})
}

// OnDataChannel 设置数据通道创建时的回调
//
// @description 设置数据通道创建时的回调，原子替换整个回调集合
// @param callback func(id string, dc *webrtc.DataChannel) 当新的数据通道创建时被调用的回调函数
func (ps *peerServiceImpl) OnDataChannel(callback func(id string, dc *webrtc.DataChannel)) {
	ps.updateCallbacks(func(cb *peerServiceCallbacks) *peerServiceCallbacks {
		newCb := *cb
		newCb.onDataChannel = callback
		return &newCb
	})
}

// updateCallbacks 原子更新回调集合
//
// @description 通过 CAS 循环原子更新回调集合，保证并发安全
// @param update func(*peerServiceCallbacks) *peerServiceCallbacks 更新函数，接收当前回调集合返回新集合
func (ps *peerServiceImpl) updateCallbacks(update func(*peerServiceCallbacks) *peerServiceCallbacks) {
	for {
		old := ps.callbacks.Load()
		newCb := update(old)
		if ps.callbacks.CompareAndSwap(old, newCb) {
			return
		}
	}
}

// convertPeerConnectionState 将 webrtc.PeerConnectionState 转换为 PeerConnectionState
//
// @description 显式映射 webrtc.PeerConnectionState 至 PeerConnectionState，
// 未识别的状态返回 PeerConnectionStateUnknown 并由调用方记录日志
// @param state webrtc.PeerConnectionState WebRTC库的连接状态
// @return PeerConnectionState 转换后的内部连接状态
func convertPeerConnectionState(state webrtc.PeerConnectionState) PeerConnectionState {
	switch state {
	case webrtc.PeerConnectionStateNew:
		return PeerConnectionStateNew
	case webrtc.PeerConnectionStateConnecting:
		return PeerConnectionStateConnecting
	case webrtc.PeerConnectionStateConnected:
		return PeerConnectionStateConnected
	case webrtc.PeerConnectionStateDisconnected:
		return PeerConnectionStateDisconnected
	case webrtc.PeerConnectionStateFailed:
		return PeerConnectionStateFailed
	case webrtc.PeerConnectionStateClosed:
		return PeerConnectionStateClosed
	default:
		return PeerConnectionStateUnknown
	}
}

// 内部处理方法

// handleConnectionStateChange 内部处理连接状态变化的方法
//
// @description 内部处理连接状态变化的方法
// @param conn *PeerConnection 对等连接实例
// @param state PeerConnectionState 连接状态
func (ps *peerServiceImpl) handleConnectionStateChange(conn *PeerConnection, state PeerConnectionState) {
	ps.logger.Debugf("[webrtc/peerservice] peer connection state changed, id: %s, state: %s", conn.ID, state)

	// 调用状态变化回调
	cb := ps.getCallbacks()
	if cb != nil && cb.onPeerConnectionStateChange != nil {
		cb.onPeerConnectionStateChange(conn.ID, state)
	}

	// 处理特殊状态
	switch state {
	case PeerConnectionStateConnected:
		ps.logger.Debugf("[webrtc/peerservice] peer connection connected, id: %s", conn.ID)
	case PeerConnectionStateConnecting:
		ps.logger.Debugf("[webrtc/peerservice] peer connection connecting, id: %s", conn.ID)
	case PeerConnectionStateDisconnected:
		ps.logger.Debugf("[webrtc/peerservice] peer connection disconnected, id: %s", conn.ID)
	case PeerConnectionStateNew:
		ps.logger.Debugf("[webrtc/peerservice] peer connection state is new, id: %s", conn.ID)
	case PeerConnectionStateFailed:
		// 连接失败，自动关闭
		ps.logger.Warnf("[webrtc/peerservice] peer connection failed, closing automatically, id: %s", conn.ID)
		go ps.ClosePeerConnection(conn.ID)
		ps.invokeFailedCallback(conn.ID, ErrConnectionFailed)
	case PeerConnectionStateClosed:
		ps.logger.Debugf("[webrtc/peerservice] peer connection closed, id: %s", conn.ID)
	case PeerConnectionStateUnknown:
		ps.logger.Warnf("[webrtc/peerservice] unknown peer connection state, id: %s", conn.ID)
	}
}

// handleTrack 内部处理媒体轨道的方法
//
// @description 内部处理收到媒体轨道的方法。
// 关键动作：启动 RTP 读取循环，持续消费 TrackRemote 上的 RTP 包，
// 保证 pion 内部正常发送 RTCP RR，避免对端因收不到反馈而触发拥塞控制导致视频冻结
// @param conn *PeerConnection 对等连接实例
// @param track *webrtc.TrackRemote 媒体轨道
// @param receiver *webrtc.RTPReceiver RTP接收器
func (ps *peerServiceImpl) handleTrack(conn *PeerConnection, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	ps.logger.Infof("[webrtc/peerservice] track arrived, peer connection id: %s, track id: %s, kind: %s", conn.ID, track.ID(), track.Kind().String())

	// 启动 RTP 读取循环，保证 RTCP RR 正常发送，避免对端视频冻结
	// 此处不传 onPacket 回调，仅做保活；上层若需处理 RTP 包可通过 OnTrack 回调自行读取
	conn.StartRTPReader(track, nil)

	// 调用轨道回调
	cb := ps.getCallbacks()
	if cb != nil && cb.onTrack != nil {
		cb.onTrack(conn.ID, track, receiver)
	}
}

// handleICECandidate 内部处理ICE候选的方法
//
// @description 内部处理收到本地ICE候选的方法。
// pion 库通过 OnICECandidate 回调生成的是本地ICE候选，需要存储到 localIceCandidates 中，
// 供 GetLocalICECandidates 查询并返回给对端。
// 注意：这与 AddICECandidateInit 存储的远程候选（iceCandidates）是分开的，不可混淆
// @param conn *PeerConnection 对等连接实例
// @param candidate *webrtc.ICECandidate 本地ICE候选
func (ps *peerServiceImpl) handleICECandidate(conn *PeerConnection, candidate *webrtc.ICECandidate) {
	ps.logger.Debugf("[webrtc/peerservice] received local ICE candidate, peer connection id: %s", conn.ID)

	if candidate != nil {
		// 将 *webrtc.ICECandidate 转换为 webrtc.ICECandidateInit 并存储到本地候选队列
		candidateInit := candidate.ToJSON()
		ps.logger.Debugf("[webrtc/peerservice] storing local ICE candidate: %s, peer connection id: %s", candidateInit.Candidate, conn.ID)
		conn.AddLocalICECandidate(candidateInit)

		// 调用ICE候选回调
		cb := ps.getCallbacks()
		if cb != nil && cb.onICECandidate != nil {
			cb.onICECandidate(conn.ID, candidate)
		}
	}
}

// handleICEConnectionStateChange 内部处理ICE连接状态变化的方法
//
// @description 内部处理ICE连接状态变化的方法
// @param conn *PeerConnection 对等连接实例
// @param state webrtc.ICEConnectionState ICE连接状态
func (ps *peerServiceImpl) handleICEConnectionStateChange(conn *PeerConnection, state webrtc.ICEConnectionState) {
	ps.logger.Debugf("[webrtc/peerservice] ICE connection state changed, peer connection id: %s, state: %s", conn.ID, state.String())

	// 调用ICE连接状态变化回调
	cb := ps.getCallbacks()
	if cb != nil && cb.onICEConnectionStateChange != nil {
		cb.onICEConnectionStateChange(conn.ID, state)
	}

	// 添加特殊状态处理的日志
	switch state {
	case webrtc.ICEConnectionStateFailed:
		ps.logger.Errorf("[webrtc/peerservice] ICE connection failed for peer connection: %s", conn.ID)
	case webrtc.ICEConnectionStateDisconnected:
		ps.logger.Warnf("[webrtc/peerservice] ICE connection disconnected for peer connection: %s", conn.ID)
	case webrtc.ICEConnectionStateConnected:
		ps.logger.Infof("[webrtc/peerservice] ICE connection established for peer connection: %s", conn.ID)
	case webrtc.ICEConnectionStateCompleted:
		ps.logger.Infof("[webrtc/peerservice] ICE connection completed for peer connection: %s", conn.ID)
	case webrtc.ICEConnectionStateNew, webrtc.ICEConnectionStateChecking, webrtc.ICEConnectionStateClosed:
		ps.logger.Debugf("[webrtc/peerservice] ICE connection state: %s, peer connection id: %s", state.String(), conn.ID)
	default:
		ps.logger.Warnf("[webrtc/peerservice] unhandled ICE connection state: %s", state.String())
	}
}

// handleDataChannel 内部处理数据通道的方法
//
// @description 内部处理数据通道创建的方法
// @param conn *PeerConnection 对等连接实例
// @param dc *webrtc.DataChannel 数据通道
func (ps *peerServiceImpl) handleDataChannel(conn *PeerConnection, dc *webrtc.DataChannel) {
	ps.logger.Debugf("[webrtc/peerservice] data channel created, peer connection id: %s, channel label: %s, channel id: %d", conn.ID, dc.Label(), dc.ID())

	// 添加DataChannel的事件处理
	dc.OnOpen(func() {
		ps.logger.Infof("[webrtc/peerservice] DataChannel '%s'-'%d' opened for peer connection: %s", dc.Label(), dc.ID(), conn.ID)
	})

	dc.OnClose(func() {
		ps.logger.Infof("[webrtc/peerservice] DataChannel '%s'-'%d' closed for peer connection: %s", dc.Label(), dc.ID(), conn.ID)
	})

	dc.OnError(func(err error) {
		ps.logger.Errorf("[webrtc/peerservice] DataChannel '%s'-'%d' error for peer connection %s: %v", dc.Label(), dc.ID(), conn.ID, err)
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		ps.logger.Infof("[webrtc/peerservice] Message from DataChannel '%s'-'%d' for peer connection %s: '%s'", dc.Label(), dc.ID(), conn.ID, string(msg.Data))
	})

	// 调用数据通道回调
	cb := ps.getCallbacks()
	if cb != nil && cb.onDataChannel != nil {
		cb.onDataChannel(conn.ID, dc)
	}
}
