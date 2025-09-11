package libp2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// 定义持久化数据结构
// type PersistentPeerInfo struct {
// 	PeerID        string                 `json:"peer_id"`
// 	Addrs         []string               `json:"addresses"`
// 	Metadata      map[string]interface{} `json:"metadata,omitempty"`
// 	IsBootstrap   bool                   `json:"is_bootstrap"`
// 	LastConnected time.Time              `json:"last_connected"`
// }

// DiscoveryService 节点发现服务
type DiscoveryService struct {
	ctx    context.Context
	host   host.Host
	logger *zap.SugaredLogger

	serviceTag string

	// 发现到的节点
	peers      map[peer.ID]struct{}
	peersMutex sync.RWMutex

	// 发现结果回调
	onPeerFound func(peer.ID, peer.AddrInfo)

	// bootstrap节点列表
	bootstrapPeers map[peer.ID]struct{}
	bootstrapMutex sync.RWMutex

	// DHT服务
	dht *dht.IpfsDHT

	// 配置
	libp2pConfig *config.Libp2pConfig

	// peerstore互斥锁
	peerstoreMutex sync.RWMutex

	// 持久化相关字段
	// persistenceFilePath string
	// saveInterval        time.Duration
	// persistenceEnabled  bool
}

// newDiscoveryService 创建一个新的节点发现服务
//
// @param ctx context.Context 上下文对象
// @param host host.Host libp2p主机实例
// @param serviceTag string 服务标签，用于mDNS发现
// @param logger *zap.SugaredLogger 日志记录器
// @param libp2pConfig *config.Libp2pConfig libp2p配置
//
// @return *DiscoveryService 发现服务实例
// @return error 错误信息，如果创建失败则返回错误
func newDiscoveryService(ctx context.Context, host host.Host, serviceTag string, logger *zap.SugaredLogger, libp2pConfig *config.Libp2pConfig) (*DiscoveryService, error) {
	logger.Infof("[discovery] creating discovery service...")
	ds := &DiscoveryService{
		ctx:            ctx,
		host:           host,
		logger:         logger,
		serviceTag:     serviceTag,
		peers:          make(map[peer.ID]struct{}),
		bootstrapPeers: make(map[peer.ID]struct{}),
		libp2pConfig:   libp2pConfig,
		onPeerFound: func(id peer.ID, info peer.AddrInfo) {
			// 默认回调，只是记录日志
			logger.Debugf("[discovery] discovered peer: %s", id)
		},
		// 持久化配置
		// persistenceEnabled:  true,
		// saveInterval:        5 * time.Minute, // 默认每5分钟保存一次
		// persistenceFilePath: filepath.Join("data", "peers.json"),
	}
	return ds, nil
}

// Start 启动节点发现服务
//
// 该方法初始化并启动mDNS服务和定期扫描协程
// @return error 错误信息，如果启动失败则返回错误
func (ds *DiscoveryService) Start() error {
	ds.logger.Infof("[discovery] starting discovery service...")
	// 创建mDNS服务
	service := mdns.NewMdnsService(
		ds.host,
		ds.serviceTag,
		ds,
	)

	if err := service.Start(); err != nil {
		ds.logger.Errorf("[discovery] starting mdns service: %v", err)
		return err
	}

	// 启动定期扫描
	go ds.scanPeers()

	// 如果启用了持久化，加载已保存的节点信息
	// if ds.persistenceEnabled {
	// 	if err := ds.LoadPeersFromFile(); err != nil {
	// 		ds.logger.Warnf("[discovery] failed to load peers from file, will continue without persisted data: %v", err)
	// 	}
	//
	// 	// 启动定期保存
	// 	go ds.periodicSavePeers()
	// }

	ds.logger.Debugf("[discovery] discovery service started with tag: %s", ds.serviceTag)
	return nil
}

