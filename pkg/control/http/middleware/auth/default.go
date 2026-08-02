package auth

// 默认值常量
//
// @description 仅在 AuthConfig 未提供对应字段时作为兜底默认值使用
const (
	// defaultTokenHeader 默认 token 头名称
	defaultTokenHeader = "Authorization"
	// defaultTokenPrefix 默认 token 前缀
	defaultTokenPrefix = "Bearer"
	// defaultFreshWindow 默认敏感操作二次校验窗口（秒）
	defaultFreshWindow = 300
	// defaultSessionTTLSeconds 默认会话有效期（秒），与 jwt/session 包保持一致
	defaultSessionTTLSeconds = 86400
)

// currentSessionKey gin.Context 中当前会话的存取键
//
// @description 不导出，外部统一通过 CurrentSession(ctx) 等辅助函数读取，避免直接操作 ctx 字符串键
const currentSessionKey = "auth_current_session"
