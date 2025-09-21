package webrtc

import (
	"context"
	"sync"

	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

// Control 是WebRTC控制器的主要接口
// 它封装了WebRTC相关的所有功能，提供给上层应用使用
// Control不直接管理连接，而是将所有连接管理委托给PeerService
// 只对外提供调用方法，简化API接口
// 支持创建、关闭和管理WebRTC对等连接
// 提供各种回调函数，用于处理连接状态变化、媒体轨道等事件
// 内部使用PeerService来实现实际的功能
// Control是线程安全的，可以在多个goroutine中并发调用
type Control struct {
	// 上下文，用于控制生命周期
	ctx    context.Context
	cancel context.CancelFunc

	// 配置信息
	webRTCConfig *config.WebRTCConfig

	// 日志记录器
	logger *zap.SugaredLogger

	// 内部使用的对等连接服务
	peerService *PeerService

	// 同步互斥锁，用于保证并发安全
	mux sync.RWMutex

	// 确保StartUp只执行一次
	once sync.Once
}

// NewWebRTCControl 创建一个新的WebRTC控制器
// 参数：
// - webRTCConfig: WebRTC配置信息，包含ICE服务器等配置
// - loggerControl: 日志控制器，用于获取日志记录器
// 返回值：
// - *Control: 创建的WebRTC控制器实例
// - error: 如果创建过程中发生错误，则返回错误信息
func NewWebRTCControl(config *config.WebRTCConfig,
	loggerControl *logger.Control,
) (*Control, error) {
	// 创建带取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 获取日志记录器
	log := loggerControl.GenLogger(logger.ModuleWebRTC)

	// 创建对等连接服务
	peerService, err := newPeerService(ctx, config, loggerControl)
	if err != nil {
		log.Errorf("[control] create peer service failed: %v", err)
		cancel()
		return nil, err
	}

	// 创建控制器实例
	control := &Control{
		ctx:          ctx,
		cancel:       cancel,
		webRTCConfig: config,
		logger:       log,
		peerService:  peerService,
	}

	log.Infof("[control] created WebRTC control")
	return control, nil
}

// StartUp 启动WebRTC控制器
// 参数：
// - failedFunc: 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Infof("[control] starting WebRTC control...")
		// WebRTC控制器主要在初始化时就完成了准备工作
		// 这里主要是记录启动日志并验证服务状态
		c.logger.Infof("[control] WebRTC control started successfully")
	})
}

// Shutdown 关闭WebRTC控制器
// 返回值：
// - error: 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	c.Close()
	return nil
}

// CreatePeerConnection 创建一个新的WebRTC对等连接
// 参数：
// - connConfig: WebRTC连接配置，如果为nil则使用默认配置
// 返回值：
// - string: 创建的连接ID
// - *WebRTCPeerConnection: 创建的对等连接实例
// - error: 如果创建过程中发生错误，则返回错误信息
func (c *Control) CreatePeerConnection(connConfig *webrtc.Configuration) (string, *PeerConnection, error) {
	c.logger.Debugf("[control] creating peer connection")

	// 自动生成一个连接ID
	id := uuid.GetUUID()

	// 使用对等服务创建连接
	conn, err := c.peerService.CreatePeerConnection(connConfig, id)
	if err != nil {
		c.logger.Errorf("[control] create peer connection failed: %v", err)
		return "", nil, err
	}

	c.logger.Debugf("[control] peer connection created, id: %s", id)
	return id, conn, nil
}

// GetPeerConnection 获取指定ID的WebRTC对等连接
// 参数：
// - id: 连接ID
// 返回值：
// - *WebRTCPeerConnection: 如果找到对应的连接，则返回该连接实例；否则返回nil
func (c *Control) GetPeerConnection(id string) *PeerConnection {
	return c.peerService.GetPeerConnection(id)
}

// GetAllPeerConnections 获取所有的WebRTC对等连接
// 返回值：
// - []*WebRTCPeerConnection: 所有对等连接的列表
func (c *Control) GetAllPeerConnections() []*PeerConnection {
	return c.peerService.GetAllPeerConnections()
}

