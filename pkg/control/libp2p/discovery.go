package libp2p

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

type DiscoveryService interface {
	InitDHT() error
	Start() error
	BootstrapDHT() error
	ConnectBootstrapPeers()
	StartHealthCheck()
	ListPeers() []peer.ID
	HasPeer(peerID peer.ID) bool
	AddPeerAddresses(peerID peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration)
	RemovePeer(peerID peer.ID)
	AddPeer(peerID peer.ID)
	GetBootstrapPeers() []peer.ID
	FindPeer(peerID peer.ID) (*peer.AddrInfo, error)
	GetDHTRoutingTableInfo() (int, error)
	Provide(key string) error
	GetPeerInfo(peerID peer.ID) *peer.AddrInfo
	FindProviders(key string, count int) ([]peer.AddrInfo, error)
	SetPeerMetadata(peerID peer.ID, key string, value interface{}) error
	GetPeerMetadata(peerID peer.ID, key string) (interface{}, error)
	GetAllPeerInfo() []*peer.AddrInfo
	GetConnectedness(peerID peer.ID) network.Connectedness
	Stop() error
}

// discoveryServiceImpl 节点发现服务
//
// @description 负责节点发现（mDNS + DHT）、节点管理（PeerStore）、持久化恢复和健康检查
// @struct
type discoveryServiceImpl struct {
	ctx    context.Context
	host   host.Host
	logger *zap.SugaredLogger

	serviceTag string

	// 发现到的节点
	peers concurrent.Map[peer.ID, struct{}]

	// bootstrap节点列表，保存peerID到原始地址的映射（用于断线重连）
	bootstrapPeers concurrent.Map[peer.ID, []multiaddr.Multiaddr]

	// DHT服务
	dht *dht.IpfsDHT

	// mDNS服务
	mdnsService mdns.Service

	// 配置
	libp2pConfig *config.Libp2pConfig

	// peerstore互斥锁，保护PeerStore读写操作
	peerstoreMutex sync.RWMutex

	// 持久化互斥锁，保护持久化文件读写操作
	persistMutex sync.Mutex

	// 持久化相关字段
	persistentConfig *PersistenceConfig
}

// newDiscoveryService 创建一个新的节点发现服务
//
// @param ctx context.Context 上下文对象
// @param host host.Host libp2p主机实例
// @param serviceTag string 服务标签，用于mDNS发现
// @param logger *zap.SugaredLogger 日志记录器
// @param libp2pConfig *config.Libp2pConfig libp2p配置
//
// @return *discoveryServiceImpl 发现服务实例
// @return error 错误信息，如果创建失败则返回错误
func newDiscoveryService(ctx context.Context, host host.Host, serviceTag string, logger *zap.SugaredLogger, libp2pConfig *config.Libp2pConfig) (DiscoveryService, error) {
	logger.Infof("[libp2p/discovery] creating discovery service...")

	// 初始化持久化配置
	persistentConfig := &PersistenceConfig{
		Enabled:      persistenceEnabled,
		FilePath:     persistenceFilePath,
		SaveInterval: saveInterval,
	}

	if libp2pConfig.Persistence != nil {
		persistentConfig.Enabled = libp2pConfig.Persistence.Enabled
		if libp2pConfig.Persistence.FilePath != "" {
			persistentConfig.FilePath = libp2pConfig.Persistence.FilePath
		}
		if libp2pConfig.Persistence.SaveInterval > 0 {
			persistentConfig.SaveInterval = time.Duration(libp2pConfig.Persistence.SaveInterval) * time.Second
		}
	}

	ds := &discoveryServiceImpl{
		ctx:              ctx,
		host:             host,
		logger:           logger,
		serviceTag:       serviceTag,
		peers:            concurrentmap.NewRWMap[peer.ID, struct{}](),
		bootstrapPeers:   concurrentmap.NewRWMap[peer.ID, []multiaddr.Multiaddr](),
		libp2pConfig:     libp2pConfig,
		persistentConfig: persistentConfig,
	}
	return ds, nil
}

