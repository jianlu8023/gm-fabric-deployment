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
}

// newDiscoveryService 创建一个新的节点发现服务
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
	}
	return ds, nil
}

// Start 启动节点发现服务
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

	ds.logger.Debugf("[discovery] discovery service started with tag: %s", ds.serviceTag)
	return nil
}

// InitDHT 初始化DHT服务
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
