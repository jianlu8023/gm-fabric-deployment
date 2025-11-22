package redis

import "errors"

var (
	// ErrNoAliveRedisClient 没有存活的Redis客户端
	// @description 当没有可用的Redis客户端时返回此错误
	ErrNoAliveRedisClient = errors.New("no alive redis client")

	// ErrRedisNotEnabled Redis未启用
	// @description 当Redis配置未启用时返回此错误
	ErrRedisNotEnabled = errors.New("redis is not enabled")

	// ErrRedisConnectionFailed Redis连接失败
	// @description 当无法连接到Redis服务器时返回此错误
	ErrRedisConnectionFailed = errors.New("failed to connect to redis")
)