// Start 启动节点发现服务
//
// @description 初始化并启动mDNS服务；注册网络连接通知器（Notifee），任何来源的节点连接建立时自动加入peers；
// 如果启用了持久化，从文件加载已保存的节点信息并启动定期保存协程
// @return error 错误信息，如果启动失败则返回错误
func (ds *discoveryServiceImpl) Start() error {
	ds.logger.Infof("[libp2p/discovery] starting discovery service...")

	// 如果启用了持久化，先加载已保存的节点信息
	if ds.persistentConfig.Enabled {
		if err := ds.LoadPeersFromFile(); err != nil {
			ds.logger.Warnf("[libp2p/discovery] failed to load peers from file, will continue without persisted data: %v", err)
		}
	}

	// 注册网络连接通知器：任何来源（mDNS/DHT/bootstrap/手动）的节点连接建立时自动加入peers
	ds.host.Network().Notify(&discoveryNotifee{ds: ds})

	// 创建mDNS服务
	ds.mdnsService = mdns.NewMdnsService(
		ds.host,
		ds.serviceTag,
		ds,
	)

	if err := ds.mdnsService.Start(); err != nil {
		ds.logger.Errorf("[libp2p/discovery] starting mdns service: %v", err)
		return err
	}

	// 如果启用了持久化，启动定期保存协程
	if ds.persistentConfig.Enabled {
		go ds.periodicSavePeers()
	}

	ds.logger.Debugf("[libp2p/discovery] discovery service started with tag: %s", ds.serviceTag)
	return nil
}

// InitDHT 初始化DHT服务
//
// @description 创建并配置分布式哈希表(DHT)服务，用于节点发现和数据索引。
// DHT模式从配置读取（server/client/auto），默认server。
// @return error 错误信息，如果初始化失败则返回错误
func (ds *discoveryServiceImpl) InitDHT() error {
	ds.logger.Infof("[libp2p/discovery] initializing DHT service...")

	// DHT模式从配置读取，默认server
	mode := dht.ModeServer
	switch ds.libp2pConfig.DHTMode {
	case "client":
		mode = dht.ModeClient
	case "auto":
		mode = dht.ModeAuto
	}

	dhtOpts := []dht.Option{dht.Mode(mode)}

	// 如果有bootstrap节点，添加到DHT选项中
	var bootstrapAddrInfos []peer.AddrInfo
	if len(ds.libp2pConfig.BootstrapList) > 0 {
		for _, addrStr := range ds.libp2pConfig.BootstrapList {
			maddr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				ds.logger.Warnf("[libp2p/discovery] failed to parse bootstrap address %s: %v", addrStr, err)
				continue
			}

			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				ds.logger.Warnf("[libp2p/discovery] failed to parse peer info from %s: %v", addrStr, err)
				continue
			}
			bootstrapAddrInfos = append(bootstrapAddrInfos, *info)
		}

		if len(bootstrapAddrInfos) > 0 {
			dhtOpts = append(dhtOpts, dht.BootstrapPeers(bootstrapAddrInfos...))
		}
	} else {
		dhtOpts = append(dhtOpts, dht.BootstrapPeers())
	}

	// 创建DHT服务
	dhtService, err := dht.New(ds.ctx, ds.host, dhtOpts...)
	if err != nil {
		ds.logger.Errorf("[libp2p/discovery] failed to create DHT service: %v", err)
		return err
	}

	ds.dht = dhtService
	ds.logger.Infof("[libp2p/discovery] DHT service initialized successfully")
	return nil
}

