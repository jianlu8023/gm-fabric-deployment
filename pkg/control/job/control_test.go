package job

import (
	"fmt"
	"testing"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/ants"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/stretchr/testify/assert"
)

func genAntsControl(t *testing.T) *ants.Control {
	antsControl := ants.NewAntsPoolControl(&config.AntsPoolConfig{
		Enabled:     true,
		PoolSize:    8,
		MaxPoolSize: 16,

		ExpiryDuration: 60,
		PreAlloc:       true,
		Nonblocking:    true,
	}, logger.NewLoggerControl(&config.LoggerConfig{
		DefaultLogLevel: "debug",
		PrintFormat:     "console",
		FilePath:        "logs/ants.log",
		MaxAge:          7,
		RotationTime:    3,
	}))
	assert.NotNil(t, antsControl, "antsControl must not be nil")
	return antsControl
}

func TestNewJobControl(t *testing.T) {
	tests := []struct {
		name          string
		loggerControl *logger.Control
	}{
		{
			name:          "WithoutLoggerControl",
			loggerControl: nil,
		},
		{
			name: "withLoggerControl",
			loggerControl: logger.NewLoggerControl(&config.LoggerConfig{
				DefaultLogLevel: "debug",
				PrintFormat:     "console",
				FilePath:        "logs/job.log",
				MaxAge:          7,
				RotationTime:    3,
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			antsControl := genAntsControl(t)
			defer func() {
				_ = antsControl.Shutdown()
			}()
			antsControl.StartUp(func(err error) {
				if err != nil {
					t.Fatal(err)
				}
			})
			control := NewJobControl(tt.loggerControl, antsControl)
			assert.NotNil(t, control, "control must not be nil")
			defer func() {
				_ = control.Shutdown()
			}()
			control.StartUp(func(err error) {
				if err != nil {
					t.Fatal(err)
				}

			})

			control.RegisterJob(&Job{
				Name:     "time.print",
				Interval: time.Second,
				Task: func() {
					fmt.Printf("time print\n")
				},
			})

			time.Sleep(time.Second * 10)
		})
	}
}
