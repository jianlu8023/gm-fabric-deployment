package webrtc

import (
	"context"
	"fmt"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

// PeerService 管理WebRTC对等连接的服务
//
// @description 管理所有WebRTC对等连接的创建、维护和关闭，并管理各种回调函数
// @struct
type PeerService struct {
	ctx        context.Context       // ctx 上下文，用于控制生命周期
	logger     *zap.SugaredLogger    // logger 日志记录器，用于记录WebRTC相关日志
	config     *config.WebRTCConfig  // config WebRTC配置信息
	api        *webrtc.API           // api webrtc.API 复用实例
	peerConfig *webrtc.Configuration // peerConfig WebRTC连接配置
	// peerConnections             map[string]*PeerConnection                 // peerConnections 对等连接映射，存储所有活跃的WebRTC连接
	// peerConnectionsMux          sync.RWMutex                               // peerConnectionsMux 对等连接映射的读写锁，保证并发安全
	peerConnections             concurrent.Map[string, *PeerConnection]    // peerConnections 对等连接映射，存储所有活跃的WebRTC连接
	onPeerConnectionCreated     func(id string, conn *PeerConnection)      // onPeerConnectionCreated 对等连接创建时的回调函数
	onPeerConnectionClosed      func(id string)                            // onPeerConnectionClosed 对等连接关闭时的回调函数
	onPeerConnectionFailed      func(id string, err error)                 // onPeerConnectionFailed 对等连接失败时的回调函数
	onPeerConnectionStateChange func(id string, state PeerConnectionState) // onPeerConnectionStateChange 对等连接状态变化时的回调函数
	onTrack                     func(id string, track *webrtc.TrackRemote,
		receiver *webrtc.RTPReceiver) // onTrack 收到媒体轨道时的回调函数
	onICECandidate             func(id string, candidate *webrtc.ICECandidate)  // onICECandidate 收到ICE候选时的回调函数
	onICEConnectionStateChange func(id string, state webrtc.ICEConnectionState) // onICEConnectionStateChange ICE连接状态变化时的回调函数
	onDataChannel              func(id string, dc *webrtc.DataChannel)          // onDataChannel 数据通道创建时的回调函数
}

// newPeerService 创建一个新的WebRTC对等连接服务
//
// @description 创建并初始化一个新的WebRTC对等连接服务实例
// @param ctx context.Context 上下文，用于控制生命周期
// @param config *config.WebRTCConfig WebRTC配置信息
// @param loggerControl *logger.Control 日志控制器
// @return *PeerService 创建的对等连接服务实例
// @return error 如果创建过程中发生错误，则返回错误信息
func newPeerService(ctx context.Context,
	config *config.WebRTCConfig,
	loggerControl *logger.Control,
) (*PeerService, error) {
	log := loggerControl.GenLogger(logger.ModuleWebRTC)
	log.Debugf("[peerservice] starting generate peer service...")

	ps := &PeerService{
		ctx:    ctx,
		config: config,
		logger: log,
		// peerConnections: make(map[string]*PeerConnection),
		peerConnections: concurrentmap.NewRWMap[string, *PeerConnection](),
	}

	// 生成 webrtc.API 实例
	if err := ps.initWebRTCAPI(loggerControl.GetConfig()); err != nil {
		ps.logger.Errorf("[peerservice] generate webrtc api failed: %v", err)
		return nil, err
	}

	// 设置默认的回调函数
	ps.setupDefaultCallbacks()

	log.Infof("[peerservice] created WebRTC peer service")
	return ps, nil
}

// setupDefaultCallbacks 设置默认的回调函数
//
// @description 设置默认的回调函数，这些回调函数提供基本的日志记录和错误处理
func (ps *PeerService) setupDefaultCallbacks() {
	// 默认的连接创建回调
	ps.onPeerConnectionCreated = func(id string, conn *PeerConnection) {
		ps.logger.Debugf("[peerservice] peer connection created, id: %s", id)
	}

	// 默认的连接关闭回调
	ps.onPeerConnectionClosed = func(id string) {
		ps.logger.Debugf("[peerservice] peer connection closed, id: %s", id)
	}

	// 默认的连接失败回调
	ps.onPeerConnectionFailed = func(id string, err error) {
		ps.logger.Errorf("[peerservice] peer connection failed, id: %s, error: %v", id, err)
	}

	// 默认的连接状态变化回调
	ps.onPeerConnectionStateChange = func(id string, state PeerConnectionState) {
		ps.logger.Debugf("[peerservice] peer connection state changed, id: %s, state: %s", id, state)
	}

	// 默认的轨道接收回调
	ps.onTrack = func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		ps.logger.Debugf("[peerservice] received track, peer connection id: %s, track id: %s", id, track.ID())
	}

	// 默认的ICE候选回调
	ps.onICECandidate = func(id string, candidate *webrtc.ICECandidate) {
		ps.logger.Debugf("[peerservice] received ICE candidate, peer connection id: %s", id)
	}

	// 默认的ICE连接状态变化回调
	ps.onICEConnectionStateChange = func(id string, state webrtc.ICEConnectionState) {
		ps.logger.Debugf("[peerservice] ICE connection state changed, peer connection id: %s, state: %s", id, state.String())
	}

	// 默认的数据通道回调
	ps.onDataChannel = func(id string, dc *webrtc.DataChannel) {
		ps.logger.Debugf("[peerservice] data channel created, peer connection id: %s, channel label: %s", id, dc.Label())
	}
}

