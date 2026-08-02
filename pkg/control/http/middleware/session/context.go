package session

// 上下文键常量
//
// @description 定义 gin.Context 中会话相关信息的存取键，由 auth 中间件在认证通过后写入，handler 层用同名常量读取；从 jwt 包迁入以解耦会话概念与 token 概念
const (
	// SessionId 会话唯一标识
	SessionId = "session_id"
	// UserId 用户ID
	UserId = "user_id"
	// UserName 用户名
	UserName = "user_name"
	// UserRole 用户角色
	UserRole = "user_role"
	// UserType 用户类型
	UserType = "user_type"
	// PermissionList 权限列表
	PermissionList = "permission_list"
	// ClientIP 客户端IP
	ClientIP = "client_ip"
	// UserAgent User-Agent
	UserAgent = "user_agent"
	// DeviceID 设备ID
	DeviceID = "device_id"
	// Scope 授权范围
	Scope = "scope"
	// TenantID 租户ID
	TenantID = "tenant_id"
)