// InitDHT 初始化DHT服务
//
// 该方法创建并配置分布式哈希表(DHT)服务，用于节点发现和数据索引
// @return error 错误信息，如果初始化失败则返回错误
func (ds *DiscoveryService) InitDHT() error {
	ds.logger.Infof("[discovery] initializing DHT service...")
	dhtOpts := []dht.Option{
		dht.Mode(dht.ModeServer),
		// dht.BootstrapPeers(),
	}

	// 如果有bootstrap节点，添加到DHT选项中
	var bootstrapAddrInfos []peer.AddrInfo
	if len(ds.libp2pConfig.BootstrapList) > 0 {
		for _, addrStr := range ds.libp2pConfig.BootstrapList {
			maddr, err := multiaddr.NewMultiaddr(addrStr)
			if err != nil {
				ds.logger.Warnf("[discovery] failed to parse bootstrap address %s: %v", addrStr, err)
				continue
			}

			info, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				ds.logger.Warnf("[discovery] failed to parse peer info from %s: %v", addrStr, err)
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
		ds.logger.Errorf("[discovery] failed to create DHT service: %v", err)
		return err
	}

	ds.dht = dhtService
	ds.logger.Infof("[discovery] DHT service initialized successfully")
	return nil
}

// BootstrapDHT 引导DHT服务
//
// 该方法连接到引导节点以引导DHT网络
// @return error 错误信息，如果引导失败则返回错误
func (ds *DiscoveryService) BootstrapDHT() error {
	if ds.dht == nil {
		return fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[discovery] bootstrapping DHT...")
	if err := ds.dht.Bootstrap(ds.ctx); err != nil {
		ds.logger.Errorf("failed to bootstrap DHT: %v", err)
		return err
	}
	return nil
}

// ConnectBootstrapPeers 连接bootstrap节点列表中的所有节点
//
// 该方法异步连接到配置中指定的所有bootstrap节点
func (ds *DiscoveryService) ConnectBootstrapPeers() {
	if len(ds.libp2pConfig.BootstrapList) == 0 {
		ds.logger.Infof("[discovery] no bootstrap peers configured")
		return
	}

	ds.logger.Infof("[discovery] connecting to bootstrap peers, count: %d", len(ds.libp2pConfig.BootstrapList))

	for _, addr := range ds.libp2pConfig.BootstrapList {
		ds.logger.Debugf("[discovery] connecting to bootstrap peer: %s", addr)
		go func(addr string) {
			// 解析multiaddr
			maddr, err := multiaddr.NewMultiaddr(addr)
			if err != nil {
				ds.logger.Errorf("[discovery] failed to parse multiaddr: %v", err)
				return
			}

			// 解析peer信息
			peerInfo, err := peer.AddrInfoFromP2pAddr(maddr)
			if err != nil {
				ds.logger.Errorf("[discovery] failed to parse peer info: %v", err)
				return
			}

			// 连接到节点
			if err := ds.host.Connect(ds.ctx, *peerInfo); err != nil {
				ds.logger.Errorf("[discovery] failed to connect to bootstrap peer %s: %v", addr, err)
				return
			}

			// 注册peer
			ds.peersMutex.Lock()
			ds.peers[peerInfo.ID] = struct{}{}
			ds.peersMutex.Unlock()

			// 添加到bootstrap节点列表
			ds.bootstrapMutex.Lock()
			ds.bootstrapPeers[peerInfo.ID] = struct{}{}
			ds.bootstrapMutex.Unlock()

			ds.logger.Infof("[discovery] successfully connected to bootstrap peer: %s", peerInfo.ID)
		}(addr)
	}
}

// StartHealthCheck 启动节点健康检查
//
// 该方法启动定期检查节点连接状态的协程
func (ds *DiscoveryService) StartHealthCheck() {
	ds.logger.Infof("[discovery] starting peer health check...")
	checkInterval := 60 * time.Second // 每分钟检查一次

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			ds.logger.Infof("[discovery] stopping peer health check...")
			return
		case <-ticker.C:
			ds.CheckPeersHealth()
		}
	}
}

