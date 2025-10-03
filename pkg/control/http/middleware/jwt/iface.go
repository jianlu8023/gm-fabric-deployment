package jwt

// SessionManager 会话管理器
type SessionManager interface {
	// GetSession 获取会话信息
	GetSession(sessionID string) (*Claims, error)
	// SetSession 设置会话信息
	SetSession(sessionID string, claims *Claims) error
	// DeleteSession 删除会话信息
	DeleteSession(sessionID string) error
	// ValidateSession 验证会话是否有效
	ValidateSession(sessionID string) bool
}
