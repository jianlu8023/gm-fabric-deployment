package auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/session"
	"go.uber.org/zap"
)

// Authenticator 认证器抽象
//
// @description 认证与授权的门面接口，屏蔽底层 token 方案（JWT/Paseto/...）与会话存储（内存/Redis/...）；http.Control 持有此接口，不持有 jwt/session 具体类型
// @interface
type Authenticator interface {
	// Middleware 必选认证中间件
	//
	// @description 无 token 或会话失效时返回 401；替代原 jwt.EnableJWT
	// @return gin.HandlerFunc Gin中间件函数
	Middleware() gin.HandlerFunc

	// OptionalMiddleware 可选认证中间件
	//
	// @description 无 token 时放行（不注入用户信息），有 token 则解析并注入上下文；适用于匿名也可访问、登录后享特权的接口
	// @return gin.HandlerFunc Gin中间件函数
	OptionalMiddleware() gin.HandlerFunc

	// Login 登录
	//
	// @description 生成 sessionID、创建会话（若启用）、签发 token（若启用 JWT）；session_only 模式返回 sessionID 作为 token
	// @param userID string 用户ID
	// @param username string 用户名
	// @param role string 用户角色
	// @return string JWT token 或 sessionID
	// @return error 错误信息
	Login(userID, username, role string) (string, error)

	// Logout 登出
	//
	// @description 删除服务端会话，使 token 即使签名合法也无法再通过 Validate；jwt_only 模式无操作
	// @param sessionID string 会话唯一标识
	// @return error 错误信息
	Logout(sessionID string) error

	// Refresh 续期会话
	//
	// @description 校验会话存在后更新 ExpiresAt，jwt 模式下若 token 即将过期则重签
	// @param sessionID string 会话唯一标识
	// @return string 新的 token（若重签），否则空字符串
	// @return error 错误信息
	Refresh(sessionID string) (string, error)
}

// 为了避免 import 循环风险，gin.HandlerFunc 在 middleware.go 中通过 gin 引入；
// 此处接口签名中 gin.HandlerFunc 需要导入 gin 包，故放在 middleware.go 同包内可见。
// 实际 gin 导入见 middleware.go。

// authMode 认证模式
//
// @description 由 AuthConfig 的 jwt.enabled 与 session.enabled 自动推断，不设显式 mode 配置项以避免配置冲突
type authMode int

const (
	// modeJWTSession JWT + 服务端会话（默认推荐）：JWT 当载体 + 服务端会话可主动失效
	modeJWTSession authMode = iota
	// modeJWTOnly 纯无状态 JWT：无服务端会话，无法主动失效
	modeJWTOnly
	// modeSessionOnly 纯 sessionID 模式：不走 JWT，sessionID 由 header/cookie/query 携带
	modeSessionOnly
)

// authManagerImpl 认证管理器实现
//
// @description 编排 jwt.JwtManager 与 session.SessionStore，对外通过 Authenticator 接口暴露能力；结构体不导出，构造函数返回接口
// @struct
type authManagerImpl struct {
	// jwtManager JWT 管理器，jwt_only/jwt_session 模式下非 nil
	jwtManager jwt.JwtManager
	// sessionStore 会话存储，jwt_session/session_only 模式下非 nil
	sessionStore session.SessionStore
	// logger 日志实例
	logger *zap.SugaredLogger
	// cfg 认证配置
	cfg *config.AuthConfig
	// mode 认证模式（自动推断）
	mode authMode
}

// NewAuthManager 创建认证管理器
//
// @description 注入 jwt 管理器与会话存储，根据 AuthConfig 自动推断认证模式；jwt_only 模式下 sessionStore 可为 nil，session_only 模式下 jwtManager 可为 nil
// @param jwtManager jwt.JwtManager JWT 管理器（可为 nil）
// @param sessionStore session.SessionStore 会话存储（可为 nil）
// @param logger *zap.SugaredLogger 日志实例
// @param cfg *config.AuthConfig 认证配置
// @return Authenticator 认证器实例
func NewAuthManager(jwtManager jwt.JwtManager, sessionStore session.SessionStore, logger *zap.SugaredLogger, cfg *config.AuthConfig) Authenticator {
	m := &authManagerImpl{
		jwtManager:   jwtManager,
		sessionStore: sessionStore,
		logger:       logger,
		cfg:          cfg,
	}
	m.mode = m.resolveMode()
	m.logger.Infof("[http/auth] authenticator created with mode: %s", m.mode)
	return m
}

// String 返回模式名称
//
// @description 用于日志输出
// @return string 模式名称
func (m authMode) String() string {
	switch m {
	case modeJWTSession:
		return "jwt_session"
	case modeJWTOnly:
		return "jwt_only"
	case modeSessionOnly:
		return "session_only"
	default:
		return fmt.Sprintf("unknown(%d)", int(m))
	}
}

// resolveMode 根据配置自动推断认证模式
//
// @description jwt.enabled + session.enabled 的组合决定模式；两者都 false 时回退到 jwt_session 并告警（由调用方在启动时校验合法性）
// @return authMode 认证模式
func (m *authManagerImpl) resolveMode() authMode {
	jwtEnabled := m.cfg != nil && m.cfg.JWT != nil && m.cfg.JWT.Enabled
	sessEnabled := m.cfg != nil && m.cfg.Session != nil && m.cfg.Session.Enabled
	switch {
	case jwtEnabled && sessEnabled:
		return modeJWTSession
	case jwtEnabled:
		return modeJWTOnly
	case sessEnabled:
		return modeSessionOnly
	default:
		// 两者都未启用，回退到 jwt_session 模式（启动校验应由 control 层拦截）
		m.logger.Warnf("[http/auth] both jwt and session disabled, fallback to jwt_session mode")
		return modeJWTSession
	}
}

// getSessionTTL 获取会话有效期
//
// @description 优先从 sessionStore（memoryStoreImpl）获取，其次从 jwtManager 获取，最后用默认值
// @return time.Duration 会话有效期
func (m *authManagerImpl) getSessionTTL() time.Duration {
	if m.cfg != nil && m.cfg.Session != nil && m.cfg.Session.TTL > 0 {
		return time.Duration(m.cfg.Session.TTL) * time.Second
	}
	if m.jwtManager != nil {
		return m.jwtManager.GetSessionTTL()
	}
	return time.Duration(defaultSessionTTLSeconds) * time.Second
}
