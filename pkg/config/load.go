package config

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"sync"
)

type Control struct {
	Config *Config
	sync.RWMutex
}

func (c *Control) Flush() error {
	c.Lock()
	defer c.Unlock()
	newConfig, err := loadConfig()
	if err != nil {
		return err
	}
	c.Config = &newConfig
	return nil
}

func (c *Control) WatchDog() {
	fmt.Printf("watching %s\n", configPath)
}

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

	return cfg, nil
}

func NewConfigControl() (*Control, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return &Control{
		Config: &config,
	}, nil
}
