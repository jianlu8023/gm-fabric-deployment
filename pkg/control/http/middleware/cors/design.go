// Package cors 提供 CORS（跨域资源共享）中间件的配置驱动实现。
//
// ====================================================================
// 设计文档 (design.go)
// ====================================================================
//
// 一、设计目标
//
// 本包旨在提供一个安全、灵活、配置驱动的 CORS 中间件，解决以下历史问题：
//
//  1. 历史问题 1：AllowOrigins:["*"] 与 AllowCredentials:true 同时使用
//     - 违反 W3C CORS 规范，浏览器会拒绝带凭证的跨域请求
//     - gin-contrib/cors v1.7.3 在该组合下会在首次请求时 panic
//  2. 历史问题 2：CORS() 函数无条件回显任意 Origin 并设置 Allow-Credentials:true
//     - 等同于"允许任何网站携带用户 Cookie 访问本服务"
//     - 典型的 CSRF 攻击向量，安全风险极高
//  3. 历史问题 14：EnableCors1 与 CORS 两个函数为死代码，未被引用
//
// 二、设计原则
//
//  1. 配置驱动 (Config-Driven)
//     所有 CORS 设置从 yaml 配置文件读取，源码中不出现任何具体 Origin、Method、Header。
//     配置结构定义在 pkg/control/config.CORSConfig，通过 HttpServerConfig.CORS 注入。
//
//  2. 白名单优先 (Whitelist First)
//     必须显式列出允许的 Origin。当需要携带 Cookie（AllowCredentials=true）时，
//     禁止使用 "*"，必须使用精确 Origin 或子域名通配符（如 https://*.example.com）。
//
//  3. 安全兜底 (Fail-Safe)
//     在 EnableCors 入口处主动校验配置：
//     - 若 AllowOrigins 含 "*" 且 AllowCredentials=true，记录 error 日志并强制
//     将 AllowCredentials 降级为 false，避免运行时 panic 与浏览器拒绝。
//     - 若 AllowOrigins 为空，根据是否开发模式给出默认策略（见下文）。
//
//  4. 复用 gin-contrib/cors (Don't Reinvent the Wheel)
//     不自行拼接 CORS 响应头（旧版 EnableCors1/CORS 即为反面教材，存在回显任意
//     Origin 的 CSRF 漏洞）。本包仅在外层做配置转换与校验，底层逻辑交给成熟库。
//
//  5. 删除死代码 (Delete Dead Code)
//     移除 EnableCors1 与 CORS 函数，避免误用。仅保留唯一的 EnableCors 入口。
//
//  6. 环境感知 (Environment-Aware)
//     当配置缺失（c.config.CORS == nil）或 Enabled=false 时：
//     - 开发模式（gin.DebugMode）：默认允许所有 Origin、关闭 Credentials，便于联调
//     - 发布模式（gin.ReleaseMode）：默认不注册 CORS 中间件，由调用方显式配置
//
// 三、配置示例
//
//	http:
//	  cors:
//	    enabled: true
//	    allow_origins:
//	      - "https://example.com"
//	      - "https://*.example.com"   # 需配合 allow_wildcard: true
//	      - "http://localhost:8080"   # 开发环境
//	    allow_methods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]
//	    allow_headers: ["Origin", "Content-Type", "Authorization", "X-Requested-With"]
//	    expose_headers: ["Content-Length", "X-Request-Id"]
//	    allow_credentials: true       # 携带 Cookie 时必须为 true，且 allow_origins 不可含 "*"
//	    allow_wildcard: true          # 启用子域名通配符匹配
//	    max_age: 7200                 # 预检缓存 2 小时（秒）
//
// 四、安全约束矩阵
//
//	+--------------------------+-------------------+-------------------+
//	| AllowOrigins             | AllowCredentials  | 是否合法          |
//	+--------------------------+-------------------+-------------------+
//	| ["*"]                    | false             | ✅ 公开 API       |
//	| ["*"]                    | true              | ❌ 规范禁止，降级 |
//	| ["https://example.com"]  | false             | ✅ 无凭证白名单   |
//	| ["https://example.com"]  | true              | ✅ 带凭证白名单   |
//	| ["https://*.example.com"]| true              | ✅ 子域名通配     |
//	| [] (空)                  | 任意              | ⚠  使用默认策略   |
//	+--------------------------+-------------------+-------------------+
//
// 五、调用链路
//
//	configs/server-win.yaml
//	    ↓ (viper 解析)
//	config.HttpServerConfig.CORS (*CORSConfig)
//	    ↓ (NewWebServerControl 注入)
//	http.Control.registerMiddlewares
//	    ↓ (校验 + 转换)
//	cors.EnableCors(cfg, logger) gin.HandlerFunc
//	    ↓ (委托)
//	gin-contrib/cors.New(libConfig)
//	    ↓ (运行时)
//	浏览器 CORS 协议交互
//
// 六、与项目其他中间件的关系
//
//   - 注册顺序：CORS 在 IP 黑白名单之后、限流之前注册（见 tls.go registerMiddlewares）
//     原因：先做 IP 层过滤，再处理跨域，避免恶意 IP 触发不必要的 CORS 计算。
//   - 与 JWT 中间件独立：CORS 仅处理跨域头，不涉及认证；JWT 中间件在 authGroup
//     子组中注册，与 CORS 不冲突。
//
// 七、扩展点
//
//	如需更复杂的 Origin 校验（如动态查询数据库白名单），可在 CORSConfig 中
//	扩展字段，并在 buildLibConfig 中转换为 gin-contrib/cors 的 AllowOriginFunc。
//	当前实现已覆盖静态白名单与子域名通配两种主流场景。
package cors

// 默认值常量
//
// @description 仅在配置未提供对应字段时作为兜底默认值使用，生产环境应显式配置
const (
	// defaultAllowMethods 默认允许的HTTP方法
	defaultAllowMethods = "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS"
	// defaultAllowHeaders 默认允许的请求头
	defaultAllowHeaders = "Origin, Content-Length, Content-Type, Authorization"
	// defaultMaxAgeSeconds 默认预检缓存时间（秒），2小时
	defaultMaxAgeSeconds = 7200
	// wildcardOrigin 通配符 Origin，表示允许所有 Origin（不可与 Credentials 同时使用）
	wildcardOrigin = "*"
)