// BootstrapDHT 引导DHT服务
//
// 该方法连接到引导节点以引导DHT网络
// @return error 错误信息，如果引导失败则返回错误
func (ds *discoveryServiceImpl) BootstrapDHT() error {
	if ds.dht == nil {
		return fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[libp2p/discovery] bootstrapping DHT...")
	if err := ds.dht.Bootstrap(ds.ctx); err != nil {
		ds.logger.Errorf("[libp2p/discovery] failed to bootstrap DHT: %v", err)
		return err
	}
	return nil
}

// ConnectBootstrapPeers 连接bootstrap节点列表中的所有节点
//
// @description 异步并发连接到配置中指定的所有bootstrap节点，每个节点带3次指数退避重试，
// 连接成功后将地址保存到bootstrapPeers供健康检查重连使用
func (ds *discoveryServiceImpl) ConnectBootstrapPeers() {
	if len(ds.libp2pConfig.BootstrapList) == 0 {
		ds.logger.Infof("[libp2p/discovery] no bootstrap peers configured")
		return
	}

	ds.logger.Infof("[libp2p/discovery] connecting to bootstrap peers, count: %d", len(ds.libp2pConfig.BootstrapList))

	for _, addr := range ds.libp2pConfig.BootstrapList {
		go ds.connectBootstrapPeer(addr)
	}
}

// connectBootstrapPeer 连接单个bootstrap节点，带指数退避重试
//
// @param addr string bootstrap节点的multiaddr字符串
func (ds *discoveryServiceImpl) connectBootstrapPeer(addr string) {
	ds.logger.Debugf("[libp2p/discovery] connecting to bootstrap peer: %s", addr)

	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		ds.logger.Errorf("[libp2p/discovery] failed to parse multiaddr %s: %v", addr, err)
		return
	}

	// 解析peer信息
	peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		ds.logger.Errorf("[libp2p/discovery] failed to parse peer info from %s: %v", addr, err)
		return
	}

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ds.ctx.Done():
			return
		case <-time.After(time.Duration(1<<uint(i)) * time.Second):
		}
		if err := ds.host.Connect(ds.ctx, *peerInfo); err != nil {
			ds.logger.Warnf("[libp2p/discovery] connect attempt %d/%d failed for bootstrap peer %s: %v", i+1, maxRetries, addr, err)
		} else {
			ds.peers.Put(peerInfo.ID, struct{}{})
			ds.bootstrapPeers.Put(peerInfo.ID, peerInfo.Addrs)
			ds.logger.Infof("[libp2p/discovery] successfully connected to bootstrap peer: %s", peerInfo.ID)
			return
		}
	}
	ds.logger.Errorf("[libp2p/discovery] failed to connect to bootstrap peer %s after %d attempts", addr, maxRetries)
}

// StartHealthCheck 启动节点健康检查
//
// @description 按配置的间隔时间定期检查节点连接状态，间隔可通过 Libp2pConfig.HealthCheckInterval 配置（秒），默认60秒
func (ds *discoveryServiceImpl) StartHealthCheck() {
	ds.logger.Infof("[libp2p/discovery] starting peer health check...")

	// 从配置获取检查间隔，默认60秒
	checkInterval := 60 * time.Second
	if ds.libp2pConfig.HealthCheckInterval > 0 {
		checkInterval = time.Duration(ds.libp2pConfig.HealthCheckInterval) * time.Second
	}
	ds.logger.Debugf("[libp2p/discovery] health check interval: %v", checkInterval)

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			ds.logger.Infof("[libp2p/discovery] stopping peer health check...")
			return
		case <-ticker.C:
			ds.CheckPeersHealth()
		}
	}
}

// CheckPeersHealth 检查所有节点的健康状态
//
// @description 遍历所有已知节点并异步检查它们的连接状态
func (ds *discoveryServiceImpl) CheckPeersHealth() {
	ds.logger.Debugf("[libp2p/discovery] checking peers health...")
	keys := ds.peers.Keys()
	for _, peerID := range keys {
		go ds.CheckPeerHealth(peerID)
	}
}

// CheckPeerHealth 检查单个节点的健康状态
//
// @description 检测到节点断线时从peers移除；如果是bootstrap节点，异步尝试重连
// @param peerID peer.ID 要检查的节点ID
func (ds *discoveryServiceImpl) CheckPeerHealth(peerID peer.ID) {
	ds.logger.Debugf("[libp2p/discovery] checking health for peer %s", peerID)

	conn := ds.host.Network().Connectedness(peerID)
	if conn != network.Connected {
		ds.logger.Warnf("[libp2p/discovery] peer %s is not connected, removing...", peerID)
		ds.peers.Del(peerID)

		// 检查是否是bootstrap节点，如果是，尝试重连
		addrs, isBootstrap := ds.bootstrapPeers.Get(peerID)
		if isBootstrap {
			ds.logger.Infof("[libp2p/discovery] bootstrap peer %s disconnected, trying to reconnect...", peerID)
			go ds.tryReconnect(peerID, addrs)
		}
	}
}

