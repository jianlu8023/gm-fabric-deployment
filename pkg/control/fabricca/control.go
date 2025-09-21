package fabricca

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	dockermount "github.com/docker/docker/api/types/mount"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/client/msp"
	fabconfig "github.com/hxx258456/fabric-sdk-go-gm/pkg/core/config"
	"github.com/hxx258456/fabric-sdk-go-gm/pkg/fabsdk"
	str "github.com/jianlu8023/go-tools/v2/pkg/helper/stringer"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	config        *config.FabricCAConfig
	dockerControl *docker.Control
	loggerControl *logger.Control
	logger        *zap.SugaredLogger
	fabsdk        *fabsdk.FabricSDK
	mspClient     *msp.Client
	once          sync.Once
	mutex         sync.RWMutex
}

func NewFabricCAControl(
	fabriccaConfig *config.FabricCAConfig,
	loggerControl *logger.Control,
	dockerControl *docker.Control,
) (*Control, error) {
	fabriccaLogger := loggerControl.GenLogger("fabricca")

	fabriccaLogger.Debugf("[control] starting generate fabric ca control...")

	control := &Control{
		config:        fabriccaConfig,
		logger:        fabriccaLogger,
		dockerControl: dockerControl,
		loggerControl: loggerControl,
	}

	return control, nil
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting up fabric ca server...")

		c.logger.Debugf("[control] get docker network info...")
		network, err := c.dockerControl.GetNetwork(docker.WithNetworkQueryName(c.config.JoinDockerNetwork))
		if err != nil {
			c.logger.Errorf("[control] get docker network failed: %s", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
		if str.IsBlank(network.Name) {
			c.logger.Errorf("[control] docker network name is blank")
			if failedFunc != nil {
				failedFunc(fmt.Errorf("fabric ca must join exist network"))
			}
			return
		}

		image, err := c.dockerControl.GetImage(docker.WithImageQueryName(c.config.ImageName))
		if err != nil {
			c.logger.Errorf("[control] get docker image failed: %s", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}

		if str.IsBlank(image.ID) {
			c.logger.Errorf("[control] docker image id is blank")
			if failedFunc != nil {
				failedFunc(fmt.Errorf("fabric ca must use exist image"))
			}
			return
		}

		c.logger.Debugf("[control] checkup docker container...")

		logsPath := filepath.Clean(filepath.Join(c.config.LocalAbsPath, "logs"))
		if _, err := os.Stat(logsPath); err != nil {
			c.logger.Debugf("[control] checkup logs dir failed: %v", err)
			if os.IsNotExist(err) {
				if err := os.MkdirAll(logsPath, os.FileMode(0o755)); err != nil {
					c.logger.Errorf("[control] create logs dir failed: %v", err)
					if failedFunc != nil {
						failedFunc(err)
						return
					}
				}
			} else {
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
		}

		containerExist, err := c.dockerControl.GetContainer(c.config.CAName)
		if err != nil {
			c.logger.Errorf("[control] get docker container failed: %s", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}

		if !str.IsBlank(containerExist.ID) {
			c.logger.Debugf("[control] %v container already exist, remove it...", c.config.CAName)
			if err := c.dockerControl.RemoveContainer(containerExist.ID, true); err != nil {
				c.logger.Errorf("[control] remove docker container failed: %s", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
		}

		// 默认 不开放fabric ca 的 server 端口
		hostConfig := &dockercontainer.HostConfig{
			Mounts: []dockermount.Mount{
				{
					Type:   "bind",
					Source: logsPath,
					Target: "/logs",
				},
				{
					Type:   "bind",
					Source: filepath.Clean(filepath.Join(c.config.LocalAbsPath)),
					Target: "/etc/hyperledger/fabric-ca-server",
				},
			},
		}

		if c.config.TestMode {
			c.logger.Warnf("[control] test mode, open fabric ca server port...")
			// 测试模式，开放 fabric ca 的 server 端口
			hostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
				nat.Port(fmt.Sprintf("%v/tcp", c.config.CAServerPort)): {
					{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%v", c.config.CAServerPort)},
				},
			}
		}

		containerId, err := c.dockerControl.CreateContainer(c.config.CAName, c.config.ImageName,
			&dockercontainer.Config{
				Image:    c.config.ImageName,
				User:     "1000",
				Hostname: c.config.CAName,
				Env: []string{
					"TZ=Asia/Shanghai",
					"FABRIC_CA_SERVER_DEBUG=true",
					"FABRIC_CA_SERVER_AFFILIATIONS=baas",
					"FABRIC_CA_HOME=/etc/hyperledger/fabric-ca-server",
					fmt.Sprintf("FABRIC_CA_SERVER_CSR_HOSTS=localhost,%v", c.config.CAName),
					fmt.Sprintf("FABRIC_CA_SERVER_CA_NAME=%v", c.config.CAName),
					fmt.Sprintf("FABRIC_CA_SERVER_TLS_ENABLED=%v", c.config.EnabledTls),
					fmt.Sprintf("FABRIC_CA_SERVER_PORT=%v", c.config.CAServerPort),
					fmt.Sprintf("BOOTSTRAP_USER_PASS=%v:%v", c.config.CAAdminUser, c.config.CAAdminPassword),
				},
				Cmd: strslice.StrSlice{
					"sh", "-c", "/usr/local/bin/start_ca.sh",
				},
				ExposedPorts: map[nat.Port]struct{}{
					nat.Port(fmt.Sprintf("%v/tcp", c.config.CAServerPort)): {},
				},
			},
			hostConfig,
			&dockernetwork.NetworkingConfig{
				EndpointsConfig: map[string]*dockernetwork.EndpointSettings{
					network.Name: {
						NetworkID: network.ID,
						IPAMConfig: &dockernetwork.EndpointIPAMConfig{
							IPv4Address: c.config.DockerNetworkIpAddr,
						},
						Aliases: []string{
							c.config.CAName,
						},
					},
				},
			},
		)
		if err != nil {
			c.logger.Errorf("[control] create docker container failed: %s", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}

		c.logger.Debugf("[control] start docker container...")

		if err = c.dockerControl.StartContainer(containerId); err != nil {
			c.logger.Errorf("[control] start docker container failed: %s", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
		time.Sleep(10 * time.Second)
		c.logger.Debugf("[control] start generate config.json...")
		configPath, err := c.genFabricSdkConfig()
		if err != nil {
			c.logger.Errorf("[control] generate fabric sdk config failed: %s", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
		configProvider := fabconfig.FromFile(configPath)
		sdk, err := fabsdk.New(configProvider,
			fabsdk.WithLoggerPkg(
				newFabricSDKLogger(
					c.loggerControl.GetConfig(),
					c.config.LogInConsole,
				),
			),
		)
		if err != nil {
			c.logger.Errorf("[control] generate fabric sdk failed: %v", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
		c.fabsdk = sdk

		if err = c.checkup(); err != nil {
			c.logger.Errorf("[control] checkup docker container failed: %s", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
	})
}

func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutting down fabric ca server...")

	container, err := c.dockerControl.GetContainer(c.config.CAName)
	if err != nil {
		c.logger.Errorf("[control] get docker container failed: %s", err)
		return err
	}

	if !str.IsBlank(container.ID) {
		if err := c.dockerControl.StopContainer(container.ID, 50); err != nil {
			c.logger.Errorf("[control] stop docker container failed: %s", err)
			return err
		}
		if err := c.dockerControl.RemoveContainer(container.ID, true); err != nil {
			c.logger.Errorf("[control] remove docker container failed: %s", err)
			return err
		}
	}
	c.fabsdk.Close()
	return nil
}

func (c *Control) checkup() error {
	c.logger.Debugf("[control] checkup fabric ca server...")

	clientProvider := c.fabsdk.Context(fabsdk.WithOrg(c.config.CAName))

	mspClient, err := msp.New(clientProvider)
	if err != nil {
		c.logger.Errorf("[control] create msp client failed: %s", err)
		return err
	}
	c.mspClient = mspClient

	if err = mspClient.Enroll(c.config.CAAdminUser, msp.WithSecret(c.config.CAAdminPassword)); err != nil {
		c.logger.Errorf("[control] enroll failed: %s", err)
		return err
	}
	c.logger.Infof("[control] checkup fabric ca server success...")
	return nil

}

func (c *Control) RegisterUserAndGetCert(user *model.UserInfo) error {
	c.logger.Debugf("[control] starting register a user from fabric ca server...")

	if c.mspClient == nil {
		return errors.New("msp client is nil")
	}
	c.logger.Debugf("[control] register user: %s", user.Username)
	_, err := c.mspClient.Register(&msp.RegistrationRequest{
		Name:           user.Username,
		Type:           "client",
		MaxEnrollments: -1,
		// affiliations
		Attributes: []msp.Attribute{
			{
				Name: "user.email", Value: user.Email, ECert: true,
			},
			{
				Name: "user.type", Value: user.UserType, ECert: true,
			},
		},
		Secret: user.Password,
		CAName: c.config.CAName,
	})
	if err != nil {
		c.logger.Errorf("[control] register user failed: %s", err)
		return err
	}

	if err = c.EnrollUser(user); err != nil {
		c.logger.Errorf("[control] enroll user failed: %s", err)
		return err
	}

	c.logger.Debugf("[control] get user %s certificate...", user.Username)
	userIdentity, err := c.mspClient.GetSigningIdentity(user.Username)
	if err != nil {
		c.logger.Errorf("[control] get signing identity failed: %s", err)
		return err
	}

	c.logger.Debugf("[control] setting user certificate...")
	user.Certificate = string(userIdentity.EnrollmentCertificate())

	return nil
}

func (c *Control) EnrollUser(user *model.UserInfo) error {
	c.logger.Debugf("[control] enroll user: %s", user.Username)
	if c.mspClient == nil {
		return errors.New("msp client is nil")
	}

	if err := c.mspClient.Enroll(user.Username, msp.WithSecret(user.Password)); err != nil {
		c.logger.Errorf("[control] enroll user failed: %s", err)
		return err
	}
	return nil

}
