package libp2p

import (
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

// discoveryNotifee 网络连接通知器
//
// @description 实现network.Notifee接口，当任何来源（mDNS/DHT/bootstrap/手动连接）的节点建立连接时，
// 自动将peer加入peers集合，解决DHT发现节点不加入peers的问题（D4）
type discoveryNotifee struct {
	ds *discoveryServiceImpl
}

// Connected 当节点连接建立时触发，自动将peer加入peers集合
func (n *discoveryNotifee) Connected(_ network.Network, c network.Conn) {
	pid := c.RemotePeer()
	if pid == n.ds.host.ID() {
		return
	}
	if !n.ds.HasPeer(pid) {
		n.ds.AddPeer(pid)
		n.ds.logger.Debugf("[libp2p/notifee] peer connected and added to peers: %s", pid)
	}
}

// Disconnected 当节点断开连接时触发，由健康检查处理，此处不操作
func (n *discoveryNotifee) Disconnected(_ network.Network, _ network.Conn) {
	n.ds.logger.Debugf("[libp2p/notifee] disconnetced...")
}

// Listen 监听地址开始监听时触发
func (n *discoveryNotifee) Listen(_ network.Network, _ multiaddr.Multiaddr) {
	n.ds.logger.Debugf("[libp2p/notifee] listen...")
}

// ListenClose 监听地址停止监听时触发
func (n *discoveryNotifee) ListenClose(_ network.Network, _ multiaddr.Multiaddr) {
	n.ds.logger.Debugf("[libp2p/notifee] listen close...")
}
