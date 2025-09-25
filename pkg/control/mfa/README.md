# MFA 控制器

MFA（多因素认证）控制器提供了Google和Microsoft两种MFA认证方式的支持，可以通过配置文件控制默认认证方式。

## 功能特性

- 支持Google Authenticator认证
- 支持Microsoft Authenticator认证
- 可通过配置文件切换默认认证提供商
- 提供密钥生成和验证码验证功能

## 配置说明

在配置文件中，MFA配置部分如下：

```json
{
  "MFAConfig": {
    "Enabled": false,
    "DefaultProvider": "google",
    "Google": {
      "Issuer": "golang-example"
    },
    "Microsoft": {
      "TenantID": "common",
      "ClientID": "",
      "ClientSecret": ""
    }
  }
}
```

- `Enabled`: 是否启用MFA认证
- `DefaultProvider`: 默认认证提供商，可选值为"google"或"microsoft"
- `Google`: Google Authenticator配置
  - `Issuer`: 发行者名称
- `Microsoft`: Microsoft Authenticator配置
  - `TenantID`: 租户ID
  - `ClientID`: 客户端ID
  - `ClientSecret`: 客户端密钥

## 使用示例

```go
import (
	"context"
	"github.com/jianlu8023/golang-example/pkg/control/mfa"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

// 创建日志控制器
logCtrl := logger.NewLoggerControl(config.GetDefaultLoggerConfig())
logCtrl.StartUp(func(err error) {
	// 处理错误
})

defers logCtrl.Shutdown()

// 创建配置控制器
cfgCtrl := config.NewConfigControl()
cfgCtrl.StartUp(func(err error) {
	// 处理错误
})

defers cfgCtrl.Shutdown()

// 创建MFA控制器
mfaCtrl := mfa.NewMFAControl(cfgCtrl, logCtrl)
mfaCtrl.StartUp(func(err error) {
	// 处理错误
})

defers mfaCtrl.Shutdown()

// 生成MFA密钥	settings, err := mfaCtrl.GenerateSecret(context.Background(), "google", "user@example.com")
if err != nil {
	// 处理错误
}

// 验证MFA验证码
valid, err := mfaCtrl.VerifyCode(context.Background(), settings.Secret, "123456")
if err != nil {
	// 处理错误
}
if valid {
	// 验证成功
}
```

## 接口说明

MFA控制器实现了以下接口方法：

- `GenerateSecret(ctx context.Context, provider string, username string) (*MFASettings, error)`: 生成MFA密钥
- `VerifyCode(ctx context.Context, secret string, code string) (bool, error)`: 验证MFA验证码
- `SetProvider(provider string) error`: 设置当前使用的认证提供商
- `GetProvider() string`: 获取当前使用的认证提供商
- `StartUp(failedFunc func(err error))`: 启动MFA控制器
- `Shutdown() error`: 关闭MFA控制器