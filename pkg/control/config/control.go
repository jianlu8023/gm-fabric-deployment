package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jianlu8023/golang-example/pkg/control/flags"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/spf13/viper"
)

// Control 控制器
type Control struct {
	flagsControl *flags.Control
	config       *Config
	mutex        sync.RWMutex
	started      bool
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

// GetCaptchaConfig 获取验证码配置
// @return *CaptchaConfig 验证码配置
func (c *Control) GetCaptchaConfig() *CaptchaConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.CaptchaConfig
}

// GetEmailConfig 获取邮件配置
// @return *EmailConfig 邮件配置
func (c *Control) GetEmailConfig() *EmailConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.EmailConfig
}

// GetAntsPoolConfig 获取Ants线程池配置
// @return *AntsPoolConfig Ants线程池配置
func (c *Control) GetAntsPoolConfig() *AntsPoolConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.AntsPoolConfig
}

func (c *Control) GetFabricCAConfig() *FabricCAConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.FabricCAConfig
}

// GetAuthzConfig 获取权限控制配置
// @return *AuthzConfig 权限控制配置
func (c *Control) GetAuthzConfig() *AuthzConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.AuthzConfig
}

func (c *Control) GetIpfsConfig() *IpfsConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.IpfsConfig
}

func (c *Control) GetIpfsClusterConfig() *IpfsClusterConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.IpfsClusterConfig
}

// GetMFAConfig 获取MFA配置
// @return *MFAConfig MFA配置
func (c *Control) GetMFAConfig() *MFAConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.MFAConfig
}

func (c *Control) GetTracerConfig() *TracerConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.TracerConfig
}

// GetRedisConfig 获取Redis配置
// @return *RedisConfig Redis配置
func (c *Control) GetRedisConfig() *RedisConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.RedisConfig
}

// GetKvDatabaseConfig 获取KvDatabase配置
// @return *KvDatabaseConfig KvDatabase配置
func (c *Control) GetKvDatabaseConfig() *KvDatabaseConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.KvDatabaseConfig
}

func (c *Control) GetWebRTCConfig() *WebRTCConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.WebRTCConfig
}

// GetAIConfig 获取AI配置
// @return *AIConfig AI配置
func (c *Control) GetAIConfig() *AIConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.AIConfig
}

// GetCertificateConfig 获取证书配置
// @return *CertificateConfig 证书配置
func (c *Control) GetCertificateConfig() *CertificateConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.CertificateConfig
}

// GetGeoIPConfig 获取GeoIP配置
// @return *GeoIPConfig GeoIP配置
func (c *Control) GetGeoIPConfig() *GeoIPConfig {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.config.GeoIPConfig
}

// GetTunnyPoolConfig 获取Tunny线程池配置
// @return *TunnyPoolConfig Tunny线程池配置
// func (c *Control) GetTunnyPoolConfig() *TunnyPoolConfig {
// 	c.mutex.RLock()
// 	defer c.mutex.RUnlock()
// 	return c.config.TunnyPoolConfig
// }

// Flush 重新加载配置文件
// @description 从配置文件重新加载配置，支持配置热更新
// @return error 重新加载过程中的错误
func (c *Control) Flush() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	fmt.Printf("开始热加载配置...\n")
	newConfig, err := c.loadConfig()
	if err != nil {
		fmt.Printf("热加载配置失败: %v\n", err)
		return fmt.Errorf("热加载配置失败: %w", err)
	}

	// 验证新配置
	if err := c.validateConfig(&newConfig); err != nil {
		fmt.Printf("新配置验证失败: %v\n", err)
		return fmt.Errorf("新配置验证失败: %w", err)
	}

	oldConfig := c.config
	c.config = &newConfig
	fmt.Printf("配置热加载成功\n")
	fmt.Printf("旧配置: %s\n", oldConfig.String())
	fmt.Printf("新配置: %s\n", newConfig.String())
	return nil
}

