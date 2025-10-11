package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jinzhu/copier"
)

type Node struct {
	AutoUid              int       `json:"auto_uid" yaml:"auto_uid"`                               // 自增ID
	NodeId               string    `json:"node_id" yaml:"node_id"`                                 // 节点ID（唯一标识）
	NodeIp               string    `json:"node_ip" yaml:"node_ip"`                                 // 节点ip
	LastAliveMessageTime time.Time `json:"last_alive_message_time" yaml:"last_alive_message_time"` // 最后一次收到心跳时间
	IsAlive              bool      `json:"is_alive" yaml:"is_alive"`                               // 是否存活状态
	IsMySelf             bool      `json:"is_my_self" yaml:"is_my_self"`                           // 是否是本机节点
}

func (n Node) String() string {
	str, _ := json.MarshalString(n)
	return str
}

func main() {

	src := &model.Libp2pNode{
		NodeIp:               "127.0.0.1",
		NodeId:               "001",
		LastAliveMessageTime: time.Now(),
		IsAlive:              sql.NullBool{Valid: true, Bool: false},
		IsMySelf:             sql.NullBool{Valid: true, Bool: true},
	}

	var dst Node

	err := copier.CopyWithOption(&dst, src, copier.Option{
		DeepCopy:    true,
		IgnoreEmpty: true,
	})
	if err != nil {
		log.Printf("copy failed: %v\n", err)
		return
	}

	fmt.Printf("src: %v\n", src)
	fmt.Printf("dst: %v\n", dst)

}
