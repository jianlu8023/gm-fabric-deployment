package job

import (
	"context"
	"sync"
	"time"

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
}

func NewJobControl(loggerControl *logger.Control) *Control {
	jobLogger := loggerControl.GenLogger(logger.ModuleJob)
	jobLogger.Infof("[control] starting create job control...")
	ctx, cancel := context.WithCancel(context.Background())

	control := &Control{
		logger: jobLogger,
		ctx:    ctx,
		cancel: cancel,
		jobs:   make([]*Job, 0, 8),
	}
	return control
}

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

func (c *Control) StopAllRegisterJobs() {
	c.logger.Debugf("[control] stopping all register jobs...")
	c.cancel()
	c.wg.Wait()
	c.logger.Infof("[control] all register jobs stopped...")
}

func (c *Control) RegisterJob(j *Job) {
	c.logger.Debugf("[control] register job name: %s", j.Name)

	c.jobMutex.Lock()
	defer c.jobMutex.Unlock()

	c.logger.Debugf("[control] register job to jobs...")
	c.jobs = append(c.jobs, j)

	c.wg.Add(1)
	go c.runJob(j)

	c.logger.Infof("[control] register job name: %s successfully...", j.Name)
}

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
			job.Task()
		}
	}
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Infof("[control] starting job server...")
		c.StartAllRegisterJobs()
	})
}

func (c *Control) Shutdown() error {
	c.logger.Infof("[control] shutting down job server...")
	c.StopAllRegisterJobs()
	return nil
}
