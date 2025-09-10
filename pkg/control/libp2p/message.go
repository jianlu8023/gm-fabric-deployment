package libp2p

import (
	"fmt"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// Message 定义消息结构
type Message struct {
	Type    string  `json:"msg_type" yaml:"msg_type"`
	Content []byte  `json:"msg_content" yaml:"msg_content"`
	From    peer.ID `json:"msg_from" yaml:"msg_from"`
	To      peer.ID `json:"msg_to" yaml:"msg_to"`
}

// MarshalJSON 自定义JSON序列化方法
func (m *Message) MarshalJSON() ([]byte, error) {
	// 创建一个中间结构体用于JSON序列化
	type Alias Message
	return json.Marshal(&struct {
		*Alias
		From string `json:"msg_from"`
		To   string `json:"msg_to"`
	}{
		Alias: (*Alias)(m),
		From:  m.From.String(), // 将peer.ID转换为字符串
		To:    m.To.String(),   // 将peer.ID转换为字符串
	})
}

// UnmarshalJSON 自定义JSON反序列化方法
func (m *Message) UnmarshalJSON(data []byte) error {
	// 创建一个中间结构体用于JSON反序列化
	type Alias Message
	aux := &struct {
		*Alias
		From string `json:"msg_from"`
		To   string `json:"msg_to"`
	}{
		Alias: (*Alias)(m),
	}

	// 先反序列化到中间结构体
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// 然后将字符串转换回peer.ID
	var err error
	if aux.From != "" {
		m.From, err = peer.Decode(aux.From)
		if err != nil {
			return fmt.Errorf("invalid from peer ID: %w", err)
		}
	}

	if aux.To != "" {
		m.To, err = peer.Decode(aux.To)
		if err != nil {
			return fmt.Errorf("invalid to peer ID: %w", err)
		}
	}

	return nil
}

// MessageHandler 消息处理函数类型
type MessageHandler func(protocolID protocol.ID, msg *Message)

// MessageWithPeer 包含消息和目标节点信息的结构体
type MessageWithPeer struct {
	PeerID     peer.ID
	ProtocolID protocol.ID
	Msg        *Message
}

const (
	BasePing                 = "base/ping"
	BasePong                 = "base/pong"
	BaseShutdown             = "base/shutdown"
	DockerNetworks           = "docker/networks"
	CollectionNode           = "collection/node"
	CollectionDockerNetworks = "collection/docker/networks"
	CollectionDockerImages   = "collection/docker/images"
	DockerImages             = "docker/images"
	Libp2pNode               = "libp2p/node"
)