// CheckPeersHealth 检查所有节点的健康状态
//
// 该方法遍历所有已知节点并异步检查它们的连接状态
func (ds *DiscoveryService) CheckPeersHealth() {
	ds.logger.Debugf("[discovery] checking peers health...")

	// 检查当前连接的所有节点
	ds.peersMutex.RLock()
	peersCopy := make([]peer.ID, 0, len(ds.peers))
	for peerID := range ds.peers {
		peersCopy = append(peersCopy, peerID)
	}
	ds.peersMutex.RUnlock()

	for _, peerID := range peersCopy {
		go ds.CheckPeerHealth(peerID)
	}
}

// CheckPeerHealth 检查单个节点的健康状态
//
// @param peerID peer.ID 要检查的节点ID
func (ds *DiscoveryService) CheckPeerHealth(peerID peer.ID) {
	ds.logger.Debugf("[discovery] checking health for peer %s", peerID)

	// 检查连接状态
	conn := ds.host.Network().Connectedness(peerID)
	if conn != network.Connected {
		ds.logger.Warnf("[discovery] peer %s is not connected, removing...", peerID)

		// 从节点列表中移除
		ds.peersMutex.Lock()
		delete(ds.peers, peerID)
		ds.peersMutex.Unlock()

		// 检查是否是bootstrap节点，如果是，则尝试重新连接
		ds.bootstrapMutex.RLock()
		_, isBootstrap := ds.bootstrapPeers[peerID]
		ds.bootstrapMutex.RUnlock()

		if isBootstrap {
			ds.logger.Infof("[discovery] peer %s is bootstrap node, will try to reconnect in next health check", peerID)
		}
	}
}

// GetBootstrapPeers 获取所有bootstrap节点
//
// @return []peer.ID bootstrap节点ID列表
func (ds *DiscoveryService) GetBootstrapPeers() []peer.ID {
	ds.bootstrapMutex.RLock()
	defer ds.bootstrapMutex.RUnlock()

	peers := make([]peer.ID, 0, len(ds.bootstrapPeers))
	for peerID := range ds.bootstrapPeers {
		peers = append(peers, peerID)
	}

	return peers
}

// Stop 停止节点发现服务
func (ds *DiscoveryService) Stop() error {
	// if ds.mdns != nil {
	// 	if err := ds.mdns.Close(); err != nil {
	// 		ds.logger.Errorf("Failed to close mDNS service: %v", err)
	// 		return err
	// 	}
	// }
	// 关闭DHT服务
	if ds.dht != nil {
		ds.logger.Debugf("[discovery] closing DHT service...")
		if err := ds.dht.Close(); err != nil {
			ds.logger.Errorf("[discovery] failed to close DHT service: %v", err)
			return err
		}
	}
	ds.logger.Infof("[discovery] discovery service stopped...")
	return nil
}

