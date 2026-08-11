package authz

import (
	"context"
	"fmt"
	"sync"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

// Control 权限控制器
type Control struct {
	config   *config.AuthzConfig
	enforcer *casbin.Enforcer
	logger   *zap.SugaredLogger
	ctx      context.Context
	once     sync.Once
}

// NewAuthzControl 创建权限控制器
// @param authzConfig *config.AuthzConfig 权限控制配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control 权限控制器实例
func NewAuthzControl(authzConfig *config.AuthzConfig, loggerControl *logger.Control) *Control {

	// 验证配置
	authzConfig = check.IF[*config.AuthzConfig](authzConfig == nil,
		getDefaultConfig(),
		authzConfig,
	)

	if !authzConfig.Enabled {
		return nil
	}

	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)

	// authzLogger := loggerControl.GenLogger(logger.ModuleAuthZ)
	authzLogger := loggerControl.GenLogger("")
	authzLogger.Infof("[authz/control] start new authz control...")

	ctx := context.Background()

	// 创建控制器
	control := &Control{
		config: authzConfig,
		logger: authzLogger,
		ctx:    ctx,
	}

	authzLogger.Infof("[authz/control] authz control initialized successfully")
	return control
}

// StartUp 启动权限控制服务
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Infof("[authz/control] starting up authz control...")
			// 如果启用了权限控制，初始化enforcer
			if err := c.initEnforcer(); err != nil {
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
			c.logger.Infof("[authz/control] authz control is enabled and ready")
		}
	})
}

// Shutdown 关闭权限控制服务
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	c.logger.Infof("[authz/control] shutting down authz control...")
	// 权限控制器没有需要特别关闭的资源
	// 主要是记录关闭日志
	c.logger.Infof("[authz/control] authz control shutdown completed")
	return nil
}

// initEnforcer 初始化enforcer
// @return error 错误信息
func (c *Control) initEnforcer() error {
	c.logger.Debugf("[authz/control] initializing enforcer...")

	// 确保文件存在
	if c.config.AutoCreateFile {
		if err := c.ensureAuthzFiles(c.config.ModelFile, c.config.PolicyFile); err != nil {
			return err
		}
	}

	// 加载模型和策略
	model, err := casbinmodel.NewModelFromFile(c.config.ModelFile)
	if err != nil {
		c.logger.Errorf("[authz/control] failed to load model file: %v", err)
		return err
	}

	adapter := fileadapter.NewAdapter(c.config.PolicyFile)

	// 创建enforcer
	enforcer, err := casbin.NewEnforcer(model, adapter)
	if err != nil {
		c.logger.Errorf("[authz/control] failed to create enforcer: %v", err)
		return err
	}

	enforcer.SetAdapter(adapter)
	enforcer.SetModel(model)

	// 启用日志
	enforcer.SetLogger(newAuthzLogger(c.logger, c.config.LogEnabled))
	// enforcer.EnableLog(c.config.LogEnabled)

	// 加载策略
	if err := enforcer.LoadPolicy(); err != nil {
		c.logger.Errorf("[authz/control] failed to load policy: %v", err)
		return err
	}

	c.enforcer = enforcer
	c.logger.Debugf("[authz/control] enforcer initialized successfully")
	return nil
}

// ensureAuthzFiles 确保权限控制文件存在
// @param modelFile string 模型文件路径
// @param policyFile string 策略文件路径
// @return error 错误信息
func (c *Control) ensureAuthzFiles(modelFile, policyFile string) error {
	defaultModel := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act`
	_ = path.WriteToFile(modelFile, defaultModel, false)

	defaultPolicy := `# 格式: p, 角色/用户, 资源, 操作
# 示例:
# p, admin, *, *
# p, user, resource1, read`

	_ = path.WriteToFile(policyFile, defaultPolicy, false)
	return nil
}

