package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Control Redis控制结构体
type Control struct {
	client        *redis.Client
	config        *config.RedisConfig
	logger        *zap.SugaredLogger
	ctx           context.Context
	cancel        context.CancelFunc
	once          sync.Once
	tracerControl *tracer.Control
}

// NewRedisControl 创建一个新的Redis控制器
// @param redisConfig *config.RedisConfig redis配置
// @param loggerControl *logger.Control 日志控制器
// @param opts ...Option 可选的配置选项
// @return *Control redis控制器
func NewRedisControl(redisConfig *config.RedisConfig, loggerControl *logger.Control, opts ...Option) *Control {
	if redisConfig == nil {
		redisConfig = getDefaultConfig()
	}

	// 如果Redis未启用，返回nil
	if !redisConfig.Enabled {
		return nil
	}

	redisLogger := loggerControl.GenLogger(logger.ModuleRedis)
	redisLogger.Infof("[control] starting new redis control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化RedisControl
	rc := &Control{
		logger: redisLogger,
		ctx:    ctx,
		cancel: cancel,
		config: redisConfig,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(rc)
	}

	redisLogger.Infof("[control] redis control started successfully")
	return rc
}

// initClient 初始化Redis客户端
// @description 初始化Redis客户端连接
// @return error 初始化过程中的错误
func (rc *Control) initClient() error {
	rc.logger.Debugf("[control] initializing redis client...")

	// 创建Redis客户端
	rc.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", rc.config.Host, rc.config.Port),
		Password: rc.config.Password,
		DB:       rc.config.DB,
		PoolSize: rc.config.PoolSize,
	})

	// 如果启用了tracer，则添加追踪支持
	if rc.tracerControl != nil {
		if err := redisotel.InstrumentTracing(rc.client,
			redisotel.WithTracerProvider(rc.tracerControl.TracerProvider()),
			redisotel.WithCallerEnabled(true),
			redisotel.WithDBStatement(true),
		); err != nil {
			rc.logger.Errorf("[control] failed to instrument tracing for redis client: %v", err)
			return err
		}
		if err := redisotel.InstrumentMetrics(rc.client,
			redisotel.WithMeterProvider(rc.tracerControl.MeterProvider()),
		); err != nil {
			rc.logger.Errorf("[control] failed to instrument metrics for redis client: %v", err)
			return err
		}
	}

	return nil
}

// StartUp 启动Redis服务
// @description 启动Redis服务，测试连接是否正常
// @param failedFunc func(err error) 启动失败回调函数
func (rc *Control) StartUp(failedFunc func(err error)) {
	rc.once.Do(func() {
		if rc.config.Enabled {
			// 初始化Redis客户端
			if err := rc.initClient(); err != nil {
				rc.logger.Errorf("[control] redis initialization client failed: %v", err)
				rc.cancel()
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}

			rc.logger.Debugf("[control] redis service is already running...")
			// 测试连接
			_, err := rc.client.Ping(rc.ctx).Result()
			if err != nil {
				rc.logger.Errorf("[control] failed to connect to redis: %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}

			rc.logger.Infof("[control] connected to redis successfully")
		}
	})
}

// Shutdown 关闭Redis客户端连接
// @description 关闭Redis客户端连接并释放资源
// @return error 关闭过程中的错误
func (rc *Control) Shutdown() error {
	rc.logger.Infof("[control] shutting down redis control...")
	rc.cancel()
	if rc.client != nil {
		if err := rc.client.Close(); err != nil {
			rc.logger.Errorf("[control] failed to close redis client: %v", err)
			return err
		}
	}
	rc.logger.Infof("[control] redis control shutdown successfully")
	return nil
}

// GetClient 获取Redis客户端
// @description 获取Redis客户端实例
// @return *redis.Client Redis客户端实例
func (rc *Control) GetClient() *redis.Client {
	return rc.client
}

// Set 设置键值对
// @description 设置键值对
// @param key string 键
// @param value interface{} 值
// @param expiration time.Duration 过期时间
// @return error 设置过程中的错误
func (rc *Control) Set(key string, value interface{}, expiration time.Duration) error {
	if rc.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	err := rc.client.Set(rc.ctx, key, value, expiration).Err()
	if err != nil {
		rc.logger.Errorf("[control] failed to set key %s: %v", key, err)
		return err
	}

	rc.logger.Debugf("[control] set key %s successfully", key)
	return nil
}

// Get 获取键对应的值
// @description 获取键对应的值
// @param key string 键
// @return string 键对应的值
// @return error 获取过程中的错误
func (rc *Control) Get(key string) (string, error) {
	if rc.client == nil {
		return "", fmt.Errorf("redis client is nil")
	}

	value, err := rc.client.Get(rc.ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			rc.logger.Debugf("[control] key %s does not exist", key)
		} else {
			rc.logger.Errorf("[control] failed to get key %s: %v", key, err)
		}
		return "", err
	}

	rc.logger.Debugf("[control] get key %s successfully", key)
	return value, nil
}

// Delete 删除键
// @description 删除指定的键
// @param keys ...string 要删除的键
// @return int64 删除的键数量
// @return error 删除过程中的错误
func (rc *Control) Delete(keys ...string) (int64, error) {
	if rc.client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	count, err := rc.client.Del(rc.ctx, keys...).Result()
	if err != nil {
		rc.logger.Errorf("[control] failed to delete keys %v: %v", keys, err)
		return 0, err
	}

	rc.logger.Debugf("[control] deleted %d keys successfully", count)
	return count, nil
}

// Exists 检查键是否存在
// @description 检查指定的键是否存在
// @param keys ...string 要检查的键
// @return int64 存在的键数量
// @return error 检查过程中的错误
func (rc *Control) Exists(keys ...string) (int64, error) {
	if rc.client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	count, err := rc.client.Exists(rc.ctx, keys...).Result()
	if err != nil {
		rc.logger.Errorf("[control] failed to check existence of keys %v: %v", keys, err)
		return 0, err
	}

	rc.logger.Debugf("[control] checked existence of keys, %d keys exist", count)
	return count, nil
}

// Expire 设置键的过期时间
// @description 设置键的过期时间
// @param key string 键
// @param expiration time.Duration 过期时间
// @return bool 是否设置成功
// @return error 设置过程中的错误
func (rc *Control) Expire(key string, expiration time.Duration) (bool, error) {
	if rc.client == nil {
		return false, fmt.Errorf("redis client is nil")
	}

	result, err := rc.client.Expire(rc.ctx, key, expiration).Result()
	if err != nil {
		rc.logger.Errorf("[control] failed to set expiration for key %s: %v", key, err)
		return false, err
	}

	rc.logger.Debugf("[control] set expiration for key %s successfully", key)
	return result, nil
}

// TTL 获取键的剩余过期时间
// @description 获取键的剩余过期时间
// @param key string 键
// @return time.Duration 剩余过期时间
// @return error 获取过程中的错误
func (rc *Control) TTL(key string) (time.Duration, error) {
	if rc.client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	duration, err := rc.client.TTL(rc.ctx, key).Result()
	if err != nil {
		rc.logger.Errorf("[control] failed to get TTL for key %s: %v", key, err)
		return 0, err
	}

	rc.logger.Debugf("[control] got TTL for key %s: %v", key, duration)
	return duration, nil
}