// FindPeer 使用DHT查找指定ID的节点
func (ds *DiscoveryService) FindPeer(peerID peer.ID) (*peer.AddrInfo, error) {
	if ds.dht == nil {
		return nil, fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[discovery] finding peer %s via DHT", peerID)
	peerInfo, err := ds.dht.FindPeer(ds.ctx, peerID)
	if err != nil {
		ds.logger.Errorf("[discovery] failed to find peer %s: %v", peerID, err)
		return nil, err
	}

	ds.logger.Infof("[discovery] found peer %s with %d addresses", peerID, len(peerInfo.Addrs))
	return &peerInfo, nil
}

// GetDHTRoutingTableInfo 获取DHT路由表信息
//
// @return int 路由表中的节点数量
// @return error 错误信息，如果DHT未初始化则返回错误
func (ds *DiscoveryService) GetDHTRoutingTableInfo() (int, error) {
	if ds.dht == nil {
		return 0, fmt.Errorf("DHT service is not initialized")
	}

	routingTable := ds.dht.RoutingTable()
	peerCount := routingTable.Size()

	ds.logger.Infof("[discovery] DHT routing table contains %d peers", peerCount)
	return peerCount, nil
}

// Provide 使用DHT提供数据索引
func (ds *DiscoveryService) Provide(key string) error {
	if ds.dht == nil {
		return fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[discovery] providing key %s to DHT", key)
	cidKey, err := cid.Cast([]byte(key))
	if err != nil {
		return err
	}
	if err := ds.dht.Provide(ds.ctx, cidKey, true); err != nil {
		ds.logger.Errorf("[discovery] failed to provide key %s: %v", key, err)
		return err
	}

	ds.logger.Infof("[discovery] successfully provided key %s to DHT", key)
	return nil
}

// FindProviders 使用DHT查找提供指定数据的节点
//
// @param key string 要查找的索引键
// @param count int 要查找的最大节点数量
// @return []peer.AddrInfo 提供该数据的节点列表
// @return error 错误信息，如果查找失败则返回错误
func (ds *DiscoveryService) FindProviders(key string, count int) ([]peer.AddrInfo, error) {
	if ds.dht == nil {
		return nil, fmt.Errorf("DHT service is not initialized")
	}

	ds.logger.Infof("[discovery] finding providers for key %s", key)

	cidKey, err := cid.Cast([]byte(key))
	if err != nil {
		return nil, err
	}
	providers := ds.dht.FindProvidersAsync(ds.ctx, cidKey, count)

	var result []peer.AddrInfo
	for p := range providers {
		result = append(result, p)
	}

	ds.logger.Infof("[discovery] found %d providers for key %s", len(result), key)
	return result, nil
}

// NotifyPeerFound 设置节点发现回调
func (ds *DiscoveryService) NotifyPeerFound(callback func(peer.ID, peer.AddrInfo)) {
	ds.logger.Infof("[discovery] register custom callback func...")
	ds.onPeerFound = callback
}

// GetDiscoveredPeers 获取所有发现的节点
//
// @return []peer.ID 已发现的节点ID列表
func (ds *DiscoveryService) GetDiscoveredPeers() []peer.ID {
	ds.logger.Infof("[discovery] get discovered peers...")
	ds.peersMutex.RLock()
	defer ds.peersMutex.RUnlock()

	peers := make([]peer.ID, 0, len(ds.peers))
	for peerID := range ds.peers {
		peers = append(peers, peerID)
	}

	return peers
}

// scanPeers 定期扫描局域网内的节点
func (ds *DiscoveryService) scanPeers() {
	ds.logger.Debugf("[discovery] starting scan peers...")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.logger.Debugf("[discovery] scanning for peers...")
			// mDNS服务会自动处理节点发现，这里可以添加额外的扫描逻辑
		}
	}
}

