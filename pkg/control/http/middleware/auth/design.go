package auth

// =====================================================================================
// auth 中间件包设计文档（design note）
// =====================================================================================
//
// 一、定位与职责
// -------------------------------------------------------------------------------------
// auth 包是认证与授权的"统一编排层"（门面/Facade），组合 jwt 包（无状态：签名/解析）
// 与 session 包（有状态：会话存储/失效），对外暴露 Authenticator 接口与中间件族，
// 替代原 jwt.EnableJWT 的职责。
//
// 拆分动机：
//   - 原 jwt 包同时承担 token 与 session 两类职责，导致 "jwt == session" 的概念混淆
//   - RBAC/权限/多租户/可选认证等更"上层"的认证能力无处安放
//   - 后续若换 token 方案（Paseto/Opaque Token），只需替换 jwt 包，auth 包不动
//
// 依赖方向（无循环）：
//   - auth → jwt（使用 jwt.JwtManager 做 token 生成/解析）
//   - auth → session（使用 session.SessionStore 做会话存取）
//   - session 不反向依赖 jwt
//
// 二、核心抽象（实现见 authenticator.go）
// -------------------------------------------------------------------------------------
// Authenticator 接口（对外暴露，Control 持有此接口）：
//   - Middleware() gin.HandlerFunc            必选认证中间件（替代 jwt.EnableJWT）
//   - OptionalMiddleware() gin.HandlerFunc    可选认证（无 token 放行，有 token 则解析注入）
//   - Login(userID, username, role) (token, err)   登录：生成 token + 创建会话
//   - Logout(sessionID) error                      登出：删除会话（token 即使签名合法也失效）
//   - Refresh(sessionID) (token, err)              续期会话，必要时重签 token
//
// authManagerImpl 结构体（不导出，构造函数返回接口）：
//   - jwtManager   jwt.JwtManager
//   - sessionStore session.SessionStore
//   - logger       *zap.SugaredLogger
//   - cfg          *config.AuthConfig
//   - mode         authMode（自动推断：jwt_session / jwt_only / session_only）
//
// 三、模式自动推断（不设 mode 配置项，靠 jwt.enabled + session.enabled 推断）
// -------------------------------------------------------------------------------------
// | jwt.enabled | session.enabled | 模式           | 说明                          |
// | true        | true            | jwt_session    | JWT 当载体 + 服务端会话可失效 |
// | true        | false           | jwt_only       | 纯无状态 JWT，无法主动失效    |
// | false       | true            | session_only   | sessionID 模式（cookie/header）|
// | false       | false           | 启动报错       | auth.enabled=true 但子开关都 false |
//
// 四、EnableAuth 流程（替代 jwt.EnableJWT，见 middleware.go）
// -------------------------------------------------------------------------------------
//   1. tracer.StartSpan(ctx, "ginMiddleware", "auth")，把 trace ctx 写回 ctx.Request
//      （修正原 jwt 包未写回 trace ctx 的链路断裂问题）
//   2. 按 mode 分流：
//      - jwt_session:  取 Bearer token → jwtManager.ParseToken → sessionStore.Validate
//                       → 更新 LastActivityTime + sessionStore.Set 续期
//      - jwt_only:      取 Bearer token → jwtManager.ParseToken
//      - session_only:  按 token_source 从 header/cookie/query 取 sessionID → sessionStore.Get
//   3. 认证通过：构造 *session.Session 注入 ctx，并 ctx.Set 各上下文键常量
//   4. 认证失败：通过 commonhttp.BaseResponse 返回 401，AbortWithStatusJSON 后不再调 ctx.Abort()
//
// 五、门禁中间件（不进接口，作为导出函数，见 middleware.go）
// -------------------------------------------------------------------------------------
//   - RequireRole(roles...)         角色门禁（粗粒度）
//   - RequirePermission(perms...)   权限门禁（粗粒度）
//   - RequireTenant(tenantID)       多租户隔离
//   - RequireFresh(windowSeconds)   敏感操作二次校验（会话最近 N 分钟内有活动）
// 与 Casbin 的边界：本包只做粗粒度硬规则，细粒度策略（带客体的资源授权）仍由 Casbin 处理。
//
// 六、登录/登出（见 login.go）
// -------------------------------------------------------------------------------------
// Login(userID, username, role):
//   1. 生成 sessionID（uuid）
//   2. 构造 session.Session{... ExpiresAt: now+ttl, LoginTime: now, LastActivityTime: now}
//   3. sessionStore.Set（若启用 session）
//   4. jwtManager.GenerateToken（若启用 jwt）→ token
//   5. session_only 模式返回 sessionID 作为 "token"
//
// Logout(sessionID):
//   1. sessionStore.Delete（若启用 session）
//   2. jwt_only 模式无操作（无服务端状态可删，token 自然过期）
//
// 七、辅助函数（见 helper.go）
// -------------------------------------------------------------------------------------
//   - CurrentSession(ctx) *session.Session   从 gin.Context 取当前会话
//   - CurrentUserID(ctx) string              从 gin.Context 取 UserID
//   - CurrentRole(ctx) string                从 gin.Context 取 Role
//   - HasPermission(ctx, perm) bool          判断当前会话是否持有某权限
//
// 八、配置（见 pkg/control/config/config.go 的 AuthConfig）
// -------------------------------------------------------------------------------------
//   auth:
//     enabled: true
//     token_source: "auto"        # header | cookie | query | auto
//     token_header: "Authorization"
//     token_prefix: "Bearer"
//     sliding_expiration: true
//     fresh_window: 300
//     optional_paths: ["/healthz"]
//     jwt:
//       enabled: true
//       secret: xxx
//       session_ttl: 86400
//     session:
//       enabled: true
//       ttl: 86400
//       store_type: "memory"
//
// 九、与 AGENTS.md 规范的对齐
// -------------------------------------------------------------------------------------
// - 中间件统一使用 commonhttp.BaseResponse 返回封装
// - 统一使用 *zap.SugaredLogger 记录日志
// - 统一调用 tracer.StartSpan 记录链路，且把 trace ctx 写回 ctx.Request
// - AbortWithStatusJSON 后不再多余调用 ctx.Abort()
// - 注释遵循 @description/@param/@return/@struct/@interface 规范
// - 所有 .go 文件最后一行为空行
// - 结构体字段小写不导出，通过方法暴露能力
//
// =====================================================================================