// createWebRTCConfiguration 统一创建WebRTC配置
//
// @description 统一创建WebRTC配置，整合之前分离的配置创建逻辑，避免重复代码并使用正确的WebRTCConfig字段名称
func (ps *PeerService) createWebRTCConfiguration() {
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
func (ps *PeerService) initWebRTCAPI(loggerConfig *config.LoggerConfig) error {
	// 使用统一的配置创建方法
	ps.createWebRTCConfiguration()

	settingEngine := webrtc.SettingEngine{
		LoggerFactory: newWebRTCLogger(loggerConfig, ps.config.LogInConsole),
	}

	if err := settingEngine.SetEphemeralUDPPortRange(uint16(ps.config.MinPort), uint16(ps.config.MaxPort)); err != nil {
		ps.logger.Errorf("[peerservice] set ephemeral udp port range failed: %v", err)
		return err
	}

	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		ps.logger.Errorf("[peerservice] register default codecs failed: %v", err)
		return err
	}

	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	interceptorRegistry := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		ps.logger.Errorf("[peerservice] register default interceptors failed: %v", err)
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
func (ps *PeerService) CreatePeerConnection(connConfig *webrtc.Configuration, id string) (*PeerConnection, error) {
	ps.logger.Debugf("[peerservice] creating new peer connection, id: %s", id)

	// 记录当前活跃的连接数
	// ps.peerConnectionsMux.RLock()
	// activeConnections := len(ps.peerConnections)
	// ps.peerConnectionsMux.RUnlock()
	activeConnections := ps.peerConnections.Len()
	ps.logger.Debugf("[peerservice] current active connections: %d", activeConnections)

	// 如果没有提供配置，使用默认配置
	if connConfig == nil {
		connConfig = ps.peerConfig
	}

	// 创建PeerConnection实例，确保所有字段都被正确初始化
	wpc := &PeerConnection{
		ID:            id,
		MediaTracks:   make(map[string]*MediaTrack),
		Send:          make(chan []byte, 100),             // 初始化发送通道
		iceCandidates: make([]webrtc.ICECandidateInit, 0), // 初始化ICE候选存储
		// iceCandidatesMux:     sync.RWMutex{},                     // 初始化ICE候选互斥锁
		remoteDescriptionSet: false, // 初始化远程描述状态
		// remoteDescriptionMux: sync.RWMutex{},                     // 初始化远程描述互斥锁
		closed: false, // 初始化关闭标志
		// closeMux:             sync.RWMutex{},                     // 初始化关闭互斥锁
	}

	var err error
	wpc.Connection, err = ps.api.NewPeerConnection(*connConfig)
	if err != nil {
		ps.logger.Errorf("[peerservice] failed to create peer connection: %v", err)
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
		ps.logger.Debugf("[peerservice] WebRTC library OnDataChannel called, peer connection id: %s, channel label: %s", wpc.ID, dc.Label())

		// 调用我们自己的OnDataChannel回调
		if wpc.OnDataChannel != nil {
			wpc.OnDataChannel(dc)
		}
	})

	wpc.Connection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		ps.logger.Debugf("[peerservice] WebRTC library OnConnectionStateChange called, peer connection id: %s, state: %s", wpc.ID, state.String())

		// 转换状态并调用我们自己的回调
		if wpc.OnStateChange != nil {
			peerState := PeerConnectionState(state.String())
			wpc.OnStateChange(peerState)
		}
	})

	wpc.Connection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		ps.logger.Debugf("[peerservice] WebRTC library OnTrack called, peer connection id: %s, track id: %s", wpc.ID, track.ID())

		// 调用我们自己的OnTrack回调
		if wpc.OnTrack != nil {
			wpc.OnTrack(track, receiver)
		}
	})

	wpc.Connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		ps.logger.Debugf("[peerservice] WebRTC library OnICECandidate called, peer connection id: %s", wpc.ID)

		// 调用我们自己的OnICECandidate回调
		if wpc.OnICECandidate != nil {
			wpc.OnICECandidate(candidate)
		}
	})

	wpc.Connection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		ps.logger.Debugf("[peerservice] WebRTC library OnICEConnectionStateChange called, peer connection id: %s, state: %s", wpc.ID, state.String())

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

	// 调用创建回调
	if ps.onPeerConnectionCreated != nil {
		go ps.onPeerConnectionCreated(id, wpc)
	}

	ps.logger.Debugf("[peerservice] peer connection created successfully, id: %s", id)
	return wpc, nil
}

