package webrtc

import (
	"context"
	"errors"
	"sync"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

// Control 是WebRTC控制器的主要接口
//
// @description 封装了WebRTC相关的所有功能，提供给上层应用使用。
// Control 不直接管理连接，而是将所有连接管理委托给 peerServiceImpl，只对外提供调用方法，简化API接口。
// 支持创建、关闭和管理WebRTC对等连接，提供各种回调函数，用于处理连接状态变化、媒体轨道等事件。
// 并发安全边界：Control 自身的 StartUp/Shutdown 通过 sync.Once 保证只执行一次；
// 对 peerServiceImpl 的并发安全由 peerServiceImpl 内部的 concurrent.Map 与 atomic.Pointer 保证
// @struct
type Control struct {
	ctx          context.Context      // ctx 上下文，用于控制生命周期
	cancel       context.CancelFunc   // cancel 上下文取消函数，用于取消所有子goroutine
	webRTCConfig *config.WebRTCConfig // webRTCConfig WebRTC配置信息，包含ICE服务器等配置
	logger       *zap.SugaredLogger   // logger 日志记录器，用于记录WebRTC相关日志
	peerService  PeerService          // peerService 内部使用的对等连接服务
	once         sync.Once            // once 确保StartUp只执行一次
}

// NewWebRTCControl 创建一个新的WebRTC控制器
//
// @description 创建一个新的WebRTC控制器实例，初始化日志记录器与对等连接服务
// @param webrtcConfig *config.WebRTCConfig WebRTC配置信息，包含ICE服务器等配置
// @param loggerControl *logger.Control 日志控制器，用于获取日志记录器
// @return *Control 创建的WebRTC控制器实例
// @return error 如果创建过程中发生错误，则返回错误信息
func NewWebRTCControl(webrtcConfig *config.WebRTCConfig, loggerControl *logger.Control) (*Control, error) {
	webrtcConfig = check.IF[*config.WebRTCConfig](webrtcConfig == nil,
		getDefaultConfig(),
		webrtcConfig,
	)
	if !webrtcConfig.Enabled {
		return nil, errors.New("webrtc is not enabled")
	}
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	// 创建带取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 获取日志记录器（按模块归属，便于日志过滤）
	// webrtcLogger := loggerControl.GenLogger(logger.ModuleWebRTC)
	webrtcLogger := loggerControl.GenLogger("")

	// 创建对等连接服务
	peerService, err := newPeerService(ctx, webrtcConfig, loggerControl)
	if err != nil {
		webrtcLogger.Errorf("[webrtc/control] create peer service failed: %v", err)
		cancel()
		return nil, err
	}

	// 创建控制器实例
	control := &Control{
		ctx:          ctx,
		cancel:       cancel,
		webRTCConfig: webrtcConfig,
		logger:       webrtcLogger,
		peerService:  peerService,
	}

	webrtcLogger.Infof("[webrtc/control] created WebRTC control")
	return control, nil
}

// StartUp 启动WebRTC控制器
//
// @description 启动WebRTC控制器，通过 sync.Once 保证仅启动一次；
// 若 peerService 未初始化则通过 failedFunc 回调通知上层
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.webRTCConfig.Enabled {
			c.logger.Infof("[webrtc/control] starting WebRTC control...")
			// 校验 peerService 是否已初始化
			if c.peerService == nil {
				err := errors.New("webrtc: peer service not initialized")
				c.logger.Errorf("[webrtc/control] startup failed: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
			c.logger.Infof("[webrtc/control] WebRTC control started successfully")
		}
	})
}

// Shutdown 关闭WebRTC控制器
//
// @description 关闭WebRTC控制器，释放所有资源并停止所有活动
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	return c.Close()
}

// CreatePeerConnection 创建一个新的WebRTC对等连接
//
// @description 创建一个新的WebRTC对等连接，自动生成连接ID
// @param connConfig *webrtc.Configuration WebRTC连接配置，如果为nil则使用默认配置
// @return string 创建的连接ID
// @return *PeerConnection 创建的对等连接实例
// @return error 如果创建过程中发生错误，则返回错误信息
func (c *Control) CreatePeerConnection(connConfig *webrtc.Configuration) (string, *PeerConnection, error) {
	c.logger.Debugf("[webrtc/control] creating peer connection")

	// 自动生成一个连接ID
	id := uuid.GetUUID()

	// 使用对等服务创建连接
	conn, err := c.peerService.CreatePeerConnection(connConfig, id)
	if err != nil {
		c.logger.Errorf("[webrtc/control] create peer connection failed: %v", err)
		return "", nil, err
	}

	c.logger.Debugf("[webrtc/control] peer connection created, id: %s", id)
	return id, conn, nil
}

// GetPeerConnection 获取指定ID的WebRTC对等连接
//
// @description 从 peerServiceImpl 获取指定ID的WebRTC对等连接
// @param id string 连接ID
// @return *PeerConnection 如果找到对应的连接，则返回该连接实例；否则返回nil
func (c *Control) GetPeerConnection(id string) *PeerConnection {
	return c.peerService.GetPeerConnection(id)
}

// GetAllPeerConnections 获取所有的WebRTC对等连接
//
// @description 获取所有WebRTC对等连接列表
// @return []*PeerConnection 所有对等连接的列表
func (c *Control) GetAllPeerConnections() []*PeerConnection {
	return c.peerService.GetAllPeerConnections()
}

