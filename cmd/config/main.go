package main

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
)

func main() {

	// if err := genDefaultConfig(); err != nil {
	// 	fmt.Printf("gen default config failed: %v\n", err)
	// 	return
	// }

	loadConfig, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}
	fmt.Println(loadConfig.Config)
}

func genDefaultConfig() error {

	cfg := config.Config{
		GrpcConfig: &config.GrpcConfig{
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
			PrintFormat:     "console",
			FilePath:        "./logs/app.log",
			MaxAge:          7,
			RotationTime:    1,
			LoggerLevel: map[string]string{
				"main": "info",
				"grpc": "debug",
			},
		},
		HttpConfig: &config.HttpServerConfig{
			Address:     ":8080",
			RunMode:     "release",
			ContextPath: "example",
			TlsEnabled:  true,
			TlsCertFile: "./certs/hserver.crt",
			TlsKeyFile:  "./certs/hserver.key",
		},
	}

	if err := saveConfig(cfg, "configs/default.yaml"); err != nil {
		fmt.Printf("save config failed: %v\n", err)
		return err
	}
	return nil
}

// saveConfig 将 Config 结构体写回到配置文件
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

	// 写入配置文件
	err := viper.WriteConfigAs(configPath)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("成功将配置写入文件: %s\n", configPath)
	return nil
}
