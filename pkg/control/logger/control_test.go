package logger

import (
	"testing"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
)

func Test_NewLoggerControl(t *testing.T) {
	testCases := []struct {
		name   string
		config *config.LoggerConfig
	}{
		{
			name:   "nil",
			config: &config.LoggerConfig{},
		},
	}

	for _, testcase := range testCases {
		loggerControl := NewLoggerControl(testcase.config)
		if loggerControl == nil {
			t.Errorf("Test_NewLoggerControl failed because logger config is nil")
		}
		if loggerControl != nil {
			testLogger := loggerControl.GenLogger(testcase.name)
			testLogger.Debugf("debug msg...")
			testLogger.Infof("info msg...")
			testLogger.Warnf("warn msg...")
			testLogger.Errorf("error msg...")
			testLogger.Fatalf("fatal msg...")
		}
	}

}