// WatchDog 监控配置文件变化
// @description 监控配置文件的变化，当配置文件发生变化时可能触发相应操作
func (c *Control) WatchDog() {
	fmt.Printf("watching %s\n", c.flagsControl.GetConfigPath())
}

// StartUp 启动配置服务器
// @description 启动配置服务器，支持多次调用以实现重新初始化
// @param failedFunc func(err error) 启动失败回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.mutex.Lock()
	if c.started {
		c.mutex.Unlock()
		fmt.Printf("config server already started, skip startup\n")
		return
	}
	c.started = true
	c.mutex.Unlock()

	// fmt.Printf("starting up config server...\n")

	// 检查是否请求显示版本信息
	// 注意：这里只打印版本信息，不直接返回，因为后续流程（如 NewServerControlFromFile）
	// 依赖配置已加载，需要继续执行 loadConfig 以保证 c.config 不为 nil
	if c.flagsControl != nil && c.flagsControl.IsVersionRequested() {
		c.flagsControl.PrintVersion()
	}

	// 加载配置文件
	config, err := c.loadConfig()
	if err != nil {
		fmt.Printf("[control] load config failed: %s\n", err)
		c.mutex.Lock()
		c.started = false
		c.mutex.Unlock()
		if failedFunc != nil {
			failedFunc(err)
			return
		}
		return
	}
	c.config = &config

	if c.flagsControl.IsDebugMode() {
		c.printConfig()
	}
}

// Shutdown 关闭配置服务器
// @description 关闭配置服务器，重置启动状态以支持重新初始化
// @return error 关闭过程中的错误
func (c *Control) Shutdown() error {
	// no-op
	fmt.Printf("shutting down config server...\n")
	c.mutex.Lock()
	c.started = false
	c.mutex.Unlock()
	return nil
}

// validateConfig 验证配置的合法性
// @description 验证配置的合法性，确保关键配置项满足约束条件
// @param cfg *Config 需要验证的配置对象
// @return error 验证过程中发现的错误
func (c *Control) validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("配置对象为空")
	}

	// 验证HTTP配置
	if cfg.HttpConfig != nil {
		if cfg.HttpConfig.Enabled {
			if cfg.HttpConfig.Address == "" {
				return errors.New("HTTP服务启用时地址不能为空")
			}
			if cfg.HttpConfig.TlsEnabled {
				if cfg.HttpConfig.TlsCertFile == "" || cfg.HttpConfig.TlsKeyFile == "" {
					return errors.New("TLS启用时必须配置证书和密钥文件路径")
				}
			}
		}
	}

	// 验证gRPC配置
	if cfg.GrpcConfig != nil && cfg.GrpcConfig.Enabled {
		if cfg.GrpcConfig.Server != nil {
			if cfg.GrpcConfig.Server.TlsEnabled {
				if cfg.GrpcConfig.Server.TlsCertFile == "" || cfg.GrpcConfig.Server.TlsKeyFile == "" {
					return errors.New("gRPC TLS启用时必须配置证书和密钥文件路径")
				}
			}
		}
	}

	// 验证JWT配置
	if cfg.HttpConfig != nil && cfg.HttpConfig.Auth != nil {
		if cfg.HttpConfig.Auth.Enabled {
			if cfg.HttpConfig.Auth.JWT != nil && cfg.HttpConfig.Auth.JWT.Enabled {
				if cfg.HttpConfig.Auth.JWT.Secret == "" {
					return errors.New("JWT启用时Secret不能为空")
				}
				if cfg.HttpConfig.Auth.JWT.SessionTTL <= 0 {
					return errors.New("JWT SessionTTL必须大于0")
				}
			}
		}
	}

	// 验证数据源配置
	if cfg.DataSourceConfig != nil && cfg.DataSourceConfig.Enabled {
		if cfg.DataSourceConfig.DataSourceType == "" {
			return errors.New("数据源启用时DataSourceType不能为空")
		}
		if cfg.DataSourceConfig.TLSEnabled {
			if cfg.DataSourceConfig.TLSCertFile == "" || cfg.DataSourceConfig.TLSKeyFile == "" {
				return errors.New("数据源TLS启用时必须配置证书和密钥文件路径")
			}
		}
	}

	// 验证Docker配置
	if cfg.DockerConfig != nil && cfg.DockerConfig.Enabled {
		if cfg.DockerConfig.TlsEnabled {
			if cfg.DockerConfig.TlsCertFile == "" || cfg.DockerConfig.TlsKeyFile == "" {
				return errors.New("Docker TLS启用时必须配置证书和密钥文件路径")
			}
		}
	}

	return nil
}

