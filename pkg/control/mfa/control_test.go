package mfa

import (
	"testing"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

func TestNewMFAControl(t *testing.T) {
	// 创建日志控制器
	logCtrl := logger.NewLoggerControl(nil)
	logCtrl.StartUp(func(err error) {
		if err != nil {
			t.Fatalf("启动日志控制器失败: %v", err)
		}
	})
	defer logCtrl.Shutdown()

	// 创建MFA控制器
	mfaCtrl, err := NewMFAControl(&config.MFAConfig{
		Enabled:         true,
		DefaultProvider: "google",
		Google: struct {
			Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty" mapstructure:"issuer"`
		}{Issuer: "golang-example"},
	},
		logCtrl)
	if err != nil {
		t.Fatalf("创建MFA控制器失败: %v", err)
	}

	// 启动MFA控制器
	mfaCtrl.StartUp(func(err error) {
		if err != nil {
			t.Fatalf("启动MFA控制器失败: %v", err)
		}
	})
	defer mfaCtrl.Shutdown()

	t.Log("MFA控制器创建和启动成功")
}

func TestMFAControl_GenerateSecret(t *testing.T) {
	// 创建日志控制器
	logCtrl := logger.NewLoggerControl(nil)
	logCtrl.StartUp(func(err error) {
		if err != nil {
			t.Fatalf("启动日志控制器失败: %v", err)
		}
	})
	defer logCtrl.Shutdown()

	// 创建MFA控制器
	mfaCtrl, err := NewMFAControl(&config.MFAConfig{
		Enabled:         true,
		DefaultProvider: "google",
		Google: struct {
			Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty" mapstructure:"issuer"`
		}{Issuer: "golang-example"},
	}, logCtrl)
	if err != nil {
		t.Fatalf("创建MFA控制器失败: %v", err)
	}
	mfaCtrl.StartUp(func(err error) {
		if err != nil {
			t.Fatalf("启动MFA控制器失败: %v", err)
		}
	})
	defer mfaCtrl.Shutdown()

	// 生成Google MFA密钥
	secret, url, err := mfaCtrl.GenerateSecret("exmapleUser")
	if err != nil {
		t.Fatalf("生成Google MFA密钥失败: %v", err)
	}

	t.Logf("Google MFA密钥生成成功: %s", secret)
	t.Logf("Google MFA QR码URL: %s", url)
	image := mfaCtrl.GenerateQrCodeImage(url)
	t.Logf("Google MFA QR码图片: %s", image)
}

func TestMFAControl_VerifyCode(t *testing.T) {
	// 这里仅测试验证逻辑，实际测试需要使用有效的密钥和验证码
	t.Skip("跳过验证测试，需要实际的密钥和验证码")
}
