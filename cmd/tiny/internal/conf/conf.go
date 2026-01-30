package conf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"github.com/jianlu8023/go-tools/v2/pkg/nethelper/ip4"
	"github.com/jianlu8023/golang-example/pkg/control/certificate"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	MaxLevel      uint8  = 0           // 允许访问的最大层级
	Port          int    = 8192        // 指定的服务端口
	IsAllowUpload bool   = false       // 是否允许上传
	IsSecure      bool   = false       // 是否开启访问登录
	RootPath      string = "/"         // 共享目录的根路径，默认值：当前目录
	Username      string = "admin"     // 访问登录的帐号
	Password      string = "adminpw"   // 访问登录的密码
	IP            string = "127.0.0.1" // 本机局域网IP
)

type config struct {
	Output io.Writer
	OS     string
	IP     string
	Pwd    string

	TlsEnabled     bool
	TlsCertPath    string
	TlsKeyPath     string
	TlsRCACertPath string

	RootPath      string // 共享目录的根路径，默认值：当前目录
	MaxLevel      uint8  // 允许访问的最大层级，默认值  0
	Port          int    // 指定的服务端口，默认值 9090
	IsAllowUpload bool   // 是否允许上传，默认值：否
	IsSecure      bool   // 是否开启访问登录，默认值：否

	SessionVal string
	Username   string // 访问登录的帐号
	Password   string // 访问登录的密码
}

func GenConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	cfgDir := filepath.Join(home, ".config", "tiny")
	return cfgDir, nil
}

func LoadConfig(logger *zap.SugaredLogger, cert *certificate.Control) error {
	cfgDir, err := GenConfigPath()
	if err != nil {
		logger.Errorf("获取配置目录失败: %v", err)
		return errors.New("获取配置目录失败")
	}

	err = cert.GenerateLeafCertificate("tiny", "/CN=localhost", "RSA", false, cfgDir+"/root.crt", cfgDir+"/root.key", "", "")
	if err != nil {
		logger.Errorf("生成服务证书失败: %v", err)
	}

	cfgFile := filepath.Join(cfgDir, "config.yaml")
	err = loadUserConfig(cfgDir, cfgFile, logger)
	if err != nil {
		logger.Errorf("加载用户配置失败: %v", err)
		return err
	}

	Config.IP, err = ip4.GetLocalIP()
	if err != nil {
		logger.Errorf("获取本机局域网IP失败: %v", err)
		return err
	}

	Config.RootPath = viper.GetString("server.road")
	if len(Config.RootPath) < 1 {
		wd, err := os.Getwd()
		if err != nil {
			logger.Errorf("获取不到共享目录: %v", err)
			return errors.New("获取不到共享路径")
		}
		Config.RootPath = wd
		Config.Pwd = wd
	}

	Config.MaxLevel = uint8(viper.GetInt("server.max_level"))
	Config.Port = viper.GetInt("server.port")
	Config.IsAllowUpload = viper.GetBool("server.allow_upload")
	Config.IsSecure = viper.GetBool("account.secure")
	Config.Username = viper.GetString("account.custom.user")
	Config.Password = viper.GetString("account.custom.pass")
	Config.TlsEnabled = viper.GetBool("server.tls_enabled")
	Config.TlsCertPath = viper.GetString("server.tls_cert_path")
	Config.TlsKeyPath = viper.GetString("server.tls_key_path")
	Config.TlsRCACertPath = viper.GetString("server.tls_rca_cert_path")
	return nil
}

// loadUserConfig 负责加载用户配置文件,如果文件不存在则创建并设置默认值
func loadUserConfig(cfgDir, cfgFile string, logger *zap.SugaredLogger) error {
	viper.AddConfigPath(cfgDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

read:
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			logger.Infof(colour.Yellow("未找到「自定义配置文件」, 正在创建中..."))
			if err := createCfgFile(cfgDir, cfgFile, logger); err != nil {
				logger.Errorf("创建用户配置文件失败: %v", err)
				return err
			}
			logger.Infof(colour.Green(fmt.Sprintf("创建成功，配置文件位于: %s", cfgFile)))
			goto read
		case viper.ConfigParseError:
			return errors.New("已找到「自定义配置文件」，但是解析失败！")
		case viper.ConfigMarshalError:
			return errors.New("已找到「自定义配置文件」，但是读取失败！")
		}
	}
	return nil
}

func createCfgFile(cfgDir, cfgFile string, logger *zap.SugaredLogger) error {
	_, err := os.Stat(cfgDir)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(cfgDir, os.ModePerm)
	}
	_, err = os.Create(cfgFile)
	if err != nil {
		return errors.New("创建自定义配置文件失败！")
	}
	if err = setDefault(cfgDir); err != nil {
		logger.Errorf(colour.Yellow("设置默认配置失败！"))
	}
	return nil
}

func setDefault(cfgDir string) error {
	viper.Set("server.port", Port)
	viper.Set("server.allow_upload", IsAllowUpload)
	viper.Set("server.max_level", MaxLevel)
	viper.Set("server.tls_enabled", true)
	viper.Set("server.tls_cert_path", filepath.Join(cfgDir, "tiny.crt"))
	viper.Set("server.tls_key_path", filepath.Join(cfgDir, "tiny.key"))
	viper.Set("server.tls_rca_cert_path", filepath.Join(cfgDir, "root.crt"))
	return viper.WriteConfig()
}