// HandlePeerFound 处理发现的节点
func (ds *DiscoveryService) HandlePeerFound(pi peer.AddrInfo) {
	ds.logger.Debugf("[discovery] handle peer found...")
	// 忽略自己
	if pi.ID == ds.host.ID() {
		return
	}

	// 检查是否已经发现过该节点
	ds.peersMutex.RLock()
	_, exists := ds.peers[pi.ID]
	ds.peersMutex.RUnlock()

	if !exists {
		// 添加到已发现节点列表
		ds.peersMutex.Lock()
		ds.peers[pi.ID] = struct{}{}
		ds.peersMutex.Unlock()

		ds.logger.Debugf("[discovery] discovered new peer: %s", pi.ID)

		// 尝试连接到发现的节点
		go func() {
			if err := ds.host.Connect(ds.ctx, pi); err != nil {
				ds.logger.Warnf("[discovery] failed to connect to discovered peer %s: %v", pi.ID, err)
			} else {
				ds.logger.Infof("[discovery] successfully connected to peer: %s", pi.ID)
				// 调用回调函数
				if ds.onPeerFound != nil {
					ds.onPeerFound(pi.ID, pi)
				}
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
func (ds *DiscoveryService) GetPeerInfo(peerID peer.ID) *peer.AddrInfo {
	ds.logger.Debugf("[peerstore] getting peer info for %s", peerID)
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
func (ds *DiscoveryService) AddPeerAddresses(peerID peer.ID, addrs []multiaddr.Multiaddr, ttl time.Duration) {
	ds.logger.Debugf("[peerstore] adding %d addresses for peer %s", len(addrs), peerID)
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
func (ds *DiscoveryService) SetPeerMetadata(peerID peer.ID, key string, value interface{}) error {
	ds.logger.Debugf("[peerstore] setting metadata for peer %s, key: %s", peerID, key)
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
func (ds *DiscoveryService) GetPeerMetadata(peerID peer.ID, key string) (interface{}, error) {
	ds.logger.Debugf("[peerstore] getting metadata for peer %s, key: %s", peerID, key)
	ds.peerstoreMutex.RLock()
	defer ds.peerstoreMutex.RUnlock()

	// 获取节点元数据
	return ds.host.Peerstore().Get(peerID, key)
}

// RemovePeerMetadata 移除节点的元数据

// periodicSavePeers 定期保存节点信息到文件
func (ds *DiscoveryService) periodicSavePeers() {
	ds.logger.Info("[discovery] starting periodic peer persistence service")
	// ticker := time.NewTicker(ds.saveInterval)
	// defer ticker.Stop()

	// for {
	// 	select {
	// 	case <-ds.ctx.Done():
	// 		ds.logger.Info("[discovery] stopping periodic peer persistence service")
	// 		// 在退出前最后保存一次
	// 		ds.SavePeersToFile()
	// 		return
	// 	case <-ticker.C:
	// 		if err := ds.SavePeersToFile(); err != nil {
	// 			ds.logger.Errorf("[discovery] failed to save peers to file: %v", err)
	// 		}
	// 	}
	// }
}

// SavePeersToFile 将节点信息保存到文件
func (ds *DiscoveryService) SavePeersToFile() error {
	ds.logger.Debug("[discovery] saving peers to file")

	// 准备持久化数据
	// peersInfo := []PersistentPeerInfo{}

	// 获取所有发现的节点
	ds.peersMutex.RLock()
	ds.peerstoreMutex.RLock()
	ds.bootstrapMutex.RLock()
	defer func() {
		ds.peersMutex.RUnlock()
		ds.peerstoreMutex.RUnlock()
		ds.bootstrapMutex.RUnlock()
	}()

	// 为每个节点准备持久化数据
	for peerID := range ds.peers {
		// 检查节点是否有效
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
		// _, isBootstrap := ds.bootstrapPeers[peerID]

		// 添加到持久化数据列表
		// peersInfo = append(peersInfo, PersistentPeerInfo{
		// 	PeerID:        peerID.String(),
		// 	Addrs:         addrStrings,
		// 	Metadata:      metadata,
		// 	IsBootstrap:   isBootstrap,
		// 	LastConnected: time.Now(),
		// })
	}

	// 创建数据目录（如果不存在）
	// dir := filepath.Dir(ds.persistenceFilePath)
	// if err := os.MkdirAll(dir, 0755); err != nil {
	// 	return fmt.Errorf("failed to create directory for peer persistence: %w", err)
	// }

	// 序列化数据
	// jsonData, err := json.MarshalIndent(peersInfo, "", "  ")
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal peer data: %w", err)
	// }

	// 写入文件
	// if err := ioutil.WriteFile(ds.persistenceFilePath, jsonData, 0644); err != nil {
	// 	return fmt.Errorf("failed to write peer data to file: %w", err)
	// }

	// ds.logger.Infof("[discovery] successfully saved %d peers to file: %s", len(peersInfo), ds.persistenceFilePath)
	return nil
}

// LoadPeersFromFile 从文件加载节点信息
func (ds *DiscoveryService) LoadPeersFromFile() error {
	ds.logger.Info("[discovery] loading peers from file")

	// 检查文件是否存在
	// if _, err := os.Stat(ds.persistenceFilePath); os.IsNotExist(err) {
	// 	ds.logger.Info("[discovery] no peer persistence file found, skipping load")
	// 	return nil
	// }

	// 读取文件内容
	// jsonData, err := ioutil.ReadFile(ds.persistenceFilePath)
	// if err != nil {
	// 	return fmt.Errorf("failed to read peer data from file: %w", err)
	// }

	// 反序列化数据
	// var peersInfo []PersistentPeerInfo
	// if err := json.Unmarshal(jsonData, &peersInfo); err != nil {
	// 	return fmt.Errorf("failed to unmarshal peer data: %w", err)
	// }

	// ds.logger.Infof("[discovery] loaded %d peers from file", len(peersInfo))

	// 处理加载的节点信息
	ds.peersMutex.Lock()
	ds.peerstoreMutex.Lock()
	ds.bootstrapMutex.Lock()
	defer func() {
		ds.peersMutex.Unlock()
		ds.peerstoreMutex.Unlock()
		ds.bootstrapMutex.Unlock()
	}()

	// for _, info := range peersInfo {
	// 解析Peer ID
	// peerID, err := peer.Decode(info.PeerID)
	// if err != nil {
	// 	ds.logger.Warnf("[discovery] failed to decode peer ID %s: %v", info.PeerID, err)
	// 	continue
	// }

	// 忽略自己
	// if peerID == ds.host.ID() {
	// 	continue
	// }

	// 解析地址
	// addrs := make([]multiaddr.Multiaddr, 0, len(info.Addrs))
	// for _, addrStr := range info.Addrs {
	// 	addr, err := multiaddr.NewMultiaddr(addrStr)
	// 	if err != nil {
	// 		ds.logger.Warnf("[discovery] failed to parse address %s: %v", addrStr, err)
	// 		continue
	// 	}
	// 	addrs = append(addrs, addr)
	// }

	// 添加节点信息到peerstore
	// if len(addrs) > 0 {
	// 	ds.host.Peerstore().AddAddrs(peerID, addrs, peerstore.PermanentAddrTTL)

	// 添加节点到发现列表
	// ds.peers[peerID] = struct{}{}

	// 如果是bootstrap节点，添加到bootstrap列表
	// if info.IsBootstrap {
	// 	ds.bootstrapPeers[peerID] = struct{}{}
	// }

	// ds.logger.Debugf("[discovery] loaded peer %s with %d addresses", peerID, len(addrs))
	// }
	// }

	ds.logger.Info("[discovery] peer loading completed")
	return nil
}

// @param peerID peer.ID 节点ID
// @param key string 元数据键
func (ds *DiscoveryService) RemovePeerMetadata(peerID peer.ID, key string) {
	ds.logger.Debugf("[peerstore] removing metadata for peer %s, key: %s", peerID, key)
	ds.peerstoreMutex.Lock()
	defer ds.peerstoreMutex.Unlock()

	// 移除节点元数据
	ds.host.Peerstore().RemovePeer(peerID)
}

// GetAllPeerInfo 获取所有已知节点的信息
//
// @return []*peer.AddrInfo 所有节点的地址信息列表
func (ds *DiscoveryService) GetAllPeerInfo() []*peer.AddrInfo {
	ds.logger.Debugf("[peerstore] getting all peer info")
	ds.peerstoreMutex.RLock()
	defer ds.peerstoreMutex.RUnlock()

	// 获取所有已知节点ID
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
func (ds *DiscoveryService) GetConnectedness(peerID peer.ID) network.Connectedness {
	ds.logger.Debugf("[peerstore] getting connectedness for peer %s", peerID)
	ds.peerstoreMutex.RLock()
	defer ds.peerstoreMutex.RUnlock()

	// 获取连接状态
	return ds.host.Network().Connectedness(peerID)
}
