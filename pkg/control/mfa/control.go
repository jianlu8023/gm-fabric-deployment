package mfa

import (
	"context"
	"fmt"
	"sync"
	
	// "github.com/google/uuid"
	// "github.com/pquerna/otp"
	// "github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

// Control MFA控制结构体
// @description 实现MFAControl接口的控制结构体
// @struct Control
type Control struct {
	config          *config.MFAConfig   // MFA配置
	logger          *zap.SugaredLogger  // 日志记录器
	ctx             context.Context     // 上下文
	cancel          context.CancelFunc  // 取消函数
	once            sync.Once           // 确保StartUp只执行一次
	providers       map[string]Provider // 认证提供商映射
	currentProvider Provider            // 当前使用的认证提供商
}

// Provider 认证提供商接口
// @interface Provider
// @method GenerateSecret(userID string) (string, string, error) 生成MFA密钥
// @method VerifyCode(userID string, code string) bool 验证MFA代码
type Provider interface {
	// GenerateSecret 生成MFA密钥
	// @param userID string 用户ID
	// @return string 密钥
	// @return string 二维码URL
	// @return error 生成过程中的错误
	GenerateSecret(userID string) (string, string, error)

	// VerifyCode 验证MFA代码
	// @param secret string 用户ID
	// @param code string MFA代码
	// @return bool 验证结果
	VerifyCode(secret string, code string) bool
}

// NewMFAControl 创建一个新的MFA控制器
// @param mfaConfig *config.MFAConfig MFA配置
// @param loggerControl *logger.Control 日志控制器
// @return MFAControl MFA控制器实例
// @return error 创建过程中的错误
func NewMFAControl(mfaConfig *config.MFAConfig, loggerControl *logger.Control) (*Control, error) {
	if mfaConfig == nil {
		mfaConfig = getDefaultConfig()
	}
	if !mfaConfig.Enabled {
		return nil, fmt.Errorf("mfa is not enabled")
	}

	mfaLogger := loggerControl.GenLogger(logger.ModuleMFA)
	mfaLogger.Infof("[control] starting new MFA control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化控制器
	control := &Control{
		logger:    mfaLogger,
		config:    mfaConfig,
		ctx:       ctx,
		cancel:    cancel,
		providers: make(map[string]Provider),
	}

	// 初始化认证提供商
	control.initProviders()

	// 设置默认认证提供商
	if err := control.SetProvider(mfaConfig.DefaultProvider); err != nil {
		mfaLogger.Errorf("[control] failed to set default provider: %v", err)
		cancel()
		return nil, err
	}

	mfaLogger.Infof("[control] MFA control started successfully")
	return control, nil
}

// initProviders 初始化认证提供商
// @description 初始化Google和Microsoft认证提供商
func (c *Control) initProviders() {
	// 初始化Google认证提供商
	c.providers[ProviderGoogle] = &GoogleProvider{
		issuer: c.config.Google.Issuer,
	}
	// 初始化Microsoft认证提供商
	c.providers[ProviderMicrosoft] = &MicrosoftProvider{

		tenantID: c.config.Microsoft.TenantID,
		clientID: c.config.Microsoft.ClientID,
	}
}

// GenerateSecret 生成MFA密钥
// @param userID string 用户ID
// @return string 密钥
// @return string 二维码URL
// @return error 生成过程中的错误
func (c *Control) GenerateSecret(userID string) (string, string, error) {
	c.logger.Debugf("[control] generating MFA secret for user: %s", userID)
	return c.currentProvider.GenerateSecret(userID)
}

// VerifyCode 验证MFA代码
// @param userID string 用户ID
// @param code string MFA代码
// @return bool 验证结果
func (c *Control) VerifyCode(userID string, code string) bool {
	c.logger.Debugf("[control] verifying MFA code for user: %s", userID)
	return c.currentProvider.VerifyCode(userID, code)
}

// SetProvider 设置认证提供商
// @param provider string 认证提供商 (google/microsoft)
// @return error 设置过程中的错误
func (c *Control) SetProvider(provider string) error {
	c.logger.Debugf("[control] setting MFA provider to: %s", provider)

	// 检查提供商是否存在
	p, exists := c.providers[provider]
	if !exists {
		err := fmt.Errorf("unsupported provider: %s", provider)
		c.logger.Errorf("[control] %v", err)
		return err
	}

	c.currentProvider = p
	c.logger.Infof("[control] MFA provider set to: %s", provider)
	return nil
}

// GetProvider 获取当前认证提供商
// @return string 当前认证提供商
func (c *Control) GetProvider() string {
	// 遍历providers映射，找到当前provider对应的键
	for k, v := range c.providers {
		if v == c.currentProvider {
			return k
		}
	}
	return ""
}

// StartUp 启动MFA服务
// @param failedFunc func(err error) 启动失败回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		// 验证配置
		if c.config.Enabled {
			c.logger.Debugf("[control] starting up MFA service...")
			// 验证默认提供商是否配置正确
			if _, exists := c.providers[c.config.DefaultProvider]; !exists {
				err := fmt.Errorf("invalid default provider: %s", c.config.DefaultProvider)
				c.logger.Errorf("[control] %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}

			// 根据提供商类型进行特定验证
			if c.config.DefaultProvider == ProviderGoogle {
				if c.config.Google.Issuer == "" {
					err := ErrMissingGoogleIssuer
					c.logger.Errorf("[control] %v", err)
					if failedFunc != nil {
						failedFunc(err)
					}
					return
				}
			} else if c.config.DefaultProvider == ProviderMicrosoft {
				if c.config.Microsoft.TenantID == "" {
					err := ErrMissingMicrosoftTenantID
					c.logger.Errorf("[control] %v", err)
					if failedFunc != nil {
						failedFunc(err)
					}
					return
				}
				if c.config.Microsoft.ClientID == "" {
					err := ErrMissingMicrosoftClientID
					c.logger.Errorf("[control] %v", err)
					if failedFunc != nil {
						failedFunc(err)
					}
					return
				}
			}

			c.logger.Infof("[control] MFA service started successfully")
		}
	})
}

// Shutdown 关闭MFA服务
// @return error 关闭过程中的错误
func (c *Control) Shutdown() error {
	c.logger.Infof("[control] shutting down MFA control...")
	c.cancel()
	c.logger.Infof("[control] MFA control shutdown successfully")
	return nil
}