// CheckPermission 检查权限
// @param sub string 主体（用户或角色）
// @param obj string 客体（资源）
// @param act string 动作（操作）
// @return bool 是否有权限
// @return error 错误信息
func (c *Control) CheckPermission(sub, obj, act string) (bool, error) {
	// 如果权限控制未启用，默认返回配置的默认值
	if !c.config.Enabled {
		c.logger.Debugf("[authz/control] authz not enabled, default allow: %v", c.config.DefaultAllow)
		return c.config.DefaultAllow, nil
	}

	// 检查enforcer是否初始化
	if c.enforcer == nil {
		c.logger.Errorf("[authz/control] enforcer not initialized")
		return c.config.DefaultAllow, fmt.Errorf("enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] checking permission: sub=%s, obj=%s, act=%s", sub, obj, act)
	ok, err := c.enforcer.Enforce(sub, obj, act)
	if err != nil {
		c.logger.Errorf("[authz/control] permission check failed: %v", err)
		return c.config.DefaultAllow, err
	}

	c.logger.Debugf("[authz/control] permission check result: %v", ok)
	return ok, nil
}

// AddPolicy 添加策略
// @param sub string 主体（用户或角色）
// @param obj string 客体（资源）
// @param act string 动作（操作）
// @return bool 是否添加成功
// @return error 错误信息
func (c *Control) AddPolicy(sub, obj, act string) (bool, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return false, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] adding policy: sub=%s, obj=%s, act=%s", sub, obj, act)
	return c.enforcer.AddPolicy(sub, obj, act)
}

// RemovePolicy 移除策略
// @param sub string 主体（用户或角色）
// @param obj string 客体（资源）
// @param act string 动作（操作）
// @return bool 是否移除成功
// @return error 错误信息
func (c *Control) RemovePolicy(sub, obj, act string) (bool, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return false, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] removing policy: sub=%s, obj=%s, act=%s", sub, obj, act)
	return c.enforcer.RemovePolicy(sub, obj, act)
}

// SavePolicy 保存策略到存储
// @return error 错误信息
func (c *Control) SavePolicy() error {
	if !c.config.Enabled || c.enforcer == nil {
		return fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] saving policy")
	return c.enforcer.SavePolicy()
}

// LoadPolicy 从存储加载策略
// @return error 错误信息
func (c *Control) LoadPolicy() error {
	if !c.config.Enabled || c.enforcer == nil {
		return fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] loading policy")
	return c.enforcer.LoadPolicy()
}

// GetRolesForUser 获取用户的所有角色
// @param user string 用户
// @return []string 角色列表
// @return error 错误信息
func (c *Control) GetRolesForUser(user string) ([]string, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return []string{}, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] getting roles for user: %s", user)
	return c.enforcer.GetRolesForUser(user)
}

// AddRoleForUser 为用户添加角色
// @param user string 用户
// @param role string 角色
// @return bool 是否添加成功
// @return error 错误信息
func (c *Control) AddRoleForUser(user, role string) (bool, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return false, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] adding role for user: user=%s, role=%s", user, role)
	return c.enforcer.AddRoleForUser(user, role)
}

// RemoveRoleForUser 移除用户的角色
// @param user string 用户
// @param role string 角色
// @return bool 是否移除成功
// @return error 错误信息
func (c *Control) RemoveRoleForUser(user, role string) (bool, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return false, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] removing role for user: user=%s, role=%s", user, role)
	return c.enforcer.DeleteRoleForUser(user, role)
}

// GetAllSubjects 获取所有主体
// @return []string 主体列表
// @return error 错误信息
func (c *Control) GetAllSubjects() ([]string, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return []string{}, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] getting all subjects")
	return c.enforcer.GetAllSubjects()
}

// GetAllObjects 获取所有客体
// @return []string 客体列表
// @return error 错误信息
func (c *Control) GetAllObjects() ([]string, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return []string{}, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] getting all objects")
	return c.enforcer.GetAllObjects()
}

// GetAllActions 获取所有动作
// @return []string 动作列表
// @return error 错误信息
func (c *Control) GetAllActions() ([]string, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return []string{}, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] getting all actions")
	return c.enforcer.GetAllActions()
}

// GetAllRoles 获取所有角色
// @return []string 角色列表
// @return error 错误信息
func (c *Control) GetAllRoles() ([]string, error) {
	if !c.config.Enabled || c.enforcer == nil {
		return []string{}, fmt.Errorf("authz not enabled or enforcer not initialized")
	}

	c.logger.Debugf("[authz/control] getting all roles")
	return c.enforcer.GetAllRoles()
}
