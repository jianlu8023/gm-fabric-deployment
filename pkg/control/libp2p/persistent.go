package libp2p

import "time"

// PersistentPeerInfo 持久化节点信息
//
// @description 用于将节点信息序列化到文件，启动时加载并尝试重连
// @struct
type PersistentPeerInfo struct {
	PeerID        string    `json:"peer_id"`                  // 节点ID
	Addrs         []string  `json:"addresses"`                // 节点地址列表
	IsBootstrap   bool      `json:"is_bootstrap"`             // 是否是bootstrap节点
	LastConnected time.Time `json:"last_connected,omitempty"` // 最后连接时间
}

const (
	persistenceEnabled  = false
	persistenceFilePath = "data/peers.json"
	saveInterval        = 300 * time.Second
)

type PersistenceConfig struct {
	Enabled      bool          `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                   // 是否启用持久化
	FilePath     string        `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`             // 持久化文件路径
	SaveInterval time.Duration `json:"save_interval,omitempty" yaml:"save_interval,omitempty" mapstructure:"save_interval"` // 定期保存间隔（秒），默认300
}
