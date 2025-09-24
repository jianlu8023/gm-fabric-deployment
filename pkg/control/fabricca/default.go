package fabricca

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.FabricCAConfig {
	return &config.FabricCAConfig{
		ImageName:           "hyperledger/fabric-ca:1.4.9",
		CAName:              "baasCA",
		EnabledTls:          true,
		CAAdminUser:         "baasAdmin",
		CAAdminPassword:     "baasAdminpw",
		CAServerPort:        7054,
		JoinDockerNetwork:   "bridge",
		DockerNetworkIpAddr: "",
		LocalAbsPath:        "/tmp/fabric-ca",
		LogInConsole:        true,
		TestMode:            true,
	}
}
