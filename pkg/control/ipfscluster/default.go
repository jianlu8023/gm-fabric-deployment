package ipfscluster

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

type StrategyType string

const (
	FailOver   StrategyType = "failover"
	RoundRobin StrategyType = "roundrobin"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
)

func getDefaultConfig() *config.IpfsClusterConfig {
	return &config.IpfsClusterConfig{
		Enabled: false,
	}
}
