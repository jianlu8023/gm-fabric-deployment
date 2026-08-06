package docker

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.DockerConfig {
	return &config.DockerConfig{
		Enabled:        false,
		Host:           "unix:///var/run/docker.sock",
		APIVersion:     "",
		TlsEnabled:     false,
		TlsCertFile:    []string{"certs/dclient.crt"},
		TlsKeyFile:     []string{"certs/dclient.key"},
		TlsCAFile:      "certs/root-cat.crt",
		DefaultTimeout: 5,
	}
}
