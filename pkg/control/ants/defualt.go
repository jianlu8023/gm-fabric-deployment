package ants

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.AntsPoolConfig {
	return &config.AntsPoolConfig{
		Enabled:        false,
		PoolSize:       8,
		MaxPoolSize:    16,
		ExpiryDuration: 60,
		PreAlloc:       true,
		Nonblocking:    true,
	}
}
