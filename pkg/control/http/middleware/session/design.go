package session

// =====================================================================================
// session 中间件包设计文档（design note）
// =====================================================================================
//
// 一、定位与职责
// -------------------------------------------------------------------------------------
// session 包专注于"会话生命周期管理"，与 token 的签名/解析彻底解耦：
//   - jwt 包：仅负责 JWT 的生成、签名校验、claims 解析（无状态部分）
//   - session 包：负责会话的存储、查询、删除、过期清理、续期（有状态部分）
//   - auth 包：编排 jwt 与 session，对外提供统一认证/授权中间件
//
// 历史背景：jwt 包内曾持有 SessionManager 接口与 MemorySessionManager 实现
// （见原 jwt/iface.go、jwt/memory.go），导致 "jwt == session" 的概念混淆。
// 本次重构将 session 相关能力迁出至本包，jwt 包瘦身为纯 token 引擎。
//
// 二、核心抽象（实现见 iface.go）
// -------------------------------------------------------------------------------------
// 1. Session 数据模型
//    独立于 jwt.Claims，承载纯会话业务字段，避免 session 包反向依赖 jwt 包：
//      SessionID / UserID / Username / Role / UserType / PermissionList
//      ClientIP / UserAgent / DeviceID / TenantID / Scope
//      LoginTime / LastActivityTime / ExpiresAt
//    auth 包负责 jwt.Claims ↔ Session 的互相映射。
//
// 2. SessionStore 接口（核心读写，认证主链路用）
//      Get(sessionID) (*Session, error)
//      Set(sessionID, *Session) error
//      Delete(sessionID) error
//      Validate(sessionID) bool
//
// 3. SessionInspector 接口（只读监控，运营/管理面板用）
//      Count() int
//      GetAll() []*Session
//    拆分依据：接口隔离原则，监控场景只需读，不应被迫实现 Set/Delete。
//    memoryStoreImpl 同时实现两者，消费者按需依赖小接口。
//
// 三、实现（见 memory.go）
// -------------------------------------------------------------------------------------
// memoryStoreImpl 基于 github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent 的 RWMap：
//   - 读写并发安全（直接用 RWMap，不手写 sync.RWMutex）
//   - 后台 goroutine 定期清理过期会话（默认 1h，可配）
//   - 通过 ctx.Done() 优雅退出清理协程
//   - Get 内部检查 ExpiresAt < now → 同步删除并返回错误（2026-08-04 修正：原异步 go Delete
//     存在"逻辑竞态"——异步删除执行前若有 Set 重新写入同一 sessionID 会误删新会话；
//     且高并发下创建大量 goroutine。concurrent.Map.Del 为 O(1) 加锁操作，同步删除开销极低）
//   - cleanupLoop 先收集待删除 key 再批量删除，并调用 iterator.Close() 释放快照切片引用
//     （2026-08-04 修正：原在迭代过程中调用 Del 虽因快照迭代器不会 panic，但逻辑混淆且漏调 Close）
//   - Validate 直接复用 Get，返回 err == nil，不重复比较 ExpiresAt
//
// 四、RedisStore 实现（未来扩展，本次不实现）
// -------------------------------------------------------------------------------------
// 依赖 pkg/control/kvdatabase 控制器，分布式部署使用：
//   - key 规则：session:{sessionID}
//   - TTL 由 Redis 自带过期管理，省去清理协程
//   - 支持跨实例会话共享，支持横向扩容
//
// 五、上下文键常量（见 context.go）
// -------------------------------------------------------------------------------------
// 从 jwt/default.go 迁入本包：SessionId / UserId / UserName / UserRole / UserType
// PermissionList / ClientIP / UserAgent / DeviceID / Scope / TenantID
// jwt 包仅保留 UserClaims（持有 *jwt.Claims）。
// auth 中间件在认证通过后，用这些常量 ctx.Set 注入用户信息，handler 层用同样的常量取。
//
// 六、配置（见 pkg/control/config/config.go 的 SessionConfig）
// -------------------------------------------------------------------------------------
//   session:
//     enabled: true
//     ttl: 86400                 # 会话有效期（秒）
//     cleanup_interval: 3600     # 过期清理间隔（秒），仅 memory 模式
//     store_type: "memory"       # memory | redis
//     sliding_expiration: true   # 滑动过期 vs 绝对过期
//     session_id_source: "auto"  # auto | header | cookie | query（仅 session_only 模式）
//     cookie_name: "GSESSIONID"  # cookie 模式名称（可选）
//
// 七、中间件
// -------------------------------------------------------------------------------------
// 本包不提供认证中间件。认证中间件统一由 auth 包提供（auth.EnableAuth），
// 以保持"session = 数据存储，auth = 流程编排"的职责边界。
//
// 八、与 AGENTS.md 规范的对齐
// -------------------------------------------------------------------------------------
// - 注释遵循 @description/@param/@return/@struct/@interface 规范
// - 所有 .go 文件最后一行为空行
// - 接口小且稳，符合接口隔离原则
// - 结构体字段小写不导出，通过方法暴露能力
//
// =====================================================================================
