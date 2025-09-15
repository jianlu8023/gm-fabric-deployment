package ants

import (
	"context"
	"sync"
	"time"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
)

// Control 线程池控制器
// @description 基于 ants/v2 实现的线程池控制器，提供任务提交、监控和管理功能
// @struct
// @property pool *ants.Pool 线程池实例
// @property config *config.AntsPoolConfig 线程池配置
// @property logger *zap.SugaredLogger 日志记录器
// @property ctx context.Context 上下文
// @property once sync.Once 确保某些操作只执行一次
// @property mu sync.RWMutex 读写锁，保护线程池状态
// @property isRunning bool 线程池是否运行中
// @property taskCount int64 当前已提交的任务数量
// @property completedCount int64 当前已完成的任务数量
// @property failedCount int64 当前已失败的任务数量
// @property waitGroup sync.WaitGroup 等待所有任务完成的同步原语
// @since v1.0.0
// @example
// 	config := &config.AntsPoolConfig{
// 		PoolSize:      10,
// 		MaxPoolSize:   20,
// 		ExpiryDuration: 10 * time.Second,
// 		PreAlloc:      true,
// 		Nonblocking:   false,
// 	}
// 	loggerControl := logger.NewLoggerControl()
// 	control := NewAntsPoolControl(config, loggerControl)
// 	control.StartUp()
// 	defer control.Shutdown()
// 	control.Submit(func() {
// 		// 任务逻辑
// 	})

// PoolConfig 线程池配置接口
// @description 定义线程池控制器需要的配置项
// @interface
// @method GetPoolSize() int 获取线程池大小
// @method GetMaxPoolSize() int 获取最大线程池大小
// @method GetExpiryDuration() time.Duration 获取工作协程过期时间
// @method IsPreAlloc() bool 是否预分配工作协程
// @method IsNonblocking() bool 是否非阻塞模式
// @since v1.0.0
// type PoolConfig interface {
// 	GetPoolSize() int
// 	GetMaxPoolSize() int
// 	GetExpiryDuration() time.Duration
// 	IsPreAlloc() bool
// 	IsNonblocking() bool
// }

// Control 线程池控制器
// @description 线程池控制器的具体实现
// @struct
// @property pool *ants.Pool 线程池实例
// @property config PoolConfig 线程池配置
// @property logger *zap.SugaredLogger 日志记录器
// @property ctx context.Context 上下文
// @property ctxCancel context.CancelFunc 上下文取消函数
// @property once sync.Once 确保某些操作只执行一次
// @property mu sync.RWMutex 读写锁，保护线程池状态
// @property isRunning bool 线程池是否运行中
// @property taskCount int64 当前已提交的任务数量
// @property completedCount int64 当前已完成的任务数量
// @property failedCount int64 当前已失败的任务数量
// @property waitGroup sync.WaitGroup 等待所有任务完成的同步原语
// @since v1.0.0
type Control struct {
	pool           *ants.Pool
	config         *config.AntsPoolConfig
	logger         *zap.SugaredLogger
	ctx            context.Context
	ctxCancel      context.CancelFunc
	once           sync.Once
	mutex          sync.RWMutex
	isRunning      bool
	taskCount      int64
	completedCount int64
	failedCount    int64
	waitGroup      sync.WaitGroup
}

// NewAntsPoolControl 创建线程池控制器
// @param config *config.AntsPoolConfig 线程池配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control 线程池控制器实例
// @since v1.0.0
func NewAntsPoolControl(config *config.AntsPoolConfig, loggerControl *logger.Control) *Control {
	antsLogger := loggerControl.GenLogger(logger.ModuleAnts)
	antsLogger.Infof("[control] start new ants pool control...")

	ctx, cancel := context.WithCancel(context.Background())

	control := &Control{
		config:    config,
		logger:    antsLogger,
		ctx:       ctx,
		ctxCancel: cancel,
	}

	antsLogger.Warnf("[control] ants pool not initialized, please call control.StartUp()")

	return control
}

// StartUp 启动线程池
// @param failedFunc func(err error) 启动失败时的回调函数
// @since v1.0.0
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting up ants pool...")

		// 配置线程池选项
		options := []ants.Option{
			ants.WithPreAlloc(c.config.IsPreAlloc()),
			ants.WithNonblocking(c.config.IsNonblocking()),
			ants.WithExpiryDuration(c.config.GetExpiryDuration()),
			ants.WithLogger(newAntsLogger(c.logger)),
		}

		// 创建线程池
		pool, err := ants.NewPool(c.config.GetPoolSize(), options...)
		if err != nil {
			c.logger.Errorf("[control] failed to create ants pool: %v", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}

		c.pool = pool
		c.isRunning = true

		c.logger.Infof("[control] ants pool started successfully, pool size: %d, max pool size: %d",
			c.config.GetPoolSize(), c.config.GetMaxPoolSize())

		// 启动监控协程
		// go c.monitor()
		if err = c.Submit(c.monitor); err != nil {
			c.logger.Errorf("[control] submit ants pool monitor task failed: %v", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
	})
}

