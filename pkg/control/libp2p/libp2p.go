package libp2p

import (
	"context"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	securitynoise "github.com/libp2p/go-libp2p/p2p/security/noise"
	securitytls "github.com/libp2p/go-libp2p/p2p/security/tls"
	transportquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	transporttcp "github.com/libp2p/go-libp2p/p2p/transport/tcp"
	transportwebrtc "github.com/libp2p/go-libp2p/p2p/transport/webrtc"
	transportwebsocket "github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	webtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"go.uber.org/zap"
)

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
	libp2pLogger.Infof("[control] starting new libp2p control...")

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
		lc.logger.Errorf("[control] libp2p initialization node faild: %v", err)
		cancel()
		return nil, err
	}

	// 初始化发现服务
	discoveryService, err := newDiscoveryService(lc.ctx, lc.host,
		lc.libp2pConfig.ServiceTag, libp2pLogger, lc.libp2pConfig)
	if err != nil {
		cancel()
		lc.logger.Errorf("[control] failt to new discovery service: %v", err)
		if err := lc.Shutdown(); err != nil {
			lc.logger.Errorf("[control] fail to shutdown libp2p service: %v", err)
			return nil, err
		}
		return nil, err
	}
	lc.logger.Debugf("[control] settings discovery service...")
	lc.discoveryService = discoveryService

	// init dht
	if err := lc.initDHT(); err != nil {
		lc.logger.Errorf("[control] libp2p initialization dht faild: %v", err)
		cancel()
		return nil, err
	}

	return lc, nil
}

func (lc *Control) GetLocalID() peer.ID {
	lc.logger.Debugf("[control] get local id...")
	return lc.host.ID()
}

func (lc *Control) initDHT() error {
	lc.logger.Debugf("[control] control init dht...")
	// 初始化discoveryService中的DHT服务
	if err := lc.discoveryService.InitDHT(); err != nil {
		lc.logger.Errorf("[control] failed to initialize DHT service: %v", err)
		return err
	}
	return nil
}

// 初始化libp2p节点
func (lc *Control) initNode() error {
	lc.logger.Debugf("[control] initializing libp2p node...")
	// 创建libp2p节点选项
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(lc.libp2pConfig.ListenAddr...),

		// libp2p.DefaultTransports,
		libp2p.Transport(transporttcp.NewTCPTransport),
		libp2p.Transport(transportquic.NewTransport),
		libp2p.Transport(transportwebsocket.New),
		libp2p.Transport(webtransport.New),
		libp2p.Transport(transportwebrtc.New),

		libp2p.Ping(false),

		libp2p.DefaultConnectionManager,

		libp2p.EnableRelay(),

		libp2p.EnableNATService(), // NAT
		libp2p.EnableAutoNATv2(),

		libp2p.Security(securitytls.ID, securitytls.New),
		libp2p.Security(securitynoise.ID, securitynoise.New),
	}

	sk, err := lc.libp2pConfig.Identity.DecodePrivateKey("")
	if err != nil {
		lc.logger.Errorf("[control] failed to decode private key: %v", err)
		return err
	}
	opts = append(opts, libp2p.Identity(sk))

	// 创建host
	h, err := libp2p.New(opts...)
	if err != nil {
		lc.logger.Errorf("[control] failed to create libp2p host: %v", err)
		return err
	}

	lc.host = h

	// 打印节点信息
	lc.logger.Infof("[control] libp2p node created successfully")
	lc.logger.Infof("[control] libp2p node ID: %s", h.ID())

	addrs := func(addrs []multiaddr.Multiaddr) string {
		if len(addrs) == 0 {
			return "[]"
		}
		str := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			str = append(str, addr.String())
		}
		return "[\n" + strings.Join(str, "\n") + "\n]"
	}(h.Addrs())

	lc.logger.Infof("[control] libp2p node addresses: \n%v", addrs)

	// 注册配置文件中的 protocol
	protocolID := protocol.ID(lc.libp2pConfig.ProtocolID)
	lc.RegisterProtocolHandler(protocolID, lc.defaultMessageHandler, lc.defaultStreamHandler)

	// 注册一些默认消息
	lc.defaultMessageRegister()

	return nil
}

// StartUp 启动libp2p服务
func (lc *Control) StartUp(failedFunc func(err error)) {
	lc.logger.Infof("[control] starting libp2p service...")
	// 启动发现服务
	if err := lc.discoveryService.Start(); err != nil {
		lc.logger.Errorf("[control] failed to start discovery service: %v", err)
		failedFunc(err)
	}

	// 启动DHT引导
	lc.logger.Infof("[control] bootstrapping DHT...")
	if err := lc.discoveryService.BootstrapDHT(); err != nil {
		lc.logger.Errorf("[control] failed to bootstrap DHT: %v", err)
		failedFunc(err)
	}

	// 连接bootstrap节点
	go lc.discoveryService.ConnectBootstrapPeers()

	// 启动健康检查
	go lc.discoveryService.StartHealthCheck()

	lc.logger.Infof("[control] libp2p service started...")
}

