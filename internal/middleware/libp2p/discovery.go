package libp2p

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"go.uber.org/zap"
)

// DiscoveryService 局域网节点发现服务
type DiscoveryService struct {
	ctx        context.Context
	host       host.Host
	logger     *zap.SugaredLogger
	serviceTag string
	// mdns       *mdns.mdnsService

	// 发现到的节点
	peers      map[peer.ID]struct{}
	peersMutex sync.RWMutex

	// 发现结果回调
	onPeerFound func(peer.ID, peer.AddrInfo)
}

// newDiscoveryService 创建一个新的节点发现服务
func newDiscoveryService(ctx context.Context, host host.Host, serviceTag string, logger *zap.SugaredLogger) (*DiscoveryService, error) {
	logger.Infof("creating discovery service...")
	ds := &DiscoveryService{
		ctx:        ctx,
		host:       host,
		logger:     logger,
		serviceTag: serviceTag,
		peers:      make(map[peer.ID]struct{}),
		onPeerFound: func(id peer.ID, info peer.AddrInfo) {
			// 默认回调，只是记录日志
			logger.Debugf("discovered peer: %s", id)
		},
	}
	return ds, nil
}

// Start 启动节点发现服务
func (ds *DiscoveryService) Start() error {
	ds.logger.Infof("starting discovery service...")
	// 创建mDNS服务
	service := mdns.NewMdnsService(
		ds.host,
		ds.serviceTag,
		ds,
	)

	if err := service.Start(); err != nil {
		ds.logger.Errorf("starting mdns service: %v", err)
		return err
	}

	// 启动定期扫描
	go ds.scanPeers()

	ds.logger.Debugf("discovery service started with tag: %s", ds.serviceTag)
	return nil
}

// Stop 停止节点发现服务
func (ds *DiscoveryService) Stop() error {
	// if ds.mdns != nil {
	// 	if err := ds.mdns.Close(); err != nil {
	// 		ds.logger.Errorf("Failed to close mDNS service: %v", err)
	// 		return err
	// 	}
	// }
	ds.logger.Infof("discovery service stopped...")
	return nil
}

// NotifyPeerFound 设置节点发现回调
func (ds *DiscoveryService) NotifyPeerFound(callback func(peer.ID, peer.AddrInfo)) {
	ds.logger.Infof("register custom callback func...")
	ds.onPeerFound = callback
}

// GetDiscoveredPeers 获取所有发现的节点
func (ds *DiscoveryService) GetDiscoveredPeers() []peer.ID {
	ds.logger.Infof("get discovered peers...")
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
	ds.logger.Infof("starting scan peers...")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.logger.Debugf("scanning for peers...")
			// mDNS服务会自动处理节点发现，这里可以添加额外的扫描逻辑
		}
	}
}

// HandlePeerFound 处理发现的节点
func (ds *DiscoveryService) HandlePeerFound(pi peer.AddrInfo) {
	ds.logger.Debugf("handle peer found...")
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

		ds.logger.Debugf("discovered new peer: %s", pi.ID)

		// 尝试连接到发现的节点
		go func() {
			if err := ds.host.Connect(ds.ctx, pi); err != nil {
				ds.logger.Warnf("failed to connect to discovered peer %s: %v", pi.ID, err)
			} else {
				ds.logger.Infof("successfully connected to peer: %s", pi.ID)
				// 调用回调函数
				if ds.onPeerFound != nil {
					ds.onPeerFound(pi.ID, pi)
				}
			}
		}()
	}
}