// ClosePeerConnection 关闭指定ID的WebRTC对等连接
//
// @description 关闭指定ID的WebRTC对等连接
// @param id string 连接ID
// @return error 如果关闭过程中发生错误，则返回错误信息
func (c *Control) ClosePeerConnection(id string) error {
	c.logger.Debugf("[webrtc/control] closing peer connection, id: %s", id)
	err := c.peerService.ClosePeerConnection(id)
	if err != nil {
		c.logger.Errorf("[webrtc/control] close peer connection failed: %v", err)
		return err
	}
	c.logger.Debugf("[webrtc/control] peer connection closed, id: %s", id)
	return nil
}

// CloseAllPeerConnections 关闭所有的WebRTC对等连接
//
// @description 关闭所有WebRTC对等连接，遍历连接列表逐个关闭
func (c *Control) CloseAllPeerConnections() {
	c.logger.Info("[webrtc/control] closing all peer connections")
	// 获取所有连接ID
	connections := c.peerService.GetAllPeerConnections()
	c.logger.Debugf("[webrtc/control] found %d connections to close", len(connections))

	// 关闭每个连接
	for _, conn := range connections {
		if err := c.peerService.ClosePeerConnection(conn.ID); err != nil {
			c.logger.Errorf("[webrtc/control] failed to close peer connection %s: %v", conn.ID, err)
		} else {
			c.logger.Debugf("[webrtc/control] peer connection closed successfully: %s", conn.ID)
		}
	}
}

// Close 关闭WebRTC控制器
//
// @description 关闭WebRTC控制器，关闭所有对等连接并取消上下文停止所有goroutine
// @return error 关闭过程中产生的错误
func (c *Control) Close() error {
	c.logger.Info("[webrtc/control] closing WebRTC control")

	// 关闭所有对等连接
	c.CloseAllPeerConnections()

	// 取消上下文，停止所有goroutine
	c.cancel()

	c.logger.Info("[webrtc/control] WebRTC control closed")
	return nil
}

// SendTo 向指定ID的连接发送消息
//
// @description 向指定ID的连接发送消息
// @param id string 连接ID
// @param message []byte 要发送的消息内容
// @return error 如果发送过程中发生错误，则返回错误信息
func (c *Control) SendTo(id string, message []byte) error {
	return c.peerService.SendTo(id, message)
}

// 回调函数设置方法

// OnPeerConnectionCreated 设置对等连接创建时的回调
//
// @description 设置对等连接创建时的回调
// @param callback func(id string, conn *PeerConnection) 当新的对等连接创建时被调用的回调函数
func (c *Control) OnPeerConnectionCreated(callback func(id string, conn *PeerConnection)) {
	c.peerService.OnPeerConnectionCreated(callback)
}

// OnPeerConnectionClosed 设置对等连接关闭时的回调
//
// @description 设置对等连接关闭时的回调
// @param callback func(id string) 当对等连接关闭时被调用的回调函数
func (c *Control) OnPeerConnectionClosed(callback func(id string)) {
	c.peerService.OnPeerConnectionClosed(callback)
}

// OnPeerConnectionFailed 设置对等连接失败时的回调
//
// @description 设置对等连接失败时的回调
// @param callback func(id string, err error) 当对等连接失败时被调用的回调函数
func (c *Control) OnPeerConnectionFailed(callback func(id string, err error)) {
	c.peerService.OnPeerConnectionFailed(callback)
}

// OnPeerConnectionStateChange 设置对等连接状态变化时的回调
//
// @description 设置对等连接状态变化时的回调
// @param callback func(id string, state PeerConnectionState) 当对等连接状态发生变化时被调用的回调函数
func (c *Control) OnPeerConnectionStateChange(callback func(id string, state PeerConnectionState)) {
	c.peerService.OnPeerConnectionStateChange(callback)
}

// OnTrack 设置收到媒体轨道时的回调
//
// @description 设置收到媒体轨道时的回调
// @param callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) 当收到新的媒体轨道时被调用的回调函数
func (c *Control) OnTrack(callback func(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver)) {
	c.peerService.OnTrack(callback)
}

// OnICECandidate 设置收到ICE候选时的回调
//
// @description 设置收到ICE候选时的回调
// @param callback func(id string, candidate *webrtc.ICECandidate) 当收到新的ICE候选时被调用的回调函数
func (c *Control) OnICECandidate(callback func(id string, candidate *webrtc.ICECandidate)) {
	c.peerService.OnICECandidate(callback)
}

// OnICEConnectionStateChange 设置ICE连接状态变化时的回调
//
// @description 设置ICE连接状态变化时的回调
// @param callback func(id string, state webrtc.ICEConnectionState) 当ICE连接状态发生变化时被调用的回调函数
func (c *Control) OnICEConnectionStateChange(callback func(id string, state webrtc.ICEConnectionState)) {
	c.peerService.OnICEConnectionStateChange(callback)
}

// OnDataChannel 设置数据通道创建时的回调
//
// @description 设置数据通道创建时的回调
// @param callback func(id string, dc *webrtc.DataChannel) 当新的数据通道创建时被调用的回调函数
func (c *Control) OnDataChannel(callback func(id string, dc *webrtc.DataChannel)) {
	c.peerService.OnDataChannel(callback)
}