// Shutdown 关闭libp2p服务
func (lc *Control) Shutdown() error {
	lc.logger.Debugf("[control] shutting down libp2p service...")

	lc.logger.Debugf("[control] broadcast my shutdown message...")
	_ = lc.BroadcastMessage(&Message{
		Type:    "base/shutdown",
		From:    lc.GetLocalID(),
		Content: []byte("bye"),
	})

	lc.logger.Debugf("[control] call context cancel...")
	// 取消上下文
	lc.cancel()

	// 停止发现服务
	if lc.discoveryService != nil {
		lc.logger.Debugf("[control] shutdown libp2p discovery service...")
		if err := lc.discoveryService.Stop(); err != nil {
			lc.logger.Errorf("[control] faild to stop discovery service: %v", err)
			return err
		}
	}

	// 关闭host
	if lc.host != nil {
		lc.logger.Debugf("[control] close libp2p host...")
		if err := lc.host.Close(); err != nil {
			lc.logger.Errorf("[control] failed to close libp2p host: %v", err)
			return err
		}
	}

	lc.logger.Infof("[control] libp2p service shutdown")
	return nil
}

// BroadcastMessage 广播消息到所有节点
func (lc *Control) BroadcastMessage(msg *Message) error {
	lc.logger.Debugf("[control] starting broadcast message...")
	lc.discoveryService.peersMutex.RLock()
	defer lc.discoveryService.peersMutex.RUnlock()

	protocolID := protocol.ID(lc.libp2pConfig.ProtocolID)

	// 遍历所有已知节点并发送消息
	for peerID := range lc.discoveryService.peers {
		msg.To = peerID
		lc.logger.Debugf("[control] broadcasting message to peer %s", peerID)
		if err := lc.sendMessage(peerID, protocolID, msg); err != nil {
			lc.logger.Warnf("[control] failed to send message to peer %s: %v", peerID, err)
		}
		time.Sleep(time.Duration(rand.IntN(500)) * time.Millisecond)
	}

	lc.logger.Debugf("[control] broadcast message to %d peers...", len(lc.discoveryService.peers))
	return nil
}

// SendMessageToPeer 发送消息到指定节点
func (lc *Control) SendMessageToPeer(peerID peer.ID, msg *Message) error {
	lc.logger.Debugf("[control] sending message to peer %s", peerID)
	protocolID := protocol.ID(lc.libp2pConfig.ProtocolID)
	return lc.sendMessage(peerID, protocolID, msg)
}

// 发送消息的内部方法
func (lc *Control) sendMessage(peerID peer.ID, protocolID protocol.ID, msg *Message) error {
	lc.logger.Debugf("[control] sending message to peer %s", peerID)
	// 创建到目标节点的流
	stream, err := lc.host.NewStream(lc.ctx, peerID, protocolID)
	if err != nil {
		lc.logger.Errorf("[control] failed to create stream to peer %s: %v", peerID, err)
		return err
	}
	defer func() {
		if err := stream.CloseWrite(); err != nil {
			lc.logger.Errorf("[control] failed to close stream writer to peer %s: %v", peerID, err)
		}
	}()

	if err = json.NewEncoder(stream).Encode(msg); err != nil {
		lc.logger.Errorf("[control] failed to encode message to peer %s: %v", peerID, err)
		return err
	}

	lc.logger.Infof("[control] send message messageType %v to peer %s sccuess...", msg.Type, peerID)
	return nil
}

// 处理接收到的流
func (lc *Control) defaultStreamHandler(stream network.Stream) {
	lc.logger.Debugf("[control] default stream handler...")
	defer func() {
		if err := stream.CloseRead(); err != nil {
			lc.logger.Errorf("[control] failed to close stream read: %v", err)
		}
	}()

	peerID := stream.Conn().RemotePeer()
	protocolID := stream.Protocol()

	lc.logger.Debugf("[control] received stream from peer %s using protocol %s", peerID, protocolID)

	// 获取协议对应的处理器
	handler, ok := lc.protocols[protocolID]
	if !ok {
		lc.logger.Errorf("[control] no handler found for protocol %s", protocolID)
		return
	}

	var msg Message
	if err := json.NewDecoder(stream).Decode(&msg); err != nil {
		lc.logger.Errorf("[control] failed to decode message from peer %s: %v", peerID, err)
		return
	}

	// 调用处理器处理消息
	handler(protocolID, &msg)
}

