package session

// 默认值常量
//
// @description 仅在配置未提供对应字段时作为兜底默认值使用
const (
	// defaultSessionTTLSeconds 默认会话有效期（秒），24 小时
	defaultSessionTTLSeconds = 86400
	// defaultCleanupIntervalSeconds 默认过期清理间隔（秒），1 小时，仅 memory 模式生效
	defaultCleanupIntervalSeconds = 3600
)