// ClosePeerConnection 关闭指定ID的WebRTC对等连接
// 参数：
// - id: 连接ID
// 返回值：
// - error: 如果关闭过程中发生错误，则返回错误信息
func (c *Control) ClosePeerConnection(id string) error {
	c.logger.Debugf("[control] closing peer connection, id: %s", id)
	return c.peerService.ClosePeerConnection(id)
}

// CloseAllPeerConnections 关闭所有的WebRTC对等连接
func (c *Control) CloseAllPeerConnections() {
	c.logger.Info("[control] closing all peer connections")
	c.peerService.CloseAllPeerConnections()
}

// Close 关闭WebRTC控制器
// 释放所有资源并停止所有活动
func (c *Control) Close() {
	c.logger.Info("[control] closing WebRTC control")

	// 关闭所有对等连接
	c.CloseAllPeerConnections()

	// 取消上下文，停止所有goroutine
	c.cancel()

	c.logger.Info("[control] WebRTC control closed")
}

// Broadcast 向所有连接广播消息
// 参数：
// - message: 要广播的消息内容
func (c *Control) Broadcast(message []byte) {
	c.peerService.Broadcast(message)
}

// SendTo 向指定ID的连接发送消息
// 参数：
// - id: 连接ID
// - message: 要发送的消息内容
// 返回值：
// - error: 如果发送过程中发生错误，则返回错误信息
func (c *Control) SendTo(id string, message []byte) error {
	return c.peerService.SendTo(id, message)
}

// 回调函数设置方法

// OnPeerConnectionCreated 设置对等连接创建时的回调
// 参数：
// - callback: 当新的对等连接创建时被调用的回调函数
func (c *Control) OnPeerConnectionCreated(callback func(id string, conn *PeerConnection)) {
	c.peerService.OnPeerConnectionCreated(callback)
}

// OnPeerConnectionClosed 设置对等连接关闭时的回调
// 参数：
// - callback: 当对等连接关闭时被调用的回调函数
func (c *Control) OnPeerConnectionClosed(callback func(id string)) {
	c.peerService.OnPeerConnectionClosed(callback)
}

// OnPeerConnectionFailed 设置对等连接失败时的回调
// 参数：
// - callback: 当对等连接失败时被调用的回调函数
func (c *Control) OnPeerConnectionFailed(callback func(id string, err error)) {
	c.peerService.OnPeerConnectionFailed(callback)
}

// OnPeerConnectionStateChange 设置对等连接状态变化时的回调
// 参数：
// - callback: 当对等连接状态发生变化时被调用的回调函数
func (c *Control) OnPeerConnectionStateChange(callback func(id string, state PeerConnectionState)) {
	c.peerService.OnPeerConnectionStateChange(callback)
}

// OnTrack 设置收到媒体轨道时的回调
// 参数：
// - callback: 当收到新的媒体轨道时被调用的回调函数
func (c *Control) OnTrack(callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)) {
	c.peerService.OnTrack(callback)
}

// OnICECandidate 设置收到ICE候选时的回调
// 参数：
// - callback: 当收到新的ICE候选时被调用的回调函数
func (c *Control) OnICECandidate(callback func(id string, candidate *webrtc.ICECandidate)) {
	c.peerService.OnICECandidate(callback)
}

// OnICEConnectionStateChange 设置ICE连接状态变化时的回调
// 参数：
// - callback: 当ICE连接状态发生变化时被调用的回调函数
func (c *Control) OnICEConnectionStateChange(callback func(id string, state webrtc.ICEConnectionState)) {
	c.peerService.OnICEConnectionStateChange(callback)
}

// OnDataChannel 设置数据通道创建时的回调
// 参数：
// - callback: 当新的数据通道创建时被调用的回调函数
func (c *Control) OnDataChannel(callback func(id string, dc *webrtc.DataChannel)) {
	c.peerService.OnDataChannel(callback)
}
