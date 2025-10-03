package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// jwtSecret JWT签名密钥
	jwtSecret  = []byte("gm-fabric-deployment-secret")
	sessionTTL = 24 * time.Hour // 会话有效期
)

// Claims 自定义JWT声明结构体
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

const (
	UserClaims     = "user_claims"
	UserId         = "user_id"
	UserName       = "user_name"
	UserRole       = "user_role"
	SessionId      = "session_id"
	UserType       = "user_type"
	PermissionList = "permission_list"
	ClientIP       = "client_ip"
	UserAgent      = "user_agent"
	DeviceID       = "device_id"
	Scope          = "scope"
	TenantID       = "tenant_id"
)