// GetBootstrapPeers 获取所有bootstrap节点
//
// @return []peer.ID bootstrap节点ID列表
func (ds *discoveryServiceImpl) GetBootstrapPeers() []peer.ID {
	return ds.bootstrapPeers.Keys()
}

// Stop 停止节点发现服务
//
// @description 如果启用了持久化，先保存节点信息到文件，然后关闭mDNS服务和DHT服务
// @return error 错误信息，如果关闭失败则返回错误
func (ds *discoveryServiceImpl) Stop() error {
	// 如果启用了持久化，停止前保存节点信息到文件
	if ds.persistentConfig.Enabled {
		ds.logger.Debugf("[libp2p/discovery] saving peers to file before stopping...")
		if err := ds.SavePeersToFile(); err != nil {
			ds.logger.Errorf("[libp2p/discovery] failed to save peers on stop: %v", err)
		}
	}

	// 关闭mDNS服务
	if ds.mdnsService != nil {
		ds.logger.Debugf("[libp2p/discovery] closing mDNS service...")
		if err := ds.mdnsService.Close(); err != nil {
			ds.logger.Errorf("[libp2p/discovery] failed to close mDNS service: %v", err)
		}
	}

	// 关闭DHT服务
	if ds.dht != nil {
		ds.logger.Debugf("[libp2p/discovery] closing DHT service...")
		if err := ds.dht.Close(); err != nil {
			ds.logger.Errorf("[libp2p/discovery] failed to close DHT service: %v", err)
			return err
		}
	}
	ds.logger.Infof("[libp2p/discovery] discovery service stopped...")
	return nil
}

