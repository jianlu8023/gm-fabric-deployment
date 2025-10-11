package ipfscluster

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// strategyType 负载均衡策略类型
// @description 定义IPFS集群客户端的负载均衡策略类型
type strategyType string

const (
	// failOver 故障转移策略
	// @description 当前节点故障时自动切换到下一个可用节点
	failOver strategyType = "failover"

	// roundRobin 轮询策略
	// @description 按顺序轮询使用各个节点
	roundRobin strategyType = "roundrobin"
)

const (
	// timeFormat 时间格式化字符串
	// @description 用于文件上传时间的格式化
	timeFormat = "2006-01-02 15:04:05.000"
)

// getDefaultConfig 获取默认的IPFS集群配置
// @description 返回一个默认的IPFS集群配置实例，Enabled字段默认为false
// @return *config.IpfsClusterConfig 默认的IPFS集群配置
func getDefaultConfig() *config.IpfsClusterConfig {
	return &config.IpfsClusterConfig{
		Enabled: false,
	}
}