// 默认的消息处理器
func (lc *Control) defaultMessageHandler(protocolId protocol.ID, msg *Message) {
	lc.logger.Debugf("[control] default message handler...")
	lc.logger.Debugf("[control] received message from %s, type: %s", msg.From, msg.Type)

	// 根据消息类型调用对应的处理器
	handler, ok := lc.handlers[msg.Type]
	if ok {
		handler(protocolId, msg)
	} else {
		lc.logger.Warnf("[control] no message handler registered for message type: %s", msg.Type)
	}
}

// RegisterMessageHandler 注册消息处理器
func (lc *Control) RegisterMessageHandler(messageType string, handler MessageHandler) {
	lc.logger.Debugf("[control] registering handler for message type: %s", messageType)
	lc.handlersMutex.Lock()
	defer lc.handlersMutex.Unlock()

	lc.handlers[messageType] = handler
	lc.logger.Infof("[control] registered handler for message type: %s", messageType)
}

func (lc *Control) RegisterNotifyPeerFound(call func(id peer.ID, info peer.AddrInfo)) {
	lc.logger.Debugf("[control] register notify peer found handler...")
	lc.discoveryService.NotifyPeerFound(call)
}

func (lc *Control) RegisterProtocolHandler(protocolID protocol.ID, protocolHandler MessageHandler, streamHandler network.StreamHandler) {
	lc.logger.Debugf("[control] register %v protocol handler...", protocolID)
	// 注册默认的消息处理协议
	lc.logger.Debugf("[control] register message protocol handler...")
	lc.protocols[protocolID] = protocolHandler

	// 设置流处理器
	lc.logger.Debugf("[control] register %v protocol stream handler...", protocolID)
	lc.host.SetStreamHandler(protocolID, streamHandler)
	lc.logger.Infof("[control] success register %v protocol handler...", protocolID)
}

func (lc *Control) defaultMessageRegister() {
	lc.logger.Debugf("register some default message...")
	lc.RegisterMessageHandler("base/ping", func(protocolId protocol.ID, msg *Message) {
		lc.logger.Debugf("[control] received %v protocol ping message from %v", protocolId, msg.From)
		pongMsg := &Message{
			From:    msg.To,
			To:      msg.From,
			Content: []byte("pong"),
			Type:    "base/pong",
		}
		_ = lc.SendMessageToPeer(pongMsg.To, pongMsg)
	})
	lc.RegisterMessageHandler("base/pong", func(protocolID protocol.ID, msg *Message) {
		lc.logger.Debugf("received %v protocol pong message from %v content %v", protocolID, msg.From, string(msg.Content))
	})
	lc.RegisterMessageHandler("base/shutdown", func(protocolId protocol.ID, msg *Message) {
		lc.logger.Debugf("[control] received %v protocol shutdown message from %v", protocolId, msg.From)
		lc.DisconnectFromPeer(msg.From)
		lc.logger.Infof("from connect peer list remove peer %v", msg.From)
	})
	lc.logger.Debugf("registed some default message...")
}

// ///////////////////////////////////////////////////////////////////////////////////////////

// GetPeers 获取所有连接的节点
func (lc *Control) GetPeers() []peer.ID {
	lc.logger.Debugf("[control] starting getting peers...")
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
	lc.logger.Debugf("[control] connecting to peer %s", addr)
	// 解析multiaddr
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		lc.logger.Errorf("[control] failed to parse multiaddr: %v", err)
		return err
	}

	// 解析peer信息
	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		lc.logger.Errorf("[control] failed to parse peer info: %v", err)
		return err
	}

	// 连接到节点
	if err := lc.host.Connect(lc.ctx, *peerInfo); err != nil {
		lc.logger.Errorf("[control] failed to connect to peer: %v", err)
		return err
	}

	// 注册peer
	lc.discoveryService.peersMutex.Lock()
	lc.discoveryService.peers[peerInfo.ID] = struct{}{}
	lc.discoveryService.peersMutex.Unlock()

	lc.logger.Infof("[control] successfully connected to peer %s", peerInfo.ID)
	return nil
}

// DisconnectFromPeer 断开与指定节点的连接
// @param peer peer.ID 节点ID
func (lc *Control) DisconnectFromPeer(peer peer.ID) {
	lc.logger.Debugf("[control] disconnecting from peer %v", peer)
	lc.discoveryService.peersMutex.Lock()
	delete(lc.discoveryService.peers, peer)
	lc.discoveryService.peersMutex.Unlock()
	lc.logger.Debugf("[control] successfully disconnected from peer %v", peer)
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
