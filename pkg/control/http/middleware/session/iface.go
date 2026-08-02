package session

// SessionStore 会话存储核心接口
//
// @description 抽象会话存储层的读写生命周期，屏蔽底层实现（内存/Redis/DB），供 auth 包与认证中间件调用；监控类能力见 SessionInspector
// @interface
type SessionStore interface {
	// Get 获取会话信息
	//
	// @description 根据 sessionID 读取会话；不存在或已过期时返回错误，过期会话会被异步删除
	// @param sessionID string 会话唯一标识
	// @return *Session 会话数据
	// @return error 错误信息
	Get(sessionID string) (*Session, error)

	// Set 设置会话信息
	//
	// @description 写入或更新会话，由调用方保证 ExpiresAt 正确
	// @param sessionID string 会话唯一标识
	// @param sess *Session 会话数据
	// @return error 错误信息
	Set(sessionID string, sess *Session) error

	// Delete 删除会话信息
	//
	// @description 删除指定会话，用于登出/强制下线
	// @param sessionID string 会话唯一标识
	// @return error 错误信息
	Delete(sessionID string) error

	// Validate 验证会话是否有效
	//
	// @description 等价于 Get + 过期检查，返回布尔结果；内部已检查过期，外层无需重复比较
	// @param sessionID string 会话唯一标识
	// @return bool 是否有效
	Validate(sessionID string) bool
}

// SessionInspector 会话只读监控接口
//
// @description 抽象会话存储的只读监控能力，供运营/管理面板使用；与 SessionStore 拆分以遵循接口隔离原则，监控场景不应被迫实现 Set/Delete
// @interface
type SessionInspector interface {
	// Count 获取当前会话数量
	//
	// @description 用于监控/运营统计
	// @return int 会话数量
	Count() int

	// GetAll 获取所有会话
	//
	// @description 返回当前存储中的全部会话快照，数据量大时慎用
	// @return []*Session 会话列表
	GetAll() []*Session
}

// Session 会话数据模型
//
// @description 承载纯会话业务字段，与 jwt.Claims 解耦；jwt 解析出的 claims 由 auth 包映射为 Session
// @struct
type Session struct {
	SessionID        string   `json:"session_id,omitempty" yaml:"session_id,omitempty"`                 // 会话唯一标识
	UserID           string   `json:"user_id,omitempty" yaml:"user_id,omitempty"`                       // 用户ID
	Username         string   `json:"user_name,omitempty" yaml:"user_name,omitempty"`                   // 用户名
	Role             string   `json:"user_role,omitempty" yaml:"user_role,omitempty"`                   // 用户角色
	UserType         string   `json:"user_type,omitempty" yaml:"user_type,omitempty"`                   // 用户类型
	PermissionList   []string `json:"permission_list,omitempty" yaml:"permission_list,omitempty"`       // 权限列表
	ClientIP         string   `json:"client_ip,omitempty" yaml:"client_ip,omitempty"`                   // 客户端IP
	UserAgent        string   `json:"user_agent,omitempty" yaml:"user_agent,omitempty"`                 // User-Agent
	DeviceID         string   `json:"device_id,omitempty" yaml:"device_id,omitempty"`                   // 设备ID
	TenantID         string   `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty"`                   // 租户ID
	Scope            string   `json:"scope,omitempty" yaml:"scope,omitempty"`                           // 授权范围
	LoginTime        int64    `json:"login_time,omitempty" yaml:"login_time,omitempty"`                 // 登录时间（unix 秒）
	LastActivityTime int64    `json:"last_activity_time,omitempty" yaml:"last_activity_time,omitempty"` // 最后活动时间（unix 秒）
	ExpiresAt        int64    `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`                 // 过期时间（unix 秒）
}
