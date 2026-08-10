package libp2p

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/jianlu8023/go-tools/v2/pkg/json"

	"github.com/libp2p/go-libp2p/core/peer"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(protocolID protocol.ID, msg *Message)

// MessageWithPeer 包含消息和目标节点信息的结构体
type MessageWithPeer struct {
	PeerID     peer.ID
	ProtocolID protocol.ID
	Msg        *Message
}

const (
	MsgBasePing                 = "base/ping"
	MsgBasePong                 = "base/pong"
	MsgBaseShutdown             = "base/shutdown"
	MsgDockerNetworks           = "docker/networks"
	MsgCollectionNode           = "collection/node"
	MsgCollectionDockerNetworks = "collection/docker/networks"
	MsgCollectionDockerImages   = "collection/docker/images"
	MsgDockerImages             = "docker/images"
	MsgLibp2pNode               = "libp2p/node"
	MsgDockerImagePull          = "docker/pull"
)

// Message 定义消息结构
//
// @description P2P消息信封，Content字段当前为明文。
// TODO(app-layer-encrypt): 接入国密应用层加密后，Content应为密文，收发时需在MarshalJSON/UnmarshalJSON中加解密
type Message struct {
	Type    string  `json:"msg_type" yaml:"msg_type"`
	Content []byte  `json:"msg_content" yaml:"msg_content"`
	From    peer.ID `json:"msg_from" yaml:"msg_from"`
	To      peer.ID `json:"msg_to" yaml:"msg_to"`
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 将peer.ID转换为字符串进行JSON序列化
// TODO(app-layer-encrypt): 在序列化前对Content做国密加密
func (m *Message) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
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
//
// @description 将JSON中的字符串转换回peer.ID
// TODO(app-layer-decrypt): 在反序列化后对Content做国密解密
func (m *Message) UnmarshalJSON(data []byte) error {
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

// DockerImagePullContent Docker镜像拉取请求的消息内容
//
// @description 作为Message.Content的payload，用于节点间传递镜像拉取请求参数
type DockerImagePullContent struct {
	ImageName    string `json:"image_name,omitempty" yaml:"image_name,omitempty"`
	Platform     string `json:"platform,omitempty" yaml:"platform,omitempty"`
	RegistryAuth string `json:"registry_auth,omitempty" yaml:"registry_auth,omitempty"`
}

func (d DockerImagePullContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(d)
}

func (d DockerImagePullContent) String() string {
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Sprintf("DockerImagePullContent{ImageName: %s, Platform: %s}", d.ImageName, d.Platform)
	}
	return string(data)
}
