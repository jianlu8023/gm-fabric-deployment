package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/str"
	"github.com/spf13/viper"
)

// Control 控制器
type Control struct {
	config *Config
	mutex  sync.RWMutex
	once   sync.Once
}

// printConfig 打印配置信息
// @description 打印当前加载的配置信息
func (c *Control) printConfig() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	fmt.Printf("Config:%v\n", c.config)
}

// GetConfig 获取完整配置
// @return *Config 完整配置对象
func (c *Control) GetConfig() *Config {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config
}

// GetLoggerConfig 获取日志配置
// @return *LoggerConfig 日志配置
func (c *Control) GetLoggerConfig() *LoggerConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.LoggerConfig
}

// GetWebConfig 获取Web服务器配置
// @return *HttpServerConfig Web服务器配置
func (c *Control) GetWebConfig() *HttpServerConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.HttpConfig
}

// GetGrpcConfig 获取grpc配置
// @return *GrpcConfig grpc配置
func (c *Control) GetGrpcConfig() *GrpcConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.GrpcConfig
}

// GetLibp2pConfig 获取libp2p配置
// @return *Libp2pConfig libp2p配置
func (c *Control) GetLibp2pConfig() *Libp2pConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.Libp2pConfig
}

// GetDataSourceConfig 获取数据库配置
// @return *DataSourceConfig 数据库配置
func (c *Control) GetDataSourceConfig() *DataSourceConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.DataSourceConfig
}

// GetDockerConfig 获取docker配置
// @return *DockerConfig Docker配置
func (c *Control) GetDockerConfig() *DockerConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.DockerConfig
}

// Flush 重新加载配置文件
// @description 从配置文件重新加载配置
// @return error 重新加载过程中的错误
func (c *Control) Flush() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	newConfig, err := loadConfig()
	if err != nil {
		return err
	}
	c.config = &newConfig
	return nil
}

// WatchDog 监控配置文件变化
// @description 监控配置文件的变化，当配置文件发生变化时可能触发相应操作
func (c *Control) WatchDog() {
	fmt.Printf("watching %s\n", configPath)
}

// StartUp 启动配置服务器
// @description 启动配置服务器（空操作）
// @param failedFunc func(err error) 启动失败回调函数
//nolint:unused
func (c *Control) StartUp(failedFunc func(err error)) {
	// no-op
	c.once.Do(func() {
		fmt.Printf("starting up config server...\n")
	})
}

// Shutdown 关闭配置服务器
// @description 关闭配置服务器（空操作）
// @return error 关闭过程中的错误
func (c *Control) Shutdown() error {
	// no-op
	fmt.Printf("shutting down config server...\n")
	return nil
}

// loadConfig 加载配置文件
// @return Config 加载的配置对象
// @return error 加载配置过程中的错误
func loadConfig() (Config, error) {
	var cfg Config

	// 获取配置文件名（不包含扩展名）
	fileName := filepath.Base(configPath)
	fileExt := filepath.Ext(configPath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)
	filePath := filepath.Dir(configPath)

	viper.SetConfigName(fileNameWithoutExt)               // 设置配置文件名
	viper.SetConfigType(strings.TrimPrefix(fileExt, ".")) // 设置配置文件类型 (yaml, json, toml 等)
	viper.AddConfigPath(filePath)                         // 设置配置文件路径

	// 如果指定了 configType, 尝试读取特定环境的配置文件
	if configType != "" {
		envSpecificFileName := fmt.Sprintf("%s-%s", fileNameWithoutExt, configType)
		viper.SetConfigName(envSpecificFileName) // 尝试读取特定环境的配置文件

		// viper.AddConfigPath(".") // 放在这里是为了优先查找当前目录下的特定环境配置文件
		// 尝试读取特定环境的配置文件. 不报错, 如果不存在就继续尝试读取默认配置文件
		err := viper.ReadInConfig()
		if err == nil {
			fmt.Printf("使用配置文件: %s\n", viper.ConfigFileUsed())
			if err := viper.Unmarshal(&cfg); err != nil {
				return cfg, fmt.Errorf("解析环境特定配置文件失败: %w", err)
			}
			return cfg, nil // 成功读取并解析环境特定配置文件
		} else {
			fmt.Printf("未找到配置文件: %s, 尝试读取默认配置文件. 错误信息: %v\n",
				envSpecificFileName, err)
			// 不需要 return,  继续尝试读取默认配置文件
		}
	}

	// 读取默认配置文件
	viper.SetConfigName(fileNameWithoutExt) // 恢复默认文件名
	viper.AddConfigPath(filePath)
	err := viper.ReadInConfig()
	if err != nil {
		// 如果找不到配置文件，返回错误
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return cfg, fmt.Errorf("未找到配置文件: %s", configPath)
		}
	}

	fmt.Printf("使用配置文件: %s\n", viper.ConfigFileUsed())

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 判断libp2p是否有配置 没有配置不检查identity
	if cfg.Libp2pConfig != nil {
		// 判断libp2p 是否设置了privKey peerId
		if cfg.Libp2pConfig.Identity == nil ||
			str.IsBlank(cfg.Libp2pConfig.Identity.PrivKey) ||
			str.IsBlank(cfg.Libp2pConfig.Identity.PeerID) {
			ident, err := CreateIdentity(Ed25519, -1)
			if err != nil {
				return cfg, fmt.Errorf("生成libp2p身份失败: %w", err)
			}
			cfg.Libp2pConfig.Identity = &ident

			viper.Set("libp2p.identity", cfg.Libp2pConfig.Identity)
			if err = viper.WriteConfig(); err != nil {
				return cfg, fmt.Errorf("更新libp2p的identity失败: %w", err)
			}
		}
	}

	return cfg, nil
}

// NewConfigControl 新建config控制器
// @return *Control 控制器
// @return error 新建过程中可能产生的错误
func NewConfigControl() (*Control, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	ctr := &Control{
		config: &config,
	}

	ctr.printConfig()

	return ctr, nil
}
