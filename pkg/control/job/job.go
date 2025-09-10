package job

import (
	"time"
)

type Job struct {
	Name     string        // 执行任务名称
	Interval time.Duration // 执行间隔
	Task     func()
}