// Shutdown 关闭线程池
// @return error 关闭过程中可能产生的错误
// @since v1.0.0
func (c *Control) Shutdown() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.isRunning || c.pool == nil {
		return nil
	}

	c.logger.Infof("[control] shutdown ants pool...")

	// 取消上下文
	c.ctxCancel()

	// 关闭线程池
	c.pool.Release()

	// 等待所有任务完成
	c.waitGroup.Wait()

	c.isRunning = false

	c.logger.Infof("[control] ants pool shutdown completely. Total tasks: %d, Completed: %d, Failed: %d",
		c.taskCount, c.completedCount, c.failedCount)

	return nil
}

// Submit 提交任务到线程池
// @param task func() 要执行的任务函数
// @return error 提交任务时可能产生的错误
// @since v1.0.0
func (c *Control) Submit(task func()) error {
	c.mutex.RLock()
	if !c.isRunning || c.pool == nil {
		c.mutex.RUnlock()
		return ants.ErrPoolClosed
	}
	c.mutex.RUnlock()

	// 增加任务计数
	c.taskCount++
	c.waitGroup.Add(1)

	// 包装任务，增加完成和错误计数
	wrappedTask := func() {
		defer c.waitGroup.Done()
		defer func() {
			if r := recover(); r != nil {
				c.failedCount++
				c.logger.Errorf("[control] task panic: %v", r)
			} else {
				c.completedCount++
			}
		}()

		task()
	}

	// 提交任务到线程池
	return c.pool.Submit(wrappedTask)
}

// SubmitWithTimeout 提交任务到线程池，并设置超时时间
// @param task func() 要执行的任务函数
// @param timeout time.Duration 超时时间
// @return error 提交任务时可能产生的错误
// @since v1.0.0
func (c *Control) SubmitWithTimeout(task func(), timeout time.Duration) error {
	// 创建一个带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()

	// 创建一个通道用于接收任务完成的信号
	done := make(chan struct{})

	// 提交任务到线程池
	err := c.Submit(func() {
		defer close(done)
		task()
	})

	if err != nil {
		return err
	}

	// 等待任务完成或超时
	select {
	case <-done:
		return nil
	case <-timeoutCtx.Done():
		return timeoutCtx.Err()
	}
}

// GetPoolStats 获取线程池统计信息
// @return map[string]interface{} 线程池统计信息
// @since v1.0.0
func (c *Control) GetPoolStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["isRunning"] = c.isRunning
	stats["taskCount"] = c.taskCount
	stats["completedCount"] = c.completedCount
	stats["failedCount"] = c.failedCount

	if c.isRunning && c.pool != nil {
		stats["runningWorkers"] = c.pool.Running()
		stats["waitingWorkers"] = c.pool.Waiting()
		stats["capTasks"] = c.pool.Cap()
		stats["freeWorkers"] = c.pool.Free()
	}

	return stats
}

// monitor 监控线程池状态
// @private
// @since v1.0.0
func (c *Control) monitor() {
	// 	每分钟打印一次线程池状态
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mutex.RLock()
			if c.isRunning && c.pool != nil {
				c.logger.Debugf("[control] ants pool stats - Running: %d, Waiting: %d, Cap: %d, Free: %d, Tasks: %d, Completed: %d, Failed: %d",
					c.pool.Running(), c.pool.Waiting(), c.pool.Cap(), c.pool.Free(),
					c.taskCount, c.completedCount, c.failedCount)
			}
			c.mutex.RUnlock()

		case <-c.ctx.Done():
			c.logger.Debugf("[control] ants pool monitor stopped")
			return
		}
	}
}

// Resize 调整线程池大小
// @param size int 新的线程池大小
// @return error 调整大小时可能产生的错误
// @since v1.0.0
func (c *Control) Resize(size int) error {
	c.mutex.RLock()
	if !c.isRunning || c.pool == nil {
		c.mutex.RUnlock()
		return ants.ErrPoolClosed
	}
	c.mutex.RUnlock()

	if size <= 0 {
		return ants.ErrInvalidPoolIndex
	}

	c.logger.Infof("[control] resizing ants pool from %d to %d", c.config.GetPoolSize(), size)

	// 调整线程池大小
	c.pool.Tune(size)

	// 如果配置支持，更新配置中的池大小
	c.config.PoolSize = size

	c.logger.Infof("[control] ants pool resized successfully to %d", size)

	return nil
}

// GetTaskCount 获取已提交的任务总数
// @return int64 任务总数
// @since v1.0.0
func (c *Control) GetTaskCount() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.taskCount
}

// GetCompletedCount 获取已完成的任务数量
// @return int64 已完成的任务数量
// @since v1.0.0
func (c *Control) GetCompletedCount() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.completedCount
}

// GetFailedCount 获取已失败的任务数量
// @return int64 已失败的任务数量
// @since v1.0.0
func (c *Control) GetFailedCount() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.failedCount
}

// IsRunning 检查线程池是否运行中
// @return bool 线程池运行状态
// @since v1.0.0
func (c *Control) IsRunning() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isRunning
}
