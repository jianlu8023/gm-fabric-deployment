package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// 默认值常量
//
// @description 仅在配置未提供对应字段时作为兜底默认值使用，禁止直接以默认 secret 上生产环境
const (
	// defaultSecret 默认JWT签名密钥，仅在配置未提供时使用，生产环境必须通过配置文件覆盖
	defaultSecret = "gm-fabric-deployment-secret"
	// defaultSessionTTLSeconds 默认会话有效期（秒），用于配置默认值
	defaultSessionTTLSeconds = 86400
)

// Claims 自定义JWT声明结构体
//
// @description 在标准JWT声明基础上扩展用户、会话、租户等业务字段
// @struct
type Claims struct {
	jwt.RegisteredClaims
	UserID           string   `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	Username         string   `json:"user_name,omitempty" yaml:"user_name,omitempty"`
	Role             string   `json:"user_role,omitempty" yaml:"user_role,omitempty"`
	SessionID        string   `json:"session_id,omitempty" yaml:"session_id,omitempty"`
	UserType         string   `json:"user_type,omitempty" yaml:"user_type,omitempty"`
	PermissionList   []string `json:"permission_list,omitempty" yaml:"permission_list,omitempty"`
	ClientIP         string   `json:"client_ip,omitempty" yaml:"client_ip,omitempty"`
	UserAgent        string   `json:"user_agent,omitempty" yaml:"user_agent,omitempty"`
	DeviceID         string   `json:"device_id,omitempty" yaml:"device_id,omitempty"`
	SessionExpires   int64    `json:"session_expires,omitempty" yaml:"session_expires,omitempty"`
	LoginTime        int64    `json:"login_time,omitempty" yaml:"login_time,omitempty"`
	LastActivityTime int64    `json:"last_activity_time,omitempty" yaml:"last_activity_time,omitempty"`
	Scope            string   `json:"scope,omitempty" yaml:"scope,omitempty"`
	TenantID         string   `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty"`
}

// 上下文键常量
//
// @description 仅保留 UserClaims（持有 *Claims）；session 相关常量（UserId/UserName/UserRole/SessionId 等）已迁至 session 包
const (
	// UserClaims gin.Context 中 JWT claims 的存取键，auth 中间件认证通过后写入
	UserClaims = "user_claims"
)
