package fabricca

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"testing"
)

func Test_genFabricSdkConfig(t *testing.T) {
	c, err := genFabricSdkConfig(&config.FabricCAConfig{
		CAName:              "BaasCA",
		LocalAbsPath:        "/home/user/codes/go/src/github.com/jianlu8023/golang-example/pkg/control/fabricca",
		DockerNetworkIpAddr: "10.0.0.1",
		CAServerPort:        7054,
		EnabledTls:          true,
		CAAdminPassword:     "adminpw",
		CAAdminUser:         "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(c)
}
