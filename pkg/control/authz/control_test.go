package authz

import (
	"os"
	"testing"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
)

func TestNewAuthzControl(t *testing.T) {
	// 1. 初始化日志控制器
	loggerConfig := &config.LoggerConfig{
		DefaultLogLevel: "debug",
		PrintFormat:     "console",
	}
	loggerControl := logger.NewLoggerControl(loggerConfig)

	// 2. 创建权限控制配置
	authzConfig := &config.AuthzConfig{
		Enabled:        true,
		ModelFile:      "configs/authz/model.conf",
		PolicyFile:     "configs/authz/policy.csv",
		AutoCreateFile: true,
		DefaultAllow:   false,
		LogEnabled:     true,
	}

	// 3. 初始化权限控制器
	authzControl := NewAuthzControl(authzConfig, loggerControl)
	_ = authzControl
}

func TestExample(t *testing.T) {
	// 1. 初始化日志控制器
	loggerConfig := &config.LoggerConfig{
		DefaultLogLevel: "debug",
		PrintFormat:     "console",
	}
	loggerControl := logger.NewLoggerControl(loggerConfig)

	// 2. 创建权限控制配置
	authzConfig := &config.AuthzConfig{
		Enabled:        true,
		ModelFile:      "configs/authz/model.conf",
		PolicyFile:     "configs/authz/policy.csv",
		AutoCreateFile: true,
		DefaultAllow:   false,
		LogEnabled:     true,
	}

	// 3. 初始化权限控制器
	authzControl := NewAuthzControl(authzConfig, loggerControl)
	authzControl.StartUp(func(err error) {
		if err != nil {
			t.Logf("failed to start authz control: %v\n", err)
			os.Exit(1)
		}
	})

	// 4. 基本权限操作示例
	t.Logf("===== 基本权限操作示例 ====")

	// 添加权限策略
	ok, err := authzControl.AddPolicy("admin", "resource1", "read")
	if err != nil {
		t.Errorf("failed to add policy: %v\n", err)
	} else {
		t.Logf("Add policy result: %v\n", ok)
	}

	// 检查权限
	ok, err = authzControl.CheckPermission("admin", "resource1", "read")
	if err != nil {
		t.Errorf("failed to check permission: %v\n", err)
	} else {
		t.Logf("Permission check result: admin can read resource1: %v\n", ok)
	}

	// 检查没有的权限
	ok, err = authzControl.CheckPermission("admin", "resource2", "read")
	if err != nil {
		t.Errorf("failed to check permission: %v\n", err)
	} else {
		t.Logf("Permission check result: admin can read resource2: %v\n", ok)
	}

	// 5. 角色操作示例
	t.Logf("===== 角色操作示例 ====")

	// 添加角色
	ok, err = authzControl.AddRoleForUser("user1", "admin")
	if err != nil {
		t.Errorf("failed to add role for user: %v\n", err)
	} else {
		t.Logf("Add role for user result: %v\n", ok)
	}

	// 获取用户的角色
	roles, err := authzControl.GetRolesForUser("user1")
	if err != nil {
		t.Errorf("failed to get roles for user: %v\n", err)

	} else {
		t.Logf("Roles for user1: %v\n", roles)
	}

	// 6. 保存策略
	t.Logf("===== 保存策略 ====")

	err = authzControl.SavePolicy()
	if err != nil {
		t.Errorf("failed to save policy: %v\n", err)

	} else {
		t.Logf("Policy saved successfully\n")
	}
	t.Logf("===== 示例运行完成 ====")
}
