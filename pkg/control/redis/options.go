package redis

import (
	"strconv"
	"strings"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/redis/go-redis/v9"
)

// Option 定义用于配置 Control 的函数类型
// @description 用于通过函数式选项模式配置 Control 结构体
// @param control *Control 要配置的 Control 实例
type Option func(control *Control)

// WithTracer 设置 tracer 控制器
// @description 设置 Redis 服务器的追踪控制器
// @param tracerControl *tracer.Control 追踪控制器实例
// @return Option 配置函数
func WithTracer(tracerControl *tracer.Control) Option {
	return func(control *Control) {
		control.tracerControl = tracerControl
	}
}

// WithHost 设置Redis主机地址
// @description 设置Redis服务器的主机地址
// @param host string Redis主机地址
// @return func(*redis.Options) Redis选项配置函数
func WithHost(host string) func(*redis.Options) {
	return func(o *redis.Options) {
		// 保存端口信息
		parts := strings.Split(o.Addr, ":")
		port := "6379" // 默认端口
		if len(parts) > 1 {
			port = parts[1]
		}
		o.Addr = host + ":" + port
	}
}

// WithPort 设置Redis端口
// @description 设置Redis服务器的端口号
// @param port int Redis端口号
// @return func(*redis.Options) Redis选项配置函数
func WithPort(port int) func(*redis.Options) {
	return func(o *redis.Options) {
		// 从Addr中提取主机并重新组合
		parts := strings.Split(o.Addr, ":")
		host := "localhost" // 默认主机
		if len(parts) > 0 && parts[0] != "" {
			host = parts[0]
		}
		o.Addr = host + ":" + strconv.Itoa(port)
	}
}

// WithPassword 设置Redis密码
// @description 设置Redis服务器的认证密码
// @param password string Redis认证密码
// @return func(*redis.Options) Redis选项配置函数
func WithPassword(password string) func(*redis.Options) {
	return func(o *redis.Options) {
		o.Password = password
	}
}

// WithDB 设置Redis数据库编号
// @description 设置要使用的Redis数据库编号
// @param db int Redis数据库编号
// @return func(*redis.Options) Redis选项配置函数
func WithDB(db int) func(*redis.Options) {
	return func(o *redis.Options) {
		o.DB = db
	}
}

// WithPoolSize 设置连接池大小
// @description 设置Redis连接池的最大连接数
// @param poolSize int 连接池大小
// @return func(*redis.Options) Redis选项配置函数
func WithPoolSize(poolSize int) func(*redis.Options) {
	return func(o *redis.Options) {
		o.PoolSize = poolSize
	}
}

// WithMaxRetries 设置最大重试次数
// @description 设置Redis操作失败时的最大重试次数
// @param maxRetries int 最大重试次数
// @return func(*redis.Options) Redis选项配置函数
func WithMaxRetries(maxRetries int) func(*redis.Options) {
	return func(o *redis.Options) {
		o.MaxRetries = maxRetries
	}
}

// WithMinRetryBackoff 设置最小重试退避时间
// @description 设置Redis重试操作的最小退避时间
// @param minRetryBackoff time.Duration 最小重试退避时间
// @return func(*redis.Options) Redis选项配置函数
func WithMinRetryBackoff(minRetryBackoff time.Duration) func(*redis.Options) {
	return func(o *redis.Options) {
		o.MinRetryBackoff = minRetryBackoff
	}
}

// WithMaxRetryBackoff 设置最大重试退避时间
// @description 设置Redis重试操作的最大退避时间
// @param maxRetryBackoff time.Duration 最大重试退避时间
// @return func(*redis.Options) Redis选项配置函数
func WithMaxRetryBackoff(maxRetryBackoff time.Duration) func(*redis.Options) {
	return func(o *redis.Options) {
		o.MaxRetryBackoff = maxRetryBackoff
	}
}

// WithDialTimeout 设置拨号超时时间
// @description 设置Redis连接拨号的超时时间
// @param dialTimeout time.Duration 拨号超时时间
// @return func(*redis.Options) Redis选项配置函数
func WithDialTimeout(dialTimeout time.Duration) func(*redis.Options) {
	return func(o *redis.Options) {
		o.DialTimeout = dialTimeout
	}
}

// WithReadTimeout 设置读取超时时间
// @description 设置Redis读取操作的超时时间
// @param readTimeout time.Duration 读取超时时间
// @return func(*redis.Options) Redis选项配置函数
func WithReadTimeout(readTimeout time.Duration) func(*redis.Options) {
	return func(o *redis.Options) {
		o.ReadTimeout = readTimeout
	}
}

// WithWriteTimeout 设置写入超时时间
// @description 设置Redis写入操作的超时时间
// @param writeTimeout time.Duration 写入超时时间
// @return func(*redis.Options) Redis选项配置函数
func WithWriteTimeout(writeTimeout time.Duration) func(*redis.Options) {
	return func(o *redis.Options) {
		o.WriteTimeout = writeTimeout
	}
}
