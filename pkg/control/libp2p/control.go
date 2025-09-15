package libp2p

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

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

	// 消息队列，用于异步发送消息
	messageQueue chan MessageWithPeer

	// 工作协程组
	wg sync.WaitGroup

	once sync.Once
}

// NewLibp2pControl 创建一个新的libp2p控制器
//
// @param libp2pConfig *config.Libp2pConfig libp2p配置
// @param loggerControl *logger.Control 日志控制器
//
// @return *Control libp2p控制器
// @return error 错误
func NewLibp2pControl(libp2pConfig *config.Libp2pConfig, loggerControl *logger.Control) (*Control, error) {
	libp2pLogger := loggerControl.GenLogger(logger.ModuleLibp2p)
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
		// 创建带缓冲的消息队列，大小可以根据需要调整
		messageQueue: make(chan MessageWithPeer, 1024),
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

// GetLocalhostPeerID 获取本地节点ID
//
// @return peer.ID 本地节点的Peer ID
func (lc *Control) GetLocalhostPeerID() peer.ID {
	lc.logger.Debugf("[control] get local id...")
	return lc.host.ID()
}

// initDHT 初始化DHT服务
//
// @return error 错误信息，如果初始化成功则返回nil
func (lc *Control) initDHT() error {
	lc.logger.Debugf("[control] control init dht...")
	// 初始化discoveryService中的DHT服务
	if err := lc.discoveryService.InitDHT(); err != nil {
		lc.logger.Errorf("[control] failed to initialize DHT service: %v", err)
		return err
	}
	return nil
}

// initNode 初始化libp2p节点
//
// @return error 错误信息，如果初始化成功则返回nil
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
//
// @param failedFunc func(err error) 启动失败时的回调函数
func (lc *Control) StartUp(failedFunc func(err error)) {
	lc.once.Do(func() {
		lc.logger.Infof("[control] starting libp2p service...")
		// 启动发现服务
		if err := lc.discoveryService.Start(); err != nil {
			lc.logger.Errorf("[control] failed to start discovery service: %v", err)
			if failedFunc != nil {
				failedFunc(err)
			}
		}

		// 启动DHT引导
		lc.logger.Infof("[control] bootstrapping DHT...")
		if err := lc.discoveryService.BootstrapDHT(); err != nil {
			lc.logger.Errorf("[control] failed to bootstrap DHT: %v", err)
			if failedFunc != nil {
				failedFunc(err)
			}
		}

		// 连接bootstrap节点
		go lc.discoveryService.ConnectBootstrapPeers()

		// 启动健康检查
		go lc.discoveryService.StartHealthCheck()

		// 启动消息处理工作协程
		lc.wg.Add(1)
		go lc.messageProcessor()

		lc.logger.Infof("[control] libp2p service started...")
	})
}

// Shutdown 关闭libp2p服务
//
// @return error 错误信息，如果关闭成功则返回nil
func (lc *Control) Shutdown() error {
	lc.logger.Debugf("[control] shutting down libp2p service...")

	lc.logger.Debugf("[control] broadcast my shutdown message...")
	_ = lc.BroadcastMessage(&Message{
		Type:    MsgBaseShutdown,
		From:    lc.GetLocalhostPeerID(),
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

	// 关闭消息队列并等待工作协程结束
	lc.logger.Debugf("[control] closing message queue and waiting for workers...")
	close(lc.messageQueue) // 关闭通道，通知工作协程结束
	lc.wg.Wait()           // 等待所有工作协程完成

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

// BroadcastMessage 广播消息到所有已知节点
//
// @param msg *Message 要广播的消息
// @return error 错误信息，如果广播过程中出现严重错误则返回错误
func (lc *Control) BroadcastMessage(msg *Message) error {
	lc.logger.Debugf("[control] starting broadcast message...")
	lc.discoveryService.peersMutex.RLock()
	defer lc.discoveryService.peersMutex.RUnlock()

	// 遍历所有已知节点并发送消息
	for peerID := range lc.discoveryService.peers {
		msg.To = peerID
		msg.From = lc.GetLocalhostPeerID()
		lc.logger.Debugf("[control] broadcasting message to peer %s", peerID)
		if err := lc.SendMessageToPeer(peerID, msg); err != nil {
			lc.logger.Warnf("[control] failed to send message to peer %s: %v", peerID, err)
		}
		time.Sleep(time.Duration(rand.IntN(500)) * time.Millisecond)
	}

	lc.logger.Debugf("[control] broadcast message to %d peers...", len(lc.discoveryService.peers))
	return nil
}

// messageProcessor 消息处理工作协程，从消息队列读取消息并分发到工作池
//
// 该方法创建固定数量的工作协程，然后从messageQueue接收消息并发送到工作池进行处理
// 当messageQueue关闭时，该方法也会关闭工作通道并退出
func (lc *Control) messageProcessor() {
	defer lc.wg.Done()
	lc.logger.Debugf("[control] message processor started")

	// 创建工作池来限制并发处理的消息数量
	// 根据系统资源和需求调整工作协程数量
	const workerCount = 5
	jobs := make(chan MessageWithPeer)

	// 启动工作协程池
	for i := 0; i < workerCount; i++ {
		lc.wg.Add(1)
		go lc.worker(jobs)
	}

	// 从队列中接收消息并发送到工作池
	for msgWithPeer := range lc.messageQueue {
		select {
		case jobs <- msgWithPeer:
			// 消息已发送到工作池
		case <-lc.ctx.Done():
			// 上下文已取消，退出循环
			break
		}
	}

	// 关闭工作通道，通知所有工作协程退出
	close(jobs)

	lc.logger.Debugf("[control] message processor shutdown")
}

// worker 工作协程，从jobs通道中获取消息并处理
//
// @param jobs <-chan MessageWithPeer 消息通道，包含待处理的消息及其目标节点信息
func (lc *Control) worker(jobs <-chan MessageWithPeer) {
	defer lc.wg.Done()
	lc.logger.Debugf("[control] message worker started")

	for msgWithPeer := range jobs {
		select {
		case <-lc.ctx.Done():
			return
		default:
			// 处理消息
			lc.processMessage(msgWithPeer)
		}
	}

	lc.logger.Debugf("[control] message worker shutdown")
}

// processMessage 处理单个消息的实际发送逻辑
//
// @param msgWithPeer MessageWithPeer 包含待发送消息及其目标节点信息的结构体
func (lc *Control) processMessage(msgWithPeer MessageWithPeer) {
	lc.discoveryService.peersMutex.RLock()
	_, exist := lc.discoveryService.peers[msgWithPeer.PeerID]
	lc.discoveryService.peersMutex.RUnlock()
	if !exist {
		// 如果连接不存在，则跳过消息
		lc.logger.Debugf("[control] peer %s not found, skipping message", msgWithPeer.PeerID)
		return
	}

	peerID := msgWithPeer.PeerID
	protocolID := msgWithPeer.ProtocolID
	msg := msgWithPeer.Msg

	// 检查上下文是否已取消
	if lc.ctx.Err() != nil {
		lc.logger.Debugf("[control] context canceled, skipping message to peer %s", peerID)
		return
	}

	// 检查连接是否存在
	if lc.host.Network().Connectedness(peerID) == network.NotConnected {
		lc.logger.Warnf("[control] not connected to peer %s, skipping message", peerID)
		return
	}

	// 创建到目标节点的流
	stream, err := lc.host.NewStream(lc.ctx, peerID, protocolID)
	if err != nil {
		lc.logger.Errorf("[control] failed to create stream to peer %s: %v", peerID, err)
		return
	}

	// 使用defer确保流正确关闭
	defer func() {
		// 先关闭写入端
		if err := stream.CloseWrite(); err != nil {
			// 只记录关闭流写入端的错误，不影响消息处理结果
			// 许多关闭错误是正常的，例如对端已关闭连接
			lc.logger.Debugf("[control] stream write closed to peer %s: %v (non-critical)", peerID, err)
		}
		// 然后关闭整个流
		if err := stream.Close(); err != nil {
			// 只记录关闭流的错误，不影响消息处理结果
			lc.logger.Debugf("[control] stream fully closed to peer %s: %v (non-critical)", peerID, err)
		}
	}()

	// 尝试发送消息
	err = json.NewEncoder(stream).Encode(msg)
	if err != nil {
		lc.logger.Errorf("[control] failed to encode message to peer %s: %v", peerID, err)
	} else {
		lc.logger.Infof("[control] send message messageType %v to peer %s success...", msg.Type, peerID)
	}
}

// SendMessageToPeer 发送消息到指定节点
//
// @param peerID peer.ID 目标节点的Peer ID
// @param msg *Message 要发送的消息对象
// @return error 错误信息，如果消息发送成功则返回nil
func (lc *Control) SendMessageToPeer(peerID peer.ID, msg *Message) error {
	lc.logger.Debugf("[control] sending message to peer %s", peerID)
	protocolID := protocol.ID(lc.libp2pConfig.ProtocolID)
	return lc.sendMessage(peerID, protocolID, msg)
}

// sendMessage 发送消息的内部方法 - 现在改为异步方式，将消息发送到队列中
//
// @param peerID peer.ID 目标节点的Peer ID
// @param protocolID protocol.ID 协议ID
// @param msg *Message 要发送的消息对象
// @return error 错误信息，如果消息成功加入队列则返回nil
func (lc *Control) sendMessage(peerID peer.ID, protocolID protocol.ID, msg *Message) error {
	lc.logger.Debugf("[control] queueing message to peer %s", peerID)

	// 创建消息副本以避免并发问题
	msgCopy := &Message{
		Type:    msg.Type,
		Content: make([]byte, len(msg.Content)),
		From:    msg.From,
		To:      msg.To,
	}
	copy(msgCopy.Content, msg.Content)

	// 检查上下文是否已取消
	if lc.ctx.Err() != nil {
		return fmt.Errorf("context canceled")
	}

	// 将消息发送到队列
	select {
	case lc.messageQueue <- MessageWithPeer{
		PeerID:     peerID,
		ProtocolID: protocolID,
		Msg:        msgCopy,
	}:
		// 消息成功加入队列
		lc.logger.Debugf("[control] message queued successfully to peer %s", peerID)
		return nil
	case <-lc.ctx.Done():
		// 上下文已取消
		return fmt.Errorf("context canceled while queueing message")
	default:
		// 通道已满，记录警告但不阻塞主线程
		lc.logger.Warnf("[control] message queue full, dropping message to peer %s", peerID)
		return fmt.Errorf("message queue is full")
	}
}

// defaultStreamHandler 默认的流处理器，处理接收到的流
//
// @param stream network.Stream 接收到的网络流，包含来自远程节点的消息
func (lc *Control) defaultStreamHandler(stream network.Stream) {
	lc.logger.Debugf("[control] default stream handler...")
	peerID := stream.Conn().RemotePeer()
	protocolID := stream.Protocol()

	// 使用defer确保流正确关闭
	defer func() {
		// 先关闭读取端
		if err := stream.CloseRead(); err != nil {
			// 只记录关闭流读取端的错误，不影响消息处理结果
			lc.logger.Debugf("[control] stream read closed from peer %s: %v (non-critical)", peerID, err)
		}
		// 然后关闭整个流
		if err := stream.Close(); err != nil {
			// 只记录关闭流的错误，不影响消息处理结果
			lc.logger.Debugf("[control] stream fully closed from peer %s: %v (non-critical)", peerID, err)
		}
	}()

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

// defaultMessageHandler 默认的消息处理器，处理未注册特定处理器的消息
//
// @param protocolId protocol.ID 消息使用的协议ID
// @param msg *Message 接收到的消息对象
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

// RegisterMessageHandler 注册特定消息类型的处理器
//
// @param messageType string 消息类型
// @param handler MessageHandler 消息处理函数，用于处理指定类型的消息
func (lc *Control) RegisterMessageHandler(messageType string, handler MessageHandler) {
	lc.logger.Debugf("[control] registering handler for message type: %s", messageType)
	lc.handlersMutex.Lock()
	defer lc.handlersMutex.Unlock()

	lc.handlers[messageType] = handler
	lc.logger.Infof("[control] registered handler for message type: %s", messageType)
}

// RegisterNotifyPeerFound 注册节点发现回调函数
//
// @param call func(id peer.ID, info peer.AddrInfo) 节点发现时的回调函数，接收节点ID和地址信息
func (lc *Control) RegisterNotifyPeerFound(call func(id peer.ID, info peer.AddrInfo)) {
	lc.logger.Debugf("[control] register notify peer found handler...")
	lc.discoveryService.NotifyPeerFound(call)
}

// RegisterProtocolHandler 注册协议处理器和流处理器
//
// @param protocolID protocol.ID 协议ID
// @param protocolHandler MessageHandler 协议消息处理函数
// @param streamHandler network.StreamHandler 流处理函数
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

// defaultMessageRegister 注册默认的消息处理器
//
// 该方法注册基础的ping/pong和shutdown消息处理逻辑
func (lc *Control) defaultMessageRegister() {
	lc.logger.Debugf("[control] register some default message...")
	lc.RegisterMessageHandler(MsgBasePing, func(protocolId protocol.ID, msg *Message) {
		lc.logger.Debugf("[control] received %v protocol ping message from %v", protocolId, msg.From)
		pongMsg := &Message{
			From:    msg.To,
			To:      msg.From,
			Content: []byte("pong"),
			Type:    MsgBasePong,
		}
		_ = lc.SendMessageToPeer(pongMsg.To, pongMsg)
	})
	lc.RegisterMessageHandler(MsgBasePong, func(protocolID protocol.ID, msg *Message) {
		lc.logger.Debugf("[control] received %v protocol pong message from %v content %v", protocolID, msg.From, string(msg.Content))
	})
	lc.RegisterMessageHandler(MsgBaseShutdown, func(protocolId protocol.ID, msg *Message) {
		lc.logger.Debugf("[control] received %v protocol shutdown message from %v", protocolId, msg.From)
		lc.DisconnectFromPeer(msg.From)
		lc.logger.Infof("[control] from connect peer list remove peer %v", msg.From)
	})
	lc.logger.Debugf("[control] registered some default message...")
}

// GetPeers 获取所有连接的节点
// @return []peer.ID 连接的节点ID列表
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
// @param addr string 节点的multiaddr地址
// @return error 连接过程中可能产生的错误
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
// @param peer peer.ID 要断开连接的节点ID
func (lc *Control) DisconnectFromPeer(peer peer.ID) {
	lc.logger.Debugf("[control] disconnecting from peer %v", peer)
	lc.discoveryService.peersMutex.Lock()
	delete(lc.discoveryService.peers, peer)
	lc.discoveryService.peersMutex.Unlock()
	lc.logger.Debugf("[control] successfully disconnected from peer %v", peer)
}

// GetBootstrapPeers 获取所有bootstrap节点
// @return []peer.ID bootstrap节点ID列表
func (lc *Control) GetBootstrapPeers() []peer.ID {
	if lc.discoveryService == nil {
		return []peer.ID{}
	}
	return lc.discoveryService.GetBootstrapPeers()
}

// FindPeer 使用DHT查找指定ID的节点
// @param peerID peer.ID 要查找的节点ID
// @return *peer.AddrInfo 节点的地址信息
// @return error 查找过程中可能产生的错误
func (lc *Control) FindPeer(peerID peer.ID) (*peer.AddrInfo, error) {
	if lc.discoveryService == nil {
		return nil, fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.FindPeer(peerID)
}

// GetDHTRoutingTableInfo 获取DHT路由表信息
// @return int 路由表中的节点数量
// @return error 获取过程中可能产生的错误
func (lc *Control) GetDHTRoutingTableInfo() (int, error) {
	if lc.discoveryService == nil {
		return 0, fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.GetDHTRoutingTableInfo()
}

// Provide 使用DHT提供数据索引
// @param key string 要提供的数据索引键
// @return error 提供过程中可能产生的错误
func (lc *Control) Provide(key string) error {
	if lc.discoveryService == nil {
		return fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.Provide(key)
}

// FindProviders 使用DHT查找提供指定数据的节点
// @param key string 要查找的数据索引键
// @param count int 要查找的节点数量
// @return []peer.AddrInfo 提供指定数据的节点地址信息列表
// @return error 查找过程中可能产生的错误
func (lc *Control) FindProviders(key string, count int) ([]peer.AddrInfo, error) {
	if lc.discoveryService == nil {
		return nil, fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.FindProviders(key, count)
}

// GetPeerInfo 获取指定节点的完整信息
//
// @param peerID peer.ID 目标节点的ID
// @return *peer.AddrInfo 节点的地址信息，如果节点不存在则返回nil
func (lc *Control) GetPeerInfo(peerID peer.ID) *peer.AddrInfo {
	if lc.discoveryService == nil {
		lc.logger.Errorf("[control] discovery service is not initialized")
		return nil
	}
	return lc.discoveryService.GetPeerInfo(peerID)
}

// AddPeerAddresses 向PeerStore中添加节点地址
//
// @param peerID peer.ID 节点ID
// @param addrs []multiaddr.Multiaddr 节点地址列表
// @param ttl time.Duration 地址的生存时间
func (lc *Control) AddPeerAddresses(peerID peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	if lc.discoveryService == nil {
		lc.logger.Errorf("[control] discovery service is not initialized")
		return
	}
	lc.discoveryService.AddPeerAddresses(peerID, addrs, ttl)
}

// SetPeerMetadata 设置节点的元数据
//
// @param peerID peer.ID 节点ID
// @param key string 元数据键
// @param value interface{} 元数据值
func (lc *Control) SetPeerMetadata(peerID peer.ID, key string, value interface{}) error {
	if lc.discoveryService == nil {
		return fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.SetPeerMetadata(peerID, key, value)
}

// GetPeerMetadata 获取节点的元数据
//
// @param peerID peer.ID 节点ID
// @param key string 元数据键
// @return interface{} 元数据值
// @return error 错误信息
func (lc *Control) GetPeerMetadata(peerID peer.ID, key string) (interface{}, error) {
	if lc.discoveryService == nil {
		return nil, fmt.Errorf("discovery service is not initialized")
	}
	return lc.discoveryService.GetPeerMetadata(peerID, key)
}

// RemovePeerMetadata 移除节点的元数据
//
// @param peerID peer.ID 节点ID
// @param key string 元数据键
func (lc *Control) RemovePeerMetadata(peerID peer.ID, key string) {
	if lc.discoveryService == nil {
		lc.logger.Errorf("[control] discovery service is not initialized")
		return
	}
	lc.discoveryService.RemovePeerMetadata(peerID, key)
}

// GetAllPeerInfo 获取所有已知节点的信息
//
// @return []*peer.AddrInfo 所有节点的地址信息列表
func (lc *Control) GetAllPeerInfo() []*peer.AddrInfo {
	if lc.discoveryService == nil {
		lc.logger.Errorf("[control] discovery service is not initialized")
		return []*peer.AddrInfo{}
	}
	return lc.discoveryService.GetAllPeerInfo()
}

// GetConnectedness 获取与指定节点的连接状态
//
// @param peerID peer.ID 节点ID
// @return network.Connectedness 连接状态
func (lc *Control) GetConnectedness(peerID peer.ID) network.Connectedness {
	if lc.discoveryService == nil {
		lc.logger.Errorf("[control] discovery service is not initialized")
		return network.NotConnected
	}
	return lc.discoveryService.GetConnectedness(peerID)
}