// loadConfig 加载配置文件
// @return Config 加载的配置对象
// @return error 加载配置过程中的错误
func (c *Control) loadConfig() (Config, error) {
	var cfg Config

	// 获取配置文件名（不包含扩展名）
	fileName := filepath.Base(c.flagsControl.GetConfigPath())
	fileExt := filepath.Ext(c.flagsControl.GetConfigPath())
	fileNameWithoutExt := strings.TrimSuffix(fileName, fileExt)
	filePath := filepath.Dir(c.flagsControl.GetConfigPath())

	// 使用局部viper实例，避免全局状态污染
	v := viper.New()
	v.SetConfigName(fileNameWithoutExt)               // 设置配置文件名
	v.SetConfigType(strings.TrimPrefix(fileExt, ".")) // 设置配置文件类型 (yaml, json, toml 等)
	v.AddConfigPath(filePath)                         // 设置配置文件路径

	// 如果指定了 configType, 尝试读取特定环境的配置文件
	if !stringer.IsBlank(c.flagsControl.GetConfigType()) {
		envSpecificFileName := fmt.Sprintf("%s-%s", fileNameWithoutExt, c.flagsControl.GetConfigType())
		v.SetConfigName(envSpecificFileName) // 尝试读取特定环境的配置文件

		// 尝试读取特定环境的配置文件. 不报错, 如果不存在就继续尝试读取默认配置文件
		err := v.ReadInConfig()
		if err == nil {
			fmt.Printf("使用配置文件: %s\n", v.ConfigFileUsed())
			if err := v.Unmarshal(&cfg); err != nil {
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
	v.SetConfigName(fileNameWithoutExt) // 恢复默认文件名
	v.AddConfigPath(filePath)
	err := v.ReadInConfig()
	if err != nil {
		// 如果找不到配置文件，返回错误
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return cfg, fmt.Errorf("未找到配置文件: %s", c.flagsControl.GetConfigPath())
		}
	}

	fmt.Printf("使用配置文件: %s\n", v.ConfigFileUsed())

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 判断libp2p是否有配置 没有配置不检查identity
	if cfg.Libp2pConfig != nil {
		// 判断libp2p 是否设置了privKey peerId
		if cfg.Libp2pConfig.Identity == nil ||
			stringer.IsBlank(cfg.Libp2pConfig.Identity.PrivKey) ||
			stringer.IsBlank(cfg.Libp2pConfig.Identity.PeerID) {
			ident, err := CreateIdentity(Ed25519, -1)
			if err != nil {
				return cfg, fmt.Errorf("生成libp2p身份失败: %w", err)
			}
			cfg.Libp2pConfig.Identity = &ident

			v.Set("libp2p.identity", cfg.Libp2pConfig.Identity)
			// 获取当前配置文件路径，用于写回
			configFilePath := v.ConfigFileUsed()
			if configFilePath != "" {
				if err = v.WriteConfigAs(configFilePath); err != nil {
					return cfg, fmt.Errorf("更新libp2p的identity失败: %w", err)
				}
			}
		}
	}

	return cfg, nil
}

// NewConfigControl 新建config控制器
// @param flagsControl *flags.Control 命令行参数控制器
// @return *Control 控制器
// @return error 新建过程中可能产生的错误
func NewConfigControl(flagsControl *flags.Control) *Control {
	return &Control{
		flagsControl: flagsControl,
	}
}
