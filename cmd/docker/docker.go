package main

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	dockermount "github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

func main() {
	loggerConfig := &config.LoggerConfig{
		DefaultLogLevel: "debug",
		StackLogLevel:   "error",
		PrintFormat:     "console",
		FilePath:        "./logs/app.log",
		MaxAge:          7,
		RotationTime:    1,
		LoggerLevel: map[string]string{
			"main": "debug",
			"grpc": "debug",
		},
	}

	loggerControl := logger.NewLoggerControl(loggerConfig)
	dockerConfig := &config.DockerConfig{
		Host:           "unix:///var/run/docker.sock",
		APIVersion:     "",
		TlsEnabled:     false,
		TlsCertFile:    []string{},
		TlsKeyFile:     []string{},
		TlsCAFile:      "",
		DefaultTimeout: 5,
	}

	dockerControl := docker.NewDockerControl(dockerConfig, loggerControl)
	dockerControl.StartUp(func(err error) {
		fmt.Printf("docker start failed: %v\n", err)
	})

	defer dockerControl.Shutdown()

	//
	// imageList, err := dockerControl.ListImages()
	// if err != nil {
	// 	fmt.Printf("docker image list failed: %v\n", err)
	// 	return
	// }
	// for _, image := range imageList {
	// 	fmt.Printf("image: %v\n", image)
	// }
	//
	// fmt.Printf("----------------------------------------\n")
	// image, err := dockerControl.GetImage(docker.WithImageQueryName("golang:1.22"))
	// if err != nil {
	// 	fmt.Printf("get docker image failed: %v\n", err)
	// 	return
	// }
	// fmt.Printf("image: %v\n", image)
	//
	// fmt.Printf("----------------------------------------\n")
	//
	// network, err := dockerControl.GetNetwork(docker.WithNetworkQueryID("58865190862c938c7db1069de764e98365cb582a50a53baf20818be633eae2eb"))
	// if err != nil {
	// 	fmt.Printf("get docker network failed: %v\n", err)
	// 	return
	// }
	// fmt.Printf("network: %v\n", network)

	// if err = dockerControl.RemoveNetwork("net_test"); err != nil {
	// 	fmt.Printf("remove docker network failed: %v\n", err)
	// 	return
	// }
	//
	// createNetwork, err := dockerControl.CreateNetwork("net_test", "bridge",
	// 	docker.WithNetworkCreateEnableIPv6(false),
	// 	docker.WithNetworkCreateLabels(map[string]string{
	// 		"key1": "value1",
	// 		"app":  "gm-fabric-deployment",
	// 	}),
	// 	docker.WithNetworkCreateIPAM(&network.IPAM{
	// 		Driver: "default",
	// 		Config: []network.IPAMConfig{
	// 			{
	// 				Subnet:  "99.99.97.0/24",
	// 				Gateway: "99.99.97.254",
	// 				IPRange: "99.99.97.0/24",
	// 			},
	// 		},
	// 		Options: map[string]string{
	// 			"key1": "value1",
	// 		},
	// 	}),
	// )
	// if err != nil {
	// 	fmt.Printf("create network failed: %v\n", err)
	// 	return
	// }
	// fmt.Printf("create network: %v\n", createNetwork)

	// inspect, err := dockerControl.InspectNetwork("aee00a60d88e")
	// if err != nil {
	// 	fmt.Printf("inspect network failed: %v\n", err)
	// 	return
	// }
	//
	// network, err := dockerControl.GetNetwork(docker.WithNetworkQueryID("aee00a60d88e"))
	// if err != nil {
	// 	fmt.Printf("get network failed: %v\n", err)
	// 	return
	// }
	// fmt.Printf("inspect network: %v\n", inspect)
	// fmt.Printf("get network: %v\n", network)

	// createContainer, err := dockerControl.CreateContainer("mynginx", "nginx:latest",
	// 	&container.Config{
	// 		Hostname:   "test",
	// 		Domainname: "test",
	// 		User:       "root",
	// 		Image:      "nginx:latest",
	// 		ExposedPorts: map[nat.Port]struct{}{
	// 			"80/tcp": {},
	// 		},
	// 	},
	// 	&container.HostConfig{
	// 		PortBindings: map[nat.Port][]nat.PortBinding{
	// 			"80/tcp": {
	// 				{HostIP: "0.0.0.0", HostPort: "8000"},
	// 			},
	// 		},
	// 		NetworkMode: network.NetworkBridge,
	// 		RestartPolicy: container.RestartPolicy{
	// 			Name: container.RestartPolicyUnlessStopped,
	// 		},
	// 		LogConfig: container.LogConfig{
	// 			Type: "json-file",
	// 			Config: map[string]string{
	// 				"max-size": "10m",
	// 				"max-file": "3",
	// 			},
	// 		},
	// 		VolumeDriver: "overlay2",
	// 		Mounts: []dockermount.Mount{
	// 			{
	// 				Type:     "bind",
	// 				Source:   "/etc/timezone",
	// 				Target:   "/etc/timezone",
	// 				ReadOnly: true,
	// 			},
	// 			{
	// 				Type:     "bind",
	// 				Source:   "/etc/localtime",
	// 				Target:   "/etc/localtime",
	// 				ReadOnly: true,
	// 			},
	// 		},
	// 	},
	// 	&network.NetworkingConfig{
	// 		EndpointsConfig: map[string]*network.EndpointSettings{
	// 			"net_test": {
	// 				IPAMConfig: &network.EndpointIPAMConfig{
	// 					IPv4Address: "99.99.97.100",
	// 				},
	// 				Aliases:   []string{"nginx"},
	// 				NetworkID: "aee00a60d88e",
	// 				// IPAddress: "99.99.97.100",
	// 			},
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	fmt.Printf("create container failed: %v\n", err)
	// 	return
	// }
	// fmt.Printf("create container: %v\n", createContainer)
	//
	// if err = dockerControl.StartContainer(createContainer); err != nil {
	// 	fmt.Printf("start container failed: %v\n", err)
	// 	return
	// }
	//
	// containers, err := dockerControl.ListContainers(true)
	// if err != nil {
	// 	fmt.Printf("list containers failed: %v\n", err)
	// 	return
	// }
	// fmt.Printf("list containers: %v\n", containers)
	//
	// time.Sleep(5 * time.Second)
	// if err = dockerControl.StopContainer(createContainer, 50); err != nil {
	// 	fmt.Printf("stop container failed: %v\n", err)
	// 	return
	// }
	//
	// if err = dockerControl.RemoveContainer(createContainer, true); err != nil {
	// 	fmt.Printf("remove container failed: %v\n", err)
	// 	return
	// }

	network, err := dockerControl.CreateNetwork("fabric_test_net", "bridge",
		docker.WithNetworkCreateEnableIPv6(false),
		docker.WithNetworkCreateAttachable(true),
		docker.WithNetworkCreateLabels(map[string]string{
			"name": "fabric_test_net",
			"desc": "deploy fabric blockchain",
		}),
		docker.WithNetworkCreateIPAM(&network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{
				{
					Subnet:  "99.99.99.0/24",
					Gateway: "99.99.99.254",
				},
			},
		}),
	)
	if err != nil {
		fmt.Printf("create network failed: %v\n", err)
		return
	}

	createContainer, err := dockerControl.CreateContainer("baasCA", "gcbaas-gm/fabric-ca:1.5.2",
		&container.Config{
			Image: "gcbaas-gm/fabric-ca:1.5.2",
			Env: []string{
				"TZ=Asia/Shanghai",
				"FABRIC_CA_SERVER_DEBUG=true",
				"FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server",
				"FABRIC_CA_SERVER_CA_NAME=baasCA",
				"FABRIC_CA_SERVER_TLS_ENABLED=true",
				"FABRIC_CA_SERVER_PORT=7054",
				"BOOTSTRAP_USER_PASS=gmadmin:gmadminpw",
			},
			User: "1000",
			ExposedPorts: map[nat.Port]struct{}{
				"7054/tcp": {},
			},
			Cmd: strslice.StrSlice{
				"sh", "-c", "/usr/local/bin/start_ca.sh",
			},
		},
		&container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{
				"7054/tcp": {
					{HostIP: "0.0.0.0", HostPort: "7054"},
				},
			},
			Mounts: []dockermount.Mount{
				{
					Type:   "bind",
					Source: "/home/user/codes/go/src/github.com/jianlu8023/golang-example/test/baasca",
					Target: "/etc/hyperledger/fabric-ca-server",
				},
				{
					Type:   "bind",
					Source: "/home/user/codes/go/src/github.com/jianlu8023/golang-example/test/baasca/logs",
					Target: "/logs",
				},
			},
		},
		&dockernetwork.NetworkingConfig{
			EndpointsConfig: map[string]*dockernetwork.EndpointSettings{
				network.Name: {
					NetworkID: network.ID,
					IPAMConfig: &dockernetwork.EndpointIPAMConfig{
						IPv4Address: "99.99.99.1",
					},
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("create container error: %v\n", err)
		return
	}

	if err = dockerControl.StartContainer(createContainer); err != nil {
		fmt.Printf("start container error: %v\n", err)
		return
	}

	// time.Sleep(30 * time.Second)
	// if err = dockerControl.StopContainer(createContainer, 50); err != nil {
	// 	fmt.Printf("stop container error: %v\n", err)
	// 	return
	// }
	// if err = dockerControl.RemoveContainer(createContainer, true); err != nil {
	// 	fmt.Printf("remove container error: %v\n", err)
	// 	return
	// }
	//
	// fmt.Printf("success...")

}