// removeConnection 内部方法：移除连接
//
// @description 内部方法：移除连接，现在直接操作peerConnections，不再需要单独的connections映射
// @param id string 连接ID
func (ps *PeerService) removeConnection(id string) {
	// ps.peerConnectionsMux.Lock()
	// conn, exists := ps.peerConnections[id]
	// if exists {
	// 	delete(ps.peerConnections, id)
	// }
	// ps.peerConnectionsMux.Unlock()
	conn, exists := ps.peerConnections.Get(id)
	if exists {
		ps.peerConnections.Del(id)
	}

	// 只有在连接存在时才执行清理操作
	if exists {
		// 关闭发送通道
		conn.closeMux.Lock()
		// 检查通道是否已经关闭
		defer func() {
			if r := recover(); r != nil {
				ps.logger.Warnf("[peerservice] channel already closed, id: %s", id)
			}
		}()
		close(conn.Send)
		conn.closeMux.Unlock()
	}
}

// GetPeerConnection 获取指定ID的对等连接
//
// @description 获取指定ID的对等连接
// @param id string 连接ID
// @return *PeerConnection 如果找到对应的连接，则返回该连接实例；否则返回nil
func (ps *PeerService) GetPeerConnection(id string) *PeerConnection {
	// ps.peerConnectionsMux.RLock()
	// defer ps.peerConnectionsMux.RUnlock()
	// conn, exists := ps.peerConnections[id]
	conn, exists := ps.peerConnections.Get(id)
	if exists {
		ps.logger.Debugf("[peerservice] peer connection found, id: %s, state: %s", id, conn.Connection.ConnectionState().String())
		return conn
	}
	ps.logger.Debugf("[peerservice] peer connection not found, id: %s", id)
	return nil
}

// GetAllPeerConnections 获取所有对等连接
//
// @description 获取所有对等连接
// @return []*PeerConnection 所有对等连接的列表
func (ps *PeerService) GetAllPeerConnections() []*PeerConnection {
	// ps.peerConnectionsMux.RLock()
	// defer ps.peerConnectionsMux.RUnlock()

	ps.logger.Debugf("[peerservice] getting all peer connections, count: %d", ps.peerConnections.Len())
	// connections := make([]*PeerConnection, 0, len(ps.peerConnections))
	// for id, conn := range ps.peerConnections {
	// 	connections = append(connections, conn)
	// 	ps.logger.Debugf("[peerservice] peer connection id: %s, state: %s", id, conn.Connection.ConnectionState().String())
	// }
	// return connections
	values := ps.peerConnections.Values()
	return values
}

