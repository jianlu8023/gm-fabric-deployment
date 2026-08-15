package docker

import (
	"runtime"

	"github.com/jianlu8023/golang-example/pkg/control/config"
)

// getDefaultConfig 获取Docker默认配置
//
// @description 根据当前运行平台生成兼容的默认Docker配置，默认未启用
// @return *config.DockerConfig Docker默认配置
func getDefaultConfig() *config.DockerConfig {
	host := "unix:///var/run/docker.sock"
	if runtime.GOOS == "windows" {
		host = "npipe:////./pipe/docker_engine"
	}
	return &config.DockerConfig{
		Enabled:        false,
		Host:           host,
		APIVersion:     "",
		TlsEnabled:     false,
		TlsCertFile:    []string{"certs/dclient.crt"},
		TlsKeyFile:     []string{"certs/dclient.key"},
		TlsCAFile:      "certs/root-ca.crt",
		DefaultTimeout: 5,
	}
}
