package libp2p

import (
	"context"
	"fmt"
	"sync"

	"github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// Message 定义消息结构
type Message struct {
	Type    string
	Content []byte
	From    peer.ID
}

// MessageHandler 消息处理函数类型
type MessageHandler func(msg *Message)

// Control 控制libp2p连接的核心结构体
type Control struct {
	logger *zap.SugaredLogger
	// libp2p相关字段
	host      host.Host
	ctx       context.Context
	cancel    context.CancelFunc
	protocols map[protocol.ID]MessageHandler

	// 消息处理
	handlers      map[string]MessageHandler
	handlersMutex sync.RWMutex

	// 发现服务
	discoveryService *DiscoveryService

	// 自定义libp2p配置
	libp2pConfig *config.Libp2pConfig
}

// NewLibp2pControl 创建一个新的libp2p控制器
func NewLibp2pControl(libp2pConfig *config.Libp2pConfig, loggerControl *logger.Control) (*Control, error) {
	libp2pLogger := loggerControl.GenLogger("libp2p")
	libp2pLogger.Infof("starting new libp2p control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	// 初始化Libp2pControl
	lc := &Control{
		logger:       libp2pLogger,
		ctx:          ctx,
		cancel:       cancel,
		protocols:    make(map[protocol.ID]MessageHandler),
		handlers:     make(map[string]MessageHandler),
		libp2pConfig: libp2pConfig,
	}

	// 初始化libp2p节点
	if err := lc.initNode(); err != nil {
		lc.logger.Errorf("libp2p initialization node faild: %v", err)
		cancel()
		return nil, err
	}

	// 初始化发现服务
	discoveryService, err := newDiscoveryService(lc.ctx, lc.host,
		lc.libp2pConfig.ServiceTag, libp2pLogger, lc.libp2pConfig)
	if err != nil {
		cancel()
		lc.logger.Errorf("failt to new discovery service: %v", err)
		if err := lc.Shutdown(); err != nil {
			lc.logger.Errorf("fail to shutdown libp2p service: %v", err)
			return nil, err
		}
		return nil, err
	}
	lc.logger.Infof("settings discovery service...")
	lc.discoveryService = discoveryService

	// init dht
	if err := lc.initDHT(); err != nil {
		lc.logger.Errorf("libp2p initialization dht faild: %v", err)
		cancel()
		return nil, err
	}

	return lc, nil
}

func (lc *Control) initDHT() error {

	// 初始化discoveryService中的DHT服务
	if err := lc.discoveryService.InitDHT(); err != nil {
		lc.logger.Errorf("failed to initialize DHT service: %v", err)
		return err
	}
	return nil
}

// 初始化libp2p节点
func (lc *Control) initNode() error {
	lc.logger.Infof("initializing libp2p node...")
	// 创建libp2p节点选项
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(lc.libp2pConfig.ListenAddr),
	}

	// 创建host
	h, err := libp2p.New(opts...)
	if err != nil {
		lc.logger.Errorf("failed to create libp2p host: %v", err)
		return err
	}

	lc.host = h

	// 打印节点信息
	lc.logger.Infof("libp2p node created successfully")
	lc.logger.Infof("libp2p node ID: %s", h.ID())
	lc.logger.Infof("libp2p node addresses: %v", h.Addrs())

	protocolID := protocol.ID(lc.libp2pConfig.ProtocolID)

	// 注册默认的消息处理协议
	lc.logger.Debugf("register default message protocol...")
	lc.protocols[protocolID] = lc.defaultMessageHandler

	// 设置流处理器
	lc.logger.Debugf("register default stream handler...")
	h.SetStreamHandler(protocolID, lc.handleStream)

	return nil
}

// Start 启动libp2p服务
func (lc *Control) Start() error {
	lc.logger.Infof("starting libp2p service...")
	// 启动发现服务
	if err := lc.discoveryService.Start(); err != nil {
		lc.logger.Errorf("failed to start discovery service: %v", err)
		return err
	}

	// 启动DHT引导
	lc.logger.Infof("bootstrapping DHT...")
	if err := lc.discoveryService.BootstrapDHT(); err != nil {
		lc.logger.Errorf("failed to bootstrap DHT: %v", err)
		return err
	}

	// 连接bootstrap节点
	go lc.discoveryService.ConnectBootstrapPeers()

	// 启动健康检查
	go lc.discoveryService.StartHealthCheck()

	lc.logger.Infof("libp2p service started...")
	return nil
}

// Shutdown 关闭libp2p服务
func (lc *Control) Shutdown() error {
	lc.logger.Infof("shutting down libp2p service...")
	lc.logger.Debugf("call context cancel...")
	// 取消上下文
	lc.cancel()

	// 停止发现服务
	if lc.discoveryService != nil {
		lc.logger.Debugf("shutdown libp2p discovery service...")
		if err := lc.discoveryService.Stop(); err != nil {
			lc.logger.Errorf("faild to stop discovery service: %v", err)
			return err
		}
	}

	// 关闭host
	if lc.host != nil {
		lc.logger.Debugf("close libp2p host...")
		if err := lc.host.Close(); err != nil {
			lc.logger.Errorf("failed to close libp2p host: %v", err)
			return err
		}
	}

	lc.logger.Infof("libp2p service shutdown")
	return nil
}

// BroadcastMessage 广播消息到所有节点
func (lc *Control) BroadcastMessage(msg *Message) error {
	lc.logger.Infof("starting broadcast message...")
	lc.discoveryService.peersMutex.RLock()
	defer lc.discoveryService.peersMutex.RUnlock()

	protocolID := protocol.ID(lc.libp2pConfig.ProtocolID)

	// 遍历所有已知节点并发送消息
	for peerID := range lc.discoveryService.peers {
		lc.logger.Debugf("broadcasting message to peer %s", peerID)
		if err := lc.sendMessage(peerID, protocolID, msg); err != nil {
			lc.logger.Warnf("failed to send message to peer %s: %v", peerID, err)
		}
	}

	lc.logger.Infof("broadcast message to %d peers...", len(lc.discoveryService.peers))
	return nil
}

// SendMessageToPeer 发送消息到指定节点
func (lc *Control) SendMessageToPeer(peerID peer.ID, msg *Message) error {
	lc.logger.Infof("sending message to peer %s", peerID)
	protocolID := protocol.ID(lc.libp2pConfig.ProtocolID)
	return lc.sendMessage(peerID, protocolID, msg)
}

// 发送消息的内部方法
func (lc *Control) sendMessage(peerID peer.ID, protocolID protocol.ID, msg *Message) error {
	lc.logger.Infof("sending message to peer %s", peerID)
	// 创建到目标节点的流
	stream, err := lc.host.NewStream(lc.ctx, peerID, protocolID)
	if err != nil {
		lc.logger.Errorf("failed to create stream to peer %s: %v", peerID, err)
		return err
	}
	defer func() {
		if err := stream.Close(); err != nil {
			lc.logger.Errorf("failed to close stream to peer %s: %v", peerID, err)
		}
	}()

	// 实现消息序列化和发送逻辑
	// 注意：这里简化了消息发送逻辑，实际项目中需要实现完整的消息编码/解码
	// todo 添加编码相关代码

	lc.logger.Infof("send message to peer %s sccuess...", peerID)
	return nil
}

// 处理接收到的流
func (lc *Control) handleStream(stream network.Stream) {
	lc.logger.Infof("default handle stream...")
	defer func() {
		if err := stream.Close(); err != nil {
			lc.logger.Errorf("failed to close stream: %v", err)
		}
	}()

	peerID := stream.Conn().RemotePeer()
	protocolID := stream.Protocol()

	lc.logger.Infof("received stream from peer %s using protocol %s", peerID, protocolID)

	// 注册peer
	// lc.peersMutex.Lock()
	// lc.peers[peerID] = struct{}{}
	// lc.peersMutex.Unlock()

	// 获取协议对应的处理器
	handler, ok := lc.protocols[protocolID]
	if !ok {
		lc.logger.Errorf("no handler found for protocol %s", protocolID)
		return
	}

	// 实现消息接收和处理逻辑
	// 注意：这里简化了消息接收逻辑，实际项目中需要实现完整的消息解码
	// 示例中仅创建一个简单的消息对象
	msg := &Message{

		From: peerID,
	}

	// 调用处理器处理消息
	handler(msg)
}

// 默认的消息处理器
func (lc *Control) defaultMessageHandler(msg *Message) {
	lc.logger.Infof("default message handler...")
	lc.logger.Debugf("received message from %s, type: %s", msg.From, msg.Type)

	lc.logger.Infof("msg %v", msg)

	// 根据消息类型调用对应的处理器
	handler, ok := lc.handlers[msg.Type]
	if ok {
		handler(msg)
	} else {
		lc.logger.Warnf("no handler registered for message type: %s", msg.Type)
	}
}

// RegisterMessageHandler 注册消息处理器
func (lc *Control) RegisterMessageHandler(messageType string, handler MessageHandler) {
	lc.logger.Debugf("registering handler for message type: %s", messageType)
	lc.handlersMutex.Lock()
	defer lc.handlersMutex.Unlock()

	lc.handlers[messageType] = handler
	lc.logger.Infof("registered handler for message type: %s", messageType)
}

// GetPeers 获取所有连接的节点
func (lc *Control) GetPeers() []peer.ID {
	lc.logger.Infof("starting getting peers...")
	lc.discoveryService.peersMutex.RLock()
	defer lc.discoveryService.peersMutex.RUnlock()

	peers := make([]peer.ID, 0, len(lc.discoveryService.peers))
	for peerID := range lc.discoveryService.peers {
		peers = append(peers, peerID)
	}

	return peers
}

// ConnectToPeer 手动连接到指定节点
func (lc *Control) ConnectToPeer(addr string) error {
	lc.logger.Infof("connecting to peer %s", addr)
	// 解析multiaddr
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		lc.logger.Errorf("failed to parse multiaddr: %v", err)
		return err
	}

	// 解析peer信息
	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		lc.logger.Errorf("failed to parse peer info: %v", err)
		return err
	}

	// 连接到节点
	if err := lc.host.Connect(lc.ctx, *peerInfo); err != nil {
		lc.logger.Errorf("failed to connect to peer: %v", err)
		return err
	}

	// 注册peer
	lc.discoveryService.peersMutex.Lock()
	lc.discoveryService.peers[peerInfo.ID] = struct{}{}
	lc.discoveryService.peersMutex.Unlock()

	lc.logger.Infof("successfully connected to peer %s", peerInfo.ID)
	return nil
}

// GetBootstrapPeers 获取所有bootstrap节点
func (lc *Control) GetBootstrapPeers() []peer.ID {
	if lc.discoveryService == nil {
		return []peer.ID{}
	}
	return lc.discoveryService.GetBootstrapPeers()
}

// FindPeer 使用DHT查找指定ID的节点
func (lc *Control) FindPeer(peerID peer.ID) (*peer.AddrInfo, error) {
	if lc.discoveryService == nil {
		return nil, fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.FindPeer(peerID)
}

// GetDHTRoutingTableInfo 获取DHT路由表信息
func (lc *Control) GetDHTRoutingTableInfo() (int, error) {
	if lc.discoveryService == nil {
		return 0, fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.GetDHTRoutingTableInfo()
}

// Provide 使用DHT提供数据索引
func (lc *Control) Provide(key string) error {
	if lc.discoveryService == nil {
		return fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.Provide(key)
}

// FindProviders 使用DHT查找提供指定数据的节点
func (lc *Control) FindProviders(key string, count int) ([]peer.AddrInfo, error) {
	if lc.discoveryService == nil {
		return nil, fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.FindProviders(key, count)
}