// FindPeer 使用DHT查找指定ID的节点
func (ds *discoveryServiceImpl) FindPeer(peerID peer.ID) (*peer.AddrInfo, error) {
	if ds.dht == nil {
		return nil, fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[libp2p/discovery] finding peer %s via DHT", peerID)
	peerInfo, err := ds.dht.FindPeer(ds.ctx, peerID)
	if err != nil {
		ds.logger.Errorf("[libp2p/discovery] failed to find peer %s: %v", peerID, err)
		return nil, err
	}

	ds.logger.Infof("[libp2p/discovery] found peer %s with %d addresses", peerID, len(peerInfo.Addrs))
	return &peerInfo, nil
}

// GetDHTRoutingTableInfo 获取DHT路由表信息
//
// @return int 路由表中的节点数量
// @return error 错误信息，如果DHT未初始化则返回错误
func (ds *discoveryServiceImpl) GetDHTRoutingTableInfo() (int, error) {
	if ds.dht == nil {
		return 0, fmt.Errorf("DHT service is not initialized")
	}

	routingTable := ds.dht.RoutingTable()
	peerCount := routingTable.Size()

	ds.logger.Infof("[libp2p/discovery] DHT routing table contains %d peers", peerCount)
	return peerCount, nil
}

// Provide 使用DHT提供数据索引
//
// @description 将指定key作为CID提供到DHT网络，供其他节点通过FindProviders查找。
// key应为合法的CID字符串，若无法解析为CID则跳过并记录警告
// @param key string 数据索引键，应为合法的CID字符串
// @return error 错误信息，如果DHT未初始化或提供失败则返回错误
func (ds *discoveryServiceImpl) Provide(key string) error {
	if ds.dht == nil {
		return fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[libp2p/discovery] providing key %s to DHT", key)
	cidKey, err := cid.Cast([]byte(key))
	if err != nil {
		// key不是合法的CID，跳过
		ds.logger.Warnf("[libp2p/discovery] key %s is not a valid CID, skipping: %v", key, err)
		return nil
	}
	if err := ds.dht.Provide(ds.ctx, cidKey, true); err != nil {
		ds.logger.Errorf("[libp2p/discovery] failed to provide key %s: %v", key, err)
		return err
	}

	ds.logger.Infof("[libp2p/discovery] successfully provided key %s to DHT", key)
	return nil
}

// FindProviders 使用DHT查找提供指定数据的节点
//
// @description 通过DHT查找提供指定key的节点列表。key应为合法的CID字符串，
// 若无法解析为CID则跳过并返回空列表
// @param key string 要查找的索引键，应为合法的CID字符串
// @param count int 要查找的最大节点数量
// @return []peer.AddrInfo 提供该数据的节点列表
// @return error 错误信息，如果DHT未初始化则返回错误
func (ds *discoveryServiceImpl) FindProviders(key string, count int) ([]peer.AddrInfo, error) {
	if ds.dht == nil {
		return nil, fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[libp2p/discovery] finding providers for key %s", key)

	cidKey, err := cid.Cast([]byte(key))
	if err != nil {
		// key不是合法的CID，跳过
		ds.logger.Warnf("[libp2p/discovery] key %s is not a valid CID, skipping: %v", key, err)
		return nil, nil
	}
	providers := ds.dht.FindProvidersAsync(ds.ctx, cidKey, count)

	var result []peer.AddrInfo
	for p := range providers {
		result = append(result, p)
	}

	ds.logger.Infof("[libp2p/discovery] found %d providers for key %s", len(result), key)
	return result, nil
}

// GetDiscoveredPeers 获取所有发现的节点
//
// @return []peer.ID 已发现的节点ID列表
func (ds *discoveryServiceImpl) GetDiscoveredPeers() []peer.ID {
	ds.logger.Infof("[libp2p/discovery] get discovered peers...")
	return ds.peers.Keys()
}

// ListPeers 获取所有已发现的节点ID列表
//
// @description 为Control层提供受控的节点查询接口，避免Control层直接访问peers字段破坏分层
// @return []peer.ID 已发现的节点ID列表
func (ds *discoveryServiceImpl) ListPeers() []peer.ID {
	return ds.peers.Keys()
}

// HasPeer 检查节点是否已发现
//
// @param peerID peer.ID 节点ID
// @return bool 是否已发现
func (ds *discoveryServiceImpl) HasPeer(peerID peer.ID) bool {
	_, ok := ds.peers.Get(peerID)
	return ok
}

// AddPeer 添加节点到已发现列表
//
// @description 为Control层提供受控的节点添加接口，避免Control层直接访问peers字段破坏分层
// @param peerID peer.ID 节点ID
func (ds *discoveryServiceImpl) AddPeer(peerID peer.ID) {
	ds.peers.Put(peerID, struct{}{})
}

// RemovePeer 从已发现列表删除节点
//
// @description 为Control层提供受控的节点删除接口，避免Control层直接访问peers字段破坏分层
// @param peerID peer.ID 节点ID
func (ds *discoveryServiceImpl) RemovePeer(peerID peer.ID) {
	ds.peers.Del(peerID)
}

// HandlePeerFound 处理mDNS发现的节点
//
// @description mDNS发现同网段节点时回调，将节点加入peers集合和PeerStore，并异步尝试连接
// @param pi peer.AddrInfo 发现的节点地址信息
func (ds *discoveryServiceImpl) HandlePeerFound(pi peer.AddrInfo) {
	ds.logger.Debugf("[libp2p/discovery] handle peer found...")

	// 忽略自己
	if pi.ID == ds.host.ID() {
		return
	}

	_, exists := ds.peers.Get(pi.ID)
	if !exists {
		ds.peers.Put(pi.ID, struct{}{})
		ds.logger.Debugf("[libp2p/discovery] discovered new peer via mDNS: %s", pi.ID)

		// 尝试连接到发现的节点
		go func() {
			if err := ds.host.Connect(ds.ctx, pi); err != nil {
				ds.logger.Warnf("[libp2p/discovery] failed to connect to discovered peer %s: %v", pi.ID, err)
			} else {
				ds.logger.Infof("[libp2p/discovery] successfully connected to peer: %s", pi.ID)
			}
		}()
	}
}

// ===================== PeerStore 功能增强 =====================

// GetPeerInfo 获取指定节点的完整信息
//
// PeerStore是libp2p中的一个核心组件，用于存储和管理网络中节点的信息，包括：
// 1. 节点标识(Peer ID)
// 2. 网络地址(Addrs)
// 3. 支持的协议(Protocols)
// 4. 连接元数据(Metadata)
// 5. 证书信息(Certs)
//
// 通过PeerStore，节点可以高效地管理与其他节点的连接信息，无需每次连接时都重新解析和验证节点信息
//
// @param peerID peer.ID 目标节点的ID
// @return *peer.AddrInfo 节点的地址信息，如果节点不存在则返回nil
func (ds *discoveryServiceImpl) GetPeerInfo(peerID peer.ID) *peer.AddrInfo {
	ds.logger.Debugf("[libp2p/peerstore] getting peer info for %s", peerID)
	ds.peerstoreMutex.RLock()
	defer ds.peerstoreMutex.RUnlock()

	// 使用host内置的PeerStore获取节点地址信息
	addrs := ds.host.Peerstore().Addrs(peerID)
	if len(addrs) == 0 {
		return nil
	}

	return &peer.AddrInfo{
		ID:    peerID,
		Addrs: addrs,
	}
}

// AddPeerAddresses 向PeerStore中添加节点地址
//
// @param peerID peer.ID 节点ID
// @param addrs []multiaddr.Multiaddr 节点地址列表
// @param ttl time.Duration 地址的生存时间
func (ds *discoveryServiceImpl) AddPeerAddresses(peerID peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	ds.logger.Debugf("[libp2p/peerstore] adding %d addresses for peer %s", len(addrs), peerID)
	ds.peerstoreMutex.Lock()
	defer ds.peerstoreMutex.Unlock()

	// 将地址添加到PeerStore中，并设置TTL
	ds.host.Peerstore().AddAddrs(peerID, addrs, ttl)
}

// SetPeerMetadata 设置节点的元数据
//
// 元数据可以存储关于节点的额外信息，如节点类型、地理位置、性能指标等
//
// @param peerID peer.ID 节点ID
// @param key string 元数据键
// @param value interface{} 元数据值
func (ds *discoveryServiceImpl) SetPeerMetadata(peerID peer.ID, key string, value interface{}) error {
	ds.logger.Debugf("[libp2p/peerstore] setting metadata for peer %s, key: %s", peerID, key)
	ds.peerstoreMutex.Lock()
	defer ds.peerstoreMutex.Unlock()

	// 设置节点元数据
	return ds.host.Peerstore().Put(peerID, key, value)
}

// GetPeerMetadata 获取节点的元数据
//
// @param peerID peer.ID 节点ID
// @param key string 元数据键
// @return interface{} 元数据值
// @return error 错误信息
func (ds *discoveryServiceImpl) GetPeerMetadata(peerID peer.ID, key string) (interface{}, error) {
	ds.logger.Debugf("[libp2p/peerstore] getting metadata for peer %s, key: %s", peerID, key)
	ds.peerstoreMutex.RLock()
	defer ds.peerstoreMutex.RUnlock()

	// 获取节点元数据
	return ds.host.Peerstore().Get(peerID, key)
}

// RemovePeerMetadata 移除节点的元数据
// @param peerID peer.ID 节点ID
// @param key string 元数据键
// func (ds *discoveryServiceImpl) RemovePeerMetadata(peerID peer.ID, key string) {
// 	ds.logger.Debugf("[peerstore] removing metadata for peer %s, key: %s", peerID, key)
// 	ds.peerstoreMutex.Lock()
// 	defer ds.peerstoreMutex.Unlock()
//
// 	// 移除节点元数据
// 	ds.host.Peerstore().RemovePeer(peerID)
// }

// periodicSavePeers 定期保存节点信息到文件
//
// @description 按配置的间隔时间定期将节点信息保存到持久化文件，上下文取消时最后保存一次再退出
func (ds *discoveryServiceImpl) periodicSavePeers() {
	ds.logger.Infof("[libp2p/discovery] starting periodic peer persistence service, interval: %v", ds.persistentConfig.SaveInterval)
	ticker := time.NewTicker(ds.persistentConfig.SaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			ds.logger.Info("[libp2p/discovery] stopping periodic peer persistence service")
			// 在退出前最后保存一次
			if err := ds.SavePeersToFile(); err != nil {
				ds.logger.Errorf("[libp2p/discovery] failed to save peers on exit: %v", err)
			}
			return
		case <-ticker.C:
			if err := ds.SavePeersToFile(); err != nil {
				ds.logger.Errorf("[libp2p/discovery] failed to save peers to file: %v", err)
			}
		}
	}
}

// SavePeersToFile 将节点信息保存到文件
//
// @description 遍历当前所有已发现的节点，从PeerStore获取地址信息，序列化为JSON写入持久化文件
// @return error 错误信息，如果保存失败则返回错误
func (ds *discoveryServiceImpl) SavePeersToFile() error {
	ds.logger.Debugf("[libp2p/discovery] saving peers to file: %s", ds.persistentConfig.FilePath)

	ds.persistMutex.Lock()
	defer ds.persistMutex.Unlock()

	// 获取所有已发现的节点
	keys := ds.peers.Keys()
	peersInfo := make([]PersistentPeerInfo, 0, len(keys))

	for _, peerID := range keys {
		// 从PeerStore获取节点地址
		addrs := ds.host.Peerstore().Addrs(peerID)
		if len(addrs) == 0 {
			continue
		}

		// 转换地址为字符串
		addrStrings := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addrStrings = append(addrStrings, addr.String())
		}

		// 获取元数据（这里简化处理，获取所有元数据可能需要更复杂的逻辑）
		// metadata := make(map[string]interface{})
		// 在实际实现中，可以通过遍历已知的元数据键来获取值
		// 这里为了简化，暂时不包含元数据

		// 检查是否是bootstrap节点
		_, isBootstrap := ds.bootstrapPeers.Get(peerID)

		peersInfo = append(peersInfo, PersistentPeerInfo{
			PeerID:        peerID.String(),
			Addrs:         addrStrings,
			IsBootstrap:   isBootstrap,
			LastConnected: time.Now(),
		})
	}

	// 创建数据目录（如果不存在）
	dir := filepath.Dir(ds.persistentConfig.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for peer persistence: %w", err)
	}

	// 序列化数据
	jsonData, err := json.MarshalIndent(peersInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal peer data: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(ds.persistentConfig.FilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write peer data to file: %w", err)
	}

	ds.logger.Infof("[libp2p/discovery] successfully saved %d peers to file: %s", len(peersInfo), ds.persistentConfig.FilePath)
	return nil
}

// LoadPeersFromFile 从文件加载节点信息
//
// @description 从持久化文件读取节点信息，添加到PeerStore和peers集合，并异步尝试连接这些节点以恢复之前的节点状态
// @return error 错误信息，如果加载失败则返回错误
func (ds *discoveryServiceImpl) LoadPeersFromFile() error {
	ds.logger.Infof("[libp2p/discovery] loading peers from file: %s", ds.persistentConfig.FilePath)

	ds.persistMutex.Lock()
	defer ds.persistMutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(ds.persistentConfig.FilePath); os.IsNotExist(err) {
		ds.logger.Info("[libp2p/discovery] no peer persistence file found, skipping load")
		return nil
	}

	// 读取文件内容
	jsonData, err := os.ReadFile(ds.persistentConfig.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read peer data from file: %w", err)
	}

	// 反序列化数据
	var peersInfo []PersistentPeerInfo
	if err := json.Unmarshal(jsonData, &peersInfo); err != nil {
		return fmt.Errorf("failed to unmarshal peer data: %w", err)
	}

	ds.logger.Infof("[libp2p/discovery] loaded %d peers from file", len(peersInfo))

	loadedCount := 0
	for _, info := range peersInfo {
		// 解析Peer ID
		peerID, err := peer.Decode(info.PeerID)
		if err != nil {
			ds.logger.Warnf("[libp2p/discovery] failed to decode peer ID %s: %v", info.PeerID, err)
			continue
		}

		// 忽略自己
		if peerID == ds.host.ID() {
			continue
		}

		// 解析地址
		addrs := make([]multiaddr.Multiaddr, 0, len(info.Addrs))
		for _, addrStr := range info.Addrs {
			addr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				ds.logger.Warnf("[libp2p/discovery] failed to parse address %s: %v", addrStr, err)
				continue
			}
			addrs = append(addrs, addr)
		}

		// 添加节点信息到PeerStore和peers集合，并异步尝试连接
		if len(addrs) > 0 {
			ds.host.Peerstore().AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)
			ds.peers.Put(peerID, struct{}{})

			// 如果是bootstrap节点，添加到bootstrap列表并保存地址供重连
			if info.IsBootstrap {
				ds.bootstrapPeers.Put(peerID, addrs)
			}

			// 异步尝试连接，恢复之前的节点状态
			go func(pid peer.ID, pAddrs []multiaddr.Multiaddr) {
				addrInfo := peer.AddrInfo{ID: pid, Addrs: pAddrs}
				if err := ds.host.Connect(ds.ctx, addrInfo); err != nil {
					ds.logger.Warnf("[libp2p/discovery] failed to reconnect to persisted peer %s: %v", pid, err)
				} else {
					ds.logger.Infof("[libp2p/discovery] successfully reconnected to persisted peer: %s", pid)
				}
			}(peerID, addrs)

			loadedCount++
		}
	}

	ds.logger.Infof("[libp2p/discovery] peer loading completed, %d peers restored", loadedCount)
	return nil
}

