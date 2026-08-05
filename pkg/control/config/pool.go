package config

import (
	"time"
)

// GetPoolSize 获取线程池大小
func (c *AntsPoolConfig) GetPoolSize() int {
	return c.PoolSize
}

// GetMaxPoolSize 获取最大线程池大小
func (c *AntsPoolConfig) GetMaxPoolSize() int {
	return c.MaxPoolSize
}

// GetExpiryDuration 获取工作协程过期时间
func (c *AntsPoolConfig) GetExpiryDuration() time.Duration {
	return time.Duration(c.ExpiryDuration) * time.Second
}

// IsPreAlloc 是否预分配工作协程
func (c *AntsPoolConfig) IsPreAlloc() bool {
	return c.PreAlloc
}

// IsNonblocking 是否非阻塞模式
func (c *AntsPoolConfig) IsNonblocking() bool {
	return c.Nonblocking
}
