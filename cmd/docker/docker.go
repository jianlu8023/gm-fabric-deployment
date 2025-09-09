package main

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	mydocker "github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
)

func main() {

	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	loggerControl := mylogger.NewLoggerControl(configControl.GetLoggerConfig())
	dockerConfig := &config.DockerConfig{
		Host:           "unix:///var/run/docker.sock",
		APIVersion:     "",
		TlsEnabled:     false,
		TlsCertFile:    "",
		TlsKeyFile:     "",
		TlsCAFile:      "",
		DefaultTimeout: 5,
	}

	dockerControl, err := mydocker.NewDockerControl(dockerConfig, loggerControl)
	if err != nil {
		fmt.Printf("load docker failed: %v\n", err)
		return
	}
	dockerControl.Start(func(err error) {
		fmt.Printf("docker start failed: %v\n", err)
	})

	defer dockerControl.Shutdown()

	imageList, err := dockerControl.ImageList()
	if err != nil {
		fmt.Printf("docker image list failed: %v\n", err)
		return
	}
	for _, image := range imageList {
		fmt.Printf("image: %v\n", image)
	}

}
