package main

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
)

func main() {

	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	dockerConfig := &config.DockerConfig{
		Host:           "unix:///var/run/docker.sock",
		APIVersion:     "",
		TlsEnabled:     false,
		TlsCertFile:    "",
		TlsKeyFile:     "",
		TlsCAFile:      "",
		DefaultTimeout: 5,
	}

	dockerControl, err := docker.NewDockerControl(dockerConfig, loggerControl)
	if err != nil {
		fmt.Printf("load docker failed: %v\n", err)
		return
	}
	dockerControl.StartUp(func(err error) {
		fmt.Printf("docker start failed: %v\n", err)
	})

	defer dockerControl.Shutdown()

	imageList, err := dockerControl.ListImages()
	if err != nil {
		fmt.Printf("docker image list failed: %v\n", err)
		return
	}
	for _, image := range imageList {
		fmt.Printf("image: %v\n", image)
	}

}