// ClosePeerConnection 关闭指定ID的对等连接
//
// @description 关闭指定ID的对等连接
// @param id string 连接ID
// @return error 如果关闭过程中发生错误，则返回错误信息
func (ps *PeerService) ClosePeerConnection(id string) error {
	ps.logger.Debugf("[peerservice] closing peer connection, id: %s", id)

	// ps.peerConnectionsMux.Lock()
	// conn, exists := ps.peerConnections[id]
	conn, exists := ps.peerConnections.Get(id)
	if exists {
		// 从映射中移除连接
		// delete(ps.peerConnections, id)
		ps.peerConnections.Del(id)
		ps.logger.Debugf("[peerservice] peer connection removed from map, id: %s", id)
	}
	// ps.peerConnectionsMux.Unlock()

	if !exists {
		ps.logger.Warnf("[peerservice] peer connection not found, id: %s", id)
		return ErrConnectionNotFound
	}

	// 检查连接状态
	if conn.Connection != nil {
		state := conn.Connection.ConnectionState()
		ps.logger.Debugf("[peerservice] peer connection state before closing: %s, id: %s", state.String(), id)

		// 如果连接已经关闭，不需要再次关闭
		if state == webrtc.PeerConnectionStateClosed {
			ps.logger.Debugf("[peerservice] peer connection already closed, id: %s", id)
			// 调用关闭回调
			if ps.onPeerConnectionClosed != nil {
				go ps.onPeerConnectionClosed(id)
			}
			return nil
		}
	}

	// 关闭对等连接
	ps.logger.Debugf("[peerservice] closing peer connection object, id: %s", id)
	if err := conn.Close(); err != nil {
		ps.logger.Errorf("[peerservice] failed to close peer connection: %v", err)
		if ps.onPeerConnectionFailed != nil {
			go ps.onPeerConnectionFailed(id, err)
		}
		// 即使关闭连接出错，也继续执行回调
	}

	// 调用关闭回调
	if ps.onPeerConnectionClosed != nil {
		go ps.onPeerConnectionClosed(id)
	}

	ps.logger.Debugf("[peerservice] peer connection closed successfully, id: %s", id)
	return nil
}

// CloseAllPeerConnections 关闭所有对等连接
//
// @description 关闭所有对等连接
func (ps *PeerService) CloseAllPeerConnections() {
	ps.logger.Infof("[peerservice] closing all peer connections")

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
// @description 向指定ID的客户端发送消息，现在直接操作peerConnections，不再需要单独的connections映射
// @param id string 连接ID
// @param message []byte 要发送的消息内容
// @return error 如果发送过程中发生错误，则返回错误信息
func (ps *PeerService) SendTo(id string, message []byte) error {
	// ps.peerConnectionsMux.RLock()
	// conn, exists := ps.peerConnections[id]
	// ps.peerConnectionsMux.RUnlock()
	conn, exists := ps.peerConnections.Get(id)

	if !exists {
		return fmt.Errorf("[peerservice] connection not found: %s", id)
	}

	// 检查通道是否已关闭
	select {
	case <-ps.ctx.Done():
		return fmt.Errorf("[peerservice] context canceled")
	default:
		// 检查连接是否已关闭
		if conn.Connection.ConnectionState() == webrtc.PeerConnectionStateClosed {
			return fmt.Errorf("[peerservice] connection closed: %s", id)
		}
	}

	// 发送消息到连接的Send通道
	// 使用select处理可能的阻塞情况
	select {
	case conn.Send <- message:
		// 消息成功发送
		return nil
	case <-time.After(1 * time.Second):
		// 发送超时
		return fmt.Errorf("[peerservice] send message to %s timeout", id)
	case <-ps.ctx.Done():
		// 上下文已取消
		return fmt.Errorf("[peerservice] context canceled")
	}
}

// OnPeerConnectionCreated 设置对等连接创建时的回调
//
// @description 设置对等连接创建时的回调
// @param callback func(id string, conn *PeerConnection) 当新的对等连接创建时被调用的回调函数
func (ps *PeerService) OnPeerConnectionCreated(callback func(id string, conn *PeerConnection)) {
	ps.onPeerConnectionCreated = callback
}

// OnPeerConnectionClosed 设置对等连接关闭时的回调
//
// @description 设置对等连接关闭时的回调
// @param callback func(id string) 当对等连接关闭时被调用的回调函数
func (ps *PeerService) OnPeerConnectionClosed(callback func(id string)) {
	ps.onPeerConnectionClosed = callback
}

// OnPeerConnectionFailed 设置对等连接失败时的回调
//
// @description 设置对等连接失败时的回调
// @param callback func(id string, err error) 当对等连接失败时被调用的回调函数
func (ps *PeerService) OnPeerConnectionFailed(callback func(id string, err error)) {
	ps.onPeerConnectionFailed = callback
}

// OnPeerConnectionStateChange 设置对等连接状态变化时的回调
//
// @description 设置对等连接状态变化时的回调
// @param callback func(id string, state PeerConnectionState) 当对等连接状态发生变化时被调用的回调函数
func (ps *PeerService) OnPeerConnectionStateChange(callback func(id string, state PeerConnectionState)) {
	ps.onPeerConnectionStateChange = callback
}

// OnTrack 设置收到媒体轨道时的回调
//
// @description 设置收到媒体轨道时的回调
// @param callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) 当收到新的媒体轨道时被调用的回调函数
func (ps *PeerService) OnTrack(callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)) {
	ps.onTrack = callback
}

