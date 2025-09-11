package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/spf13/viper"
)

func main() {

	if err := genDefaultConfig(); err != nil {
		fmt.Printf("gen default config failed: %v\n", err)
		return
	}

	loadConfig, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}
	fmt.Println(loadConfig.GetConfig())
}

func genDefaultConfig() error {
	ident, err := config.CreateIdentity(config.Ed25519, -1)
	if err != nil {
		fmt.Printf("create identity failed: %v\n", err)
		return err
	}
	cfg := config.Config{
		GrpcConfig: &config.GrpcConfig{
			Enabled: true,
			Server: &config.GrpcServerConfig{
				Host:           "127.0.0.1:65534",
				MaxRecvMsgSize: 5242880,
				MaxSendMsgSize: 5242880,
				ChunkSize:      4194304,
				TlsEnabled:     true,
				TlsCertFile:    "./certs/gserver.crt",
				TlsKeyFile:     "./certs/gserver.key",
				TlsRCACertFile: "./certs/root-ca.crt",
			},
			Client: &config.GrpcClientConfig{
				Host:               "127.0.0.1:65533",
				MaxCallRecvMsgSize: 5242880,
				MaxCallSendMsgSize: 5242880,
				ChunkSize:          4194304,
				CallTimeout:        99999,
				TlsEnabled:         true,
				TlsCertFile:        "./certs/gclient.crt",
				TlsKeyFile:         "./certs/gclient.key",
				TlsRCACertFile:     "./certs/root-ca.crt",
			},
		},
		LoggerConfig: &config.LoggerConfig{
			DefaultLogLevel: "debug",
			StackLogLevel:   "fatal",
			PrintFormat:     "console",
			FilePath:        "./logs/app.log",
			MaxAge:          7,
			RotationTime:    1,
			LoggerLevel: map[string]string{
				"main":   "info",
				"grpc":   "debug",
				"libp2p": "info",
			},
		},
		HttpConfig: &config.HttpServerConfig{
			Enabled:     true,
			Address:     ":8080",
			RunMode:     "release",
			ContextPath: "example",
			TlsEnabled:  true,
			TlsCertFile: "./certs/hserver.crt",
			TlsKeyFile:  "./certs/hserver.key",
		},
		Libp2pConfig: &config.Libp2pConfig{
			Enabled: true,
			ListenAddr: []string{
				"/ip4/0.0.0.0/tcp/2000",
				"/ip6/::1/tcp/2000",
				"/ip4/0.0.0.0/udp/2000/quic-v1",
				"/ip4/127.0.0.1/udp/2000/quic-v1",
				"/ip4/0.0.0.0/udp/2000/quic-v1/webtransport",
				"/ip4/0.0.0.0/udp/2000/webrtc-direct",
				"/dns4/localhost/udp/2000/ws",
			}, // 监听所有接口的2000端口
			ProtocolID: "/gm-fabric/chat/1.0.0", // 自定义协议ID
			ServiceTag: "gm-fabric-deployment",
			Identity:   &ident,
			BootstrapList: []string{
				"/ip4/127.0.0.1/tcp/2000/p2p/12D3KooWAiFrSLkqdaz4KC423i7ZgRks8xjrR1CbjnjJ39f8V1f6",
			},
		},
		DataSourceConfig: &config.DataSourceConfig{
			Enabled:        true,
			DataSourceType: "sqlite3",
			DataBaseName:   "db_name",
			DataBasePath:   "./db/gm-fabric.db",
			UserName:       "username",
			Password:       "password",
			Host:           "localhost",
			Port:           3306,
			MaxIdleConn:    10,
			MaxOpenConn:    50,
			LogInConsole:   true,
		},
		DockerConfig: &config.DockerConfig{
			Enabled: true,
			// Host:"tcp://127.0.0.1:2375",
			Host:           "unix:///var/run/docker.sock",
			APIVersion:     "1",
			TlsEnabled:     false,
			TlsCertFile:    "./certs/dserver.crt",
			TlsKeyFile:     "./certs/dserver.key",
			TlsCAFile:      "./certs/root-ca.crt",
			DefaultTimeout: 5,
		},
	}

	if err := saveConfig(cfg, "configs/default.yaml"); err != nil {
		fmt.Printf("save config failed: %v\n", err)
		return err
	}
	return nil
}

// saveConfig 将 config 结构体写回到配置文件
func saveConfig(cfg config.Config, configPath string) error {
	// 获取配置文件名（不包含扩展名）
	fileName := filepath.Base(configPath)
	fileExt := filepath.Ext(configPath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)
	filePath := filepath.Dir(configPath)

	viper.SetConfigName(fileNameWithoutExt)               // 设置配置文件名
	viper.SetConfigType(strings.TrimPrefix(fileExt, ".")) // 设置配置文件类型 (yaml, json, toml 等)
	viper.AddConfigPath(filePath)                         // 设置配置文件路径

	// 使用viper.Set来设置每一个值
	viper.Set("grpc", cfg.GrpcConfig)
	viper.Set("logger", cfg.LoggerConfig)
	viper.Set("http", cfg.HttpConfig)
	viper.Set("libp2p", cfg.Libp2pConfig)
	viper.Set("datasource", cfg.DataSourceConfig)
	viper.Set("docker", cfg.DockerConfig)

	// 写入配置文件
	err := viper.WriteConfigAs(configPath)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("成功将配置写入文件: %s\n", configPath)
	return nil
}
