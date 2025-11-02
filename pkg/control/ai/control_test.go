package ai

import (
	"testing"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

func Test_example(t *testing.T) {
	ExampleUsage(&config.AIConfig{
		Enabled:            true,
		APIKey:             "",
		APIEndpoint:        "",
		DefaultModel:       "",
		Timeout:            9999,
		InsecureSkipVerify: false,
	}, logger.NewLoggerControl(nil))
}
