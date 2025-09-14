package job

import (
	"context"
	"sync"
	"time"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/ants"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	logger *zap.SugaredLogger
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc

	jobs []*Job

	jobMutex sync.RWMutex

	wg sync.WaitGroup

	// antsPoolControl ants线程池控制器，用于在单独的goroutine中执行任务
	antsPoolControl *ants.Control
}

// NewJobControl 创建作业控制器
// @param loggerControl *logger.Control 日志控制器
// @param antsPoolControl *ants.Control 线程池控制器
// @return *Control 作业控制器实例
func NewJobControl(loggerControl *logger.Control, antsPoolControl *ants.Control) *Control {
	jobLogger := loggerControl.GenLogger(logger.ModuleJob)
	jobLogger.Infof("[control] starting create job control...")
	ctx, cancel := context.WithCancel(context.Background())

	control := &Control{
		logger:          jobLogger,
		ctx:             ctx,
		cancel:          cancel,
		jobs:            make([]*Job, 0, 8),
		antsPoolControl: antsPoolControl,
	}
	return control
}

// StartAllRegisterJobs 启动所有注册的作业
func (c *Control) StartAllRegisterJobs() {
	c.logger.Debugf("[control] starting all register jobs...")
	c.jobMutex.Lock()
	defer c.jobMutex.Unlock()
	for _, job := range c.jobs {
		c.logger.Debugf("[control] starting job name: %s", job.Name)
		c.wg.Add(1)
		go c.runJob(job)
	}
}

// StopAllRegisterJobs 停止所有注册的作业
func (c *Control) StopAllRegisterJobs() {
	c.logger.Debugf("[control] stopping all register jobs...")
	c.cancel()
	c.wg.Wait()
	c.logger.Infof("[control] all register jobs stopped...")
}

// RegisterJob 注册一个新的作业
// @param j *Job 要注册的作业对象
func (c *Control) RegisterJob(j *Job) {
	if c.antsPoolControl == nil {
		c.logger.Warnf("[control] no runner ants pool, skip register job...")
		return
	}

	c.logger.Debugf("[control] register job name: %s", j.Name)

	c.jobMutex.Lock()
	defer c.jobMutex.Unlock()

	c.logger.Debugf("[control] register job to jobs...")
	c.jobs = append(c.jobs, j)

	c.wg.Add(1)
	go c.runJob(j)

	c.logger.Infof("[control] register job name: %s successfully...", j.Name)
}

// runJob 运行指定的作业
// @param job *Job 要运行的作业对象
func (c *Control) runJob(job *Job) {
	defer c.wg.Done()
	ticker := time.NewTicker(job.Interval)
	c.logger.Infof("[control] starting job name: %s", job.Name)
	for {
		select {
		case <-c.ctx.Done():
			c.logger.Infof("[control] job name: %s closed", job.Name)
			return
		case <-ticker.C:
			c.logger.Debugf("[control] job name: %s running...", job.Name)
			// job.Task()
			if err := c.antsPoolControl.Submit(job.Task); err != nil {
				c.logger.Warnf("[control] submit job to ants pool failed: %v", err)
			}
		}
	}
}

// StartUp 启动作业服务器
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Infof("[control] starting job server...")
		c.StartAllRegisterJobs()
	})
}

// Shutdown 关闭作业服务器
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	c.logger.Infof("[control] shutting down job server...")
	c.StopAllRegisterJobs()
	return nil
}
