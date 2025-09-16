package fabricca

import (
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"os"
	"path/filepath"
)

func (c *Control) genFabricSdkConfig() (string, error) {
	var url string
	if c.config.EnabledTls {
		url = fmt.Sprintf("https://https://%s:%d", c.config.DockerNetworkIpAddr, c.config.CAServerPort)
	} else {
		url = fmt.Sprintf("http://%s:%d", c.config.DockerNetworkIpAddr, c.config.CAServerPort)
	}

	m := map[string]interface{}{
		"version": "1.0.0",
		"name":    "gm-fabric-baas-ca",
		"client": map[string]interface{}{
			"organization": c.config.CAName,
			"logging": map[string]string{
				"level": "debug",
			},
		},
		"organizations": map[string]interface{}{
			c.config.CAName: map[string]interface{}{
				"mspid":      fmt.Sprintf("%vMSP", c.config.CAName),
				"cryptoPath": filepath.Clean(filepath.Join(c.config.LocalAbsPath, "msp")),
				"certificateAuthorities": []string{
					c.config.CAName,
				},
			},
		},
		"certificateAuthorities": map[string]interface{}{
			c.config.CAName: map[string]interface{}{
				"url": url,
				"tlsCACerts": map[string]interface{}{
					"path": filepath.Clean(filepath.Join(c.config.LocalAbsPath, "tls-cert.pem")),
				},
				"caName": c.config.CAName,
				"registrar": map[string]interface{}{
					"enrollId":     c.config.CAAdminUser,
					"enrollSecret": c.config.CAAdminPassword,
				},
			},
		},
	}

	configPath := filepath.Clean(filepath.Join(c.config.LocalAbsPath, "config.json"))

	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, os.FileMode(0o755)); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	file, err := atomicfile.New(configPath, os.FileMode(0o644))
	if err != nil {
		_ = file.Abort()
		return "", err
	}
	defer func(file *atomicfile.File) {
		if err := file.Close(); err != nil {
			_ = file.Abort()
		}
	}(file)

	if err = json.NewEncoder(file).Encode(m); err != nil {
		_ = file.Abort()
		return "", nil
	}

	return configPath, nil
}
