package config

import (
	"time"
)

// GetPoolSize 获取线程池大小
func (a *AntsPoolConfig) GetPoolSize() int {
	return a.PoolSize
}

// GetMaxPoolSize 获取最大线程池大小
func (a *AntsPoolConfig) GetMaxPoolSize() int {
	return a.MaxPoolSize
}

// GetExpiryDuration 获取工作协程过期时间
func (a *AntsPoolConfig) GetExpiryDuration() time.Duration {
	return time.Duration(a.ExpiryDuration) * time.Second
}

// IsPreAlloc 是否预分配工作协程
func (a *AntsPoolConfig) IsPreAlloc() bool {
	return a.PreAlloc
}

// IsNonblocking 是否非阻塞模式
func (a *AntsPoolConfig) IsNonblocking() bool {
	return a.Nonblocking
}
