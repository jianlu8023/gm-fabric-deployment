package helper

import (
	"strconv"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/sony/sonyflake"
	"go.uber.org/zap"
)

type Control struct {
	logger *zap.SugaredLogger

	sf *sonyflake.Sonyflake
}

func NewHelperControl(loggerControl *logger.Control) *Control {
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	helperLogger := loggerControl.GenLogger("helper")

	sf := sonyflake.NewSonyflake(sonyflake.Settings{})

	control := &Control{
		sf:     sf,
		logger: helperLogger,
	}
	return control
}

func (c *Control) GenID() (string, error) {
	id, err := c.sf.NextID()
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}