// OnICECandidate 设置收到ICE候选时的回调
//
// @description 设置收到ICE候选时的回调
// @param callback func(id string, candidate *webrtc.ICECandidate) 当收到新的ICE候选时被调用的回调函数
func (ps *PeerService) OnICECandidate(callback func(id string, candidate *webrtc.ICECandidate)) {
	ps.onICECandidate = callback
}

// OnICEConnectionStateChange 设置ICE连接状态变化时的回调
//
// @description 设置ICE连接状态变化时的回调
// @param callback func(id string, state webrtc.ICEConnectionState) 当ICE连接状态发生变化时被调用的回调函数
func (ps *PeerService) OnICEConnectionStateChange(callback func(id string, state webrtc.ICEConnectionState)) {
	ps.onICEConnectionStateChange = callback
}

// OnDataChannel 设置数据通道创建时的回调
//
// @description 设置数据通道创建时的回调
// @param callback func(id string, dc *webrtc.DataChannel) 当新的数据通道创建时被调用的回调函数
func (ps *PeerService) OnDataChannel(callback func(id string, dc *webrtc.DataChannel)) {
	ps.onDataChannel = callback
}

// 内部处理方法

// handleConnectionStateChange 内部处理连接状态变化的方法
//
// @description 内部处理连接状态变化的方法
// @param conn *PeerConnection 对等连接实例
// @param state PeerConnectionState 连接状态
func (ps *PeerService) handleConnectionStateChange(conn *PeerConnection, state PeerConnectionState) {
	ps.logger.Debugf("[peerservice] peer connection state changed, id: %s, state: %s", conn.ID, state)

	// 调用状态变化回调
	if ps.onPeerConnectionStateChange != nil {
		ps.onPeerConnectionStateChange(conn.ID, state)
	}

	// 处理特殊状态
	switch state {
	case PeerConnectionStateConnected:
		ps.logger.Debugf("[peerservice] peer connection connected, id: %s", conn.ID)
	case PeerConnectionStateConnecting:
		ps.logger.Debugf("[peerservice] peer connection connecting, id: %s", conn.ID)
	case PeerConnectionStateDisconnected:
		ps.logger.Debugf("[peerservice] peer connection disconnected, id: %s", conn.ID)
	case PeerConnectionStateNew:
		ps.logger.Debugf("[peerservice] peer connection state is new, id: %s", conn.ID)
	case PeerConnectionStateFailed:
		// 连接失败，自动关闭
		ps.logger.Warnf("[peerservice] peer connection failed, closing automatically, id: %s", conn.ID)
		go ps.ClosePeerConnection(conn.ID)
		if ps.onPeerConnectionFailed != nil {
			ps.onPeerConnectionFailed(conn.ID, ErrConnectionClosed)
		}
	case PeerConnectionStateClosed:
		ps.logger.Debugf("[peerservice] peer connection closed, id: %s", conn.ID)
	}
}