// GetAllPeerInfo 获取所有已知节点的信息
//
// @description 不持有外层RLock，由GetPeerInfo各自加锁，避免RWMutex不可重入导致死锁
// @return []*peer.AddrInfo 所有节点的地址信息列表
func (ds *discoveryServiceImpl) GetAllPeerInfo() []*peer.AddrInfo {
	ds.logger.Debugf("[libp2p/peerstore] getting all peer info")
	peerIDs := ds.GetDiscoveredPeers()
	result := make([]*peer.AddrInfo, 0, len(peerIDs))

	// 为每个节点ID获取完整信息
	for _, peerID := range peerIDs {
		info := ds.GetPeerInfo(peerID)
		if info != nil {
			result = append(result, info)
		}
	}

	return result
}

// GetConnectedness 获取与指定节点的连接状态
//
// @param peerID peer.ID 节点ID
// @return network.Connectedness 连接状态
func (ds *discoveryServiceImpl) GetConnectedness(peerID peer.ID) network.Connectedness {
	ds.logger.Debugf("[libp2p/peerstore] getting connectedness for peer %s", peerID)
	ds.peerstoreMutex.RLock()
	defer ds.peerstoreMutex.RUnlock()

	// 获取连接状态
	return ds.host.Network().Connectedness(peerID)
}

// tryReconnect 尝试重连指定节点，使用指数退避策略
//
// @description 用于bootstrap节点断线重连，最多重试3次，每次间隔指数增长
// @param peerID peer.ID 目标节点ID
// @param addrs []multiaddr.Multiaddr 目标节点地址列表
func (ds *discoveryServiceImpl) tryReconnect(peerID peer.ID, addrs []multiaddr.Multiaddr) {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ds.ctx.Done():
			return
		case <-time.After(time.Duration(1<<uint(i)) * time.Second):
		}
		if err := ds.host.Connect(ds.ctx, peer.AddrInfo{ID: peerID, Addrs: addrs}); err != nil {
			ds.logger.Warnf("[libp2p/discovery] reconnect attempt %d/%d failed for peer %s: %v", i+1, maxRetries, peerID, err)
		} else {
			ds.logger.Infof("[libp2p/discovery] successfully reconnected to bootstrap peer: %s", peerID)
			ds.peers.Put(peerID, struct{}{})
			return
		}
	}
	ds.logger.Errorf("[libp2p/discovery] failed to reconnect to bootstrap peer %s after %d attempts", peerID, maxRetries)
}
