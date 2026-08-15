package grpc

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.GrpcConfig {
	return &config.GrpcConfig{
		Enabled:             false,
		NodeID:              "default-node",
		HealthCheckInterval: 60,
	}
}