// handleTrack 内部处理媒体轨道的方法
//
// @description 内部处理收到媒体轨道的方法
// @param conn *PeerConnection 对等连接实例
// @param track *webrtc.TrackRemote 媒体轨道
// @param receiver *webrtc.RTPReceiver RTP接收器
func (ps *PeerService) handleTrack(conn *PeerConnection, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	ps.logger.Debugf("[peerservice] received track, peer connection id: %s, track id: %s", conn.ID, track.ID())

	// 调用轨道回调
	if ps.onTrack != nil {
		ps.onTrack(conn.ID, track, receiver)
	}
}

// handleICECandidate 内部处理ICE候选的方法
//
// @description 内部处理收到ICE候选的方法
// @param conn *PeerConnection 对等连接实例
// @param candidate *webrtc.ICECandidate ICE候选
func (ps *PeerService) handleICECandidate(conn *PeerConnection, candidate *webrtc.ICECandidate) {
	ps.logger.Debugf("[peerservice] received ICE candidate, peer connection id: %s", conn.ID)

	// 存储ICE候选到PeerConnection中
	if candidate != nil {
		// 将ICECandidate转换为ICECandidateInit并存储
		conn.iceCandidatesMux.Lock()
		conn.iceCandidates = append(conn.iceCandidates, candidate.ToJSON())
		conn.iceCandidatesMux.Unlock()
	}

	// 调用ICE候选回调
	if ps.onICECandidate != nil && candidate != nil {
		ps.onICECandidate(conn.ID, candidate)
	}
}

// handleICEConnectionStateChange 内部处理ICE连接状态变化的方法
//
// @description 内部处理ICE连接状态变化的方法
// @param conn *PeerConnection 对等连接实例
// @param state webrtc.ICEConnectionState ICE连接状态
func (ps *PeerService) handleICEConnectionStateChange(conn *PeerConnection, state webrtc.ICEConnectionState) {
	ps.logger.Debugf("[peerservice] ICE connection state changed, peer connection id: %s, state: %s", conn.ID, state.String())

	// 调用ICE连接状态变化回调
	if ps.onICEConnectionStateChange != nil {
		ps.onICEConnectionStateChange(conn.ID, state)
	}

	// 添加特殊状态处理的日志
	switch state {
	case webrtc.ICEConnectionStateFailed:
		ps.logger.Errorf("[peerservice] ICE connection failed for peer connection: %s", conn.ID)
	case webrtc.ICEConnectionStateDisconnected:
		ps.logger.Warnf("[peerservice] ICE connection disconnected for peer connection: %s", conn.ID)
	case webrtc.ICEConnectionStateConnected:
		ps.logger.Infof("[peerservice] ICE connection established for peer connection: %s", conn.ID)
	case webrtc.ICEConnectionStateCompleted:
		ps.logger.Infof("[peerservice] ICE connection completed for peer connection: %s", conn.ID)
	default:
		ps.logger.Warnf("[peerservice] unhandled ICE connection state: %s", state.String())
	}
}

// handleDataChannel 内部处理数据通道的方法
//
// @description 内部处理数据通道创建的方法
// @param conn *PeerConnection 对等连接实例
// @param dc *webrtc.DataChannel 数据通道
func (ps *PeerService) handleDataChannel(conn *PeerConnection, dc *webrtc.DataChannel) {
	ps.logger.Debugf("[peerservice] data channel created, peer connection id: %s, channel label: %s, channel id: %d", conn.ID, dc.Label(), dc.ID())

	// 添加DataChannel的事件处理
	dc.OnOpen(func() {
		ps.logger.Infof("[peerservice] DataChannel '%s'-'%d' opened for peer connection: %s", dc.Label(), dc.ID(), conn.ID)
	})

	dc.OnClose(func() {
		ps.logger.Infof("[peerservice] DataChannel '%s'-'%d' closed for peer connection: %s", dc.Label(), dc.ID(), conn.ID)
	})

	dc.OnError(func(err error) {
		ps.logger.Errorf("[peerservice] DataChannel '%s'-'%d' error for peer connection %s: %v", dc.Label(), dc.ID(), conn.ID, err)
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		ps.logger.Infof("[peerservice] Message from DataChannel '%s'-'%d' for peer connection %s: '%s'", dc.Label(), dc.ID(), conn.ID, string(msg.Data))
	})

	// 调用数据通道回调
	if ps.onDataChannel != nil {
		ps.onDataChannel(conn.ID, dc)
	}
}
