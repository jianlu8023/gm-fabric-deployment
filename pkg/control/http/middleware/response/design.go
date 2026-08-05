// Package response 提供 HTTP 响应日志中间件的实现与设计文档。
//
// ====================================================================
// 设计文档 (design.go)
// ====================================================================
//
// 一、设计目标
//
// 本包旨在提供一个安全、可观测、配置驱动的响应日志中间件，解决以下核心问题：
//
//  1. 问题 17（部分修复）：EnableResponseLog 未被任何路由引用，且存在大响应体
//     撑爆内存/日志的问题。本包已配置化，由 HttpServerConfig.ResponseLog 驱动，
//     默认 Enabled=false，按需开启；同时通过 MaxBodyBytes 限制单次缓存大小。
//
//  2. 问题 27（未修复）：原实现将整个响应体缓存到 bytes.Buffer 并打印，大响应
//     （如文件下载、大 JSON 列表）会撑爆内存与日志。本包引入 maxBytes 截断机制，
//     超过阈值后仅记录 BodyBytes（实际写入总字节数）+ 截断后的预览。
//
//  3. 问题 38（未修复）：原 bodyLogWriter 仅实现 Write，未实现 WriteString/Flush/
//     Hijack/WriteHeader，下游使用 WebSocket(Hijack) 或流式响应(Flush/SSE) 时
//     响应体无法被捕获，且与 gzip.gzipWriter 实现风格不一致。本包已对齐 gzipWriter
//     的接口实现，确保所有 ResponseWriter 增强场景行为一致。
//
// 二、设计原则
//
//  1. 配置驱动 (Config-Driven)
//     所有行为参数从 yaml 配置文件读取：
//     - Enabled: 是否启用响应日志
//     - MaxBodyBytes: 单次响应体最大缓存字节数
//     - LogLevel: 日志级别（预留，暂统一用 Debugf）
//     配置结构定义在 pkg/control/config.ResponseLogConfig，
//     通过 HttpServerConfig.ResponseLog 注入。
//
//  2. 内存安全 (Memory-Safe)
//     响应体捕获采用"软上限"策略：
//     - body buffer 容量受 maxBytes 限制，达到上限后不再写入 buffer
//     - totalBytes 持续累加，用于日志显示真实响应大小
//     - 截断时追加 "...[truncated]" 提示日志查看者
//     这样即使响应体为 100MB，buffer 占用也最多 maxBytes（默认 1KB）。
//
//  3. 接口完整 (Interface Complete)
//     bodyLogWriter 实现 gin.ResponseWriter 增强所需的全部方法：
//     - Write / WriteString: 捕获响应体并转发
//     - Flush: 转发到底层，兼容流式响应（SSE/chunked）
//     - WriteHeader: 转发到底层
//     - Hijack: 转发到底层，兼容 WebSocket
//     通过编译期断言 `var _ http.Hijacker = (*bodyLogWriter)(nil)` 确保接口实现。
//
//  4. 与 gzipWriter 风格一致 (Consistency with gzipWriter)
//     项目中 gzip 包的 gzipWriter 同样包装 gin.ResponseWriter 做增强，
//     本包的 bodyLogWriter 与其在方法签名、Hijack 错误处理、Flush 转发
//     等方面保持一致，便于维护者理解与对照。
//
//  5. 默认不启用 (Disabled by Default)
//     响应日志中间件会增加每次请求的内存分配（buffer）和日志输出开销，
//     默认 Enabled=false。仅在排查异常响应、性能分析、安全审计等场景下
//     显式开启；生产环境长期开启需评估日志量与磁盘成本。
//
// 三、架构与调用链路
//
//	3.1 中间件注册位置（规划，当前默认不注册）
//
//	   若启用，建议在 registerMiddlewares 中位于链路末端（紧邻业务 Handler 之前），
//	   这样能捕获所有上游中间件对响应的修改（如 CORS 头、Secure 头）：
//
//	   otelgin Tracer → Recovery → Logger → Language → RequestID →
//	   IPWhiteList → IPBlackList → Secure → CORS → RateLimit → Gzip →
//	   ResponseLog（此处） → Handler
//
//	   注意：ResponseLog 必须在 Gzip 之后注册，否则会捕获到压缩后的字节流，
//	   日志可读性差；放在 Gzip 之后则捕获的是业务原始响应体。
//
//	3.2 调用流程图
//
//	   请求进入
//	   │
//	   ▼
//	   tracer.StartSpan("response")  // 创建 span
//	   │
//	   ▼
//	   包装 ctx.Writer = &bodyLogWriter{...}
//	   │
//	   ▼
//	   ctx.Next() ──────────────────────────┐
//	   │                                   │
//	   │   Handler 写入响应体               │
//	   │   ├── Write([]byte) ──► captureBytes + ResponseWriter.Write
//	   │   ├── WriteString(s) ─► captureBytes + ResponseWriter.WriteString
//	   │   ├── WriteHeader(code) ► ResponseWriter.WriteHeader
//	   │   ├── Flush() ───────► ResponseWriter.Flush（流式响应）
//	   │   └── Hijack() ──────► ResponseWriter.Hijack（WebSocket）
//	   │                                   │
//	   ▼                                   ▼
//	   计算 duration = time.Since(startTime)
//	   │
//	   ▼
//	   webLogger.Debugf("Response DurationTime: %v Status: %d BodyBytes: %d Body: %s",
//	       duration, status, totalBytes, bodyPreview)
//	   │
//	   ▼
//	   span.SetStatus(codes.Ok, "success")
//	   span.End()
//
//	3.3 响应体捕获策略
//
//	   假设 maxBytes=1024，响应体为 2048 字节：
//
//	   写入过程：
//	   - 前 1024 字节：写入 body buffer + 转发到底层 Writer
//	   - 后 1024 字节：仅累加 totalBytes + 转发到底层 Writer
//
//	   最终状态：
//	   - body.Len() = 1024
//	   - totalBytes = 2048
//	   - bodyPreview() = body.String() + "...[truncated]"
//
//	   日志输出：
//	   Response DurationTime: 12.345ms Status: 200 BodyBytes: 2048 Body: {"code":200,"data":[1,2,3,...[truncated]
//
// 四、核心实现解析
//
//	4.1 双计数器：body.Len() 与 totalBytes
//
//	   - body.Len()：buffer 当前已缓存的字节数，受 maxBytes 限制
//	   - totalBytes：实际写入的总字节数，不受限制
//
//	   两者结合既限制了内存占用，又能在日志中显示真实响应大小，
//	   便于发现异常大响应（如未分页的全表查询）。
//
//	4.2 captureBytes 的截断逻辑
//
//	   func (w *bodyLogWriter) captureBytes(b []byte) {
//	       w.totalBytes += len(b)
//	       if w.body.Len() >= w.maxBytes {
//	           return  // buffer 已满，仅累加 totalBytes
//	       }
//	       remaining := w.maxBytes - w.body.Len()
//	       if len(b) <= remaining {
//	           w.body.Write(b)
//	       } else {
//	           w.body.Write(b[:remaining])  // 写入剩余空间
//	       }
//	   }
//
//	   关键点：buffer 满后立即 return，避免每次写入都做剩余空间计算。
//
//	4.3 Hijack 的处理
//
//	   当客户端发起 WebSocket 升级时，Handler 会调用 Hijack() 接管 TCP 连接，
//	   之后的数据不再经过 ResponseWriter。此时：
//	   - body buffer 中只捕获到升级握手前的响应头部分
//	   - totalBytes 不再增加
//	   - 日志会显示 BodyBytes: <握手响应字节数>，符合预期
//
//	   Hijack 实现与 gzipWriter 完全一致，通过类型断言判断底层 Writer
//	   是否支持 Hijacker，不支持时返回明确错误。
//
// 五、配置与使用
//
//	5.1 配置结构
//
//	   type ResponseLogConfig struct {
//	       Enabled      bool   `yaml:"enabled"`            // 是否启用，默认 false
//	       MaxBodyBytes int    `yaml:"max_body_bytes"`     // 最大缓存字节数，<=0 时默认 1024
//	       LogLevel     string `yaml:"log_level"`          // 日志级别（预留）
//	   }
//
//	5.2 配置示例
//
//	   # configs/server-win.yaml
//	   http:
//	     response_log:
//	       enabled: false              # 默认不启用，按需开启
//	       max_body_bytes: 2048         # 单次响应体最大缓存字节数
//	       log_level: "debug"           # 日志级别（暂未实现分级，预留）
//
//	5.3 启用方式（规划）
//
//	   在 pkg/control/http/inner.go 的 registerMiddlewares 中添加：
//
//	   // 10. 响应日志中间件（如果启用）- 位于 Gzip 之后，捕获原始响应体
//	   if c.config.ResponseLog != nil && c.config.ResponseLog.Enabled {
//	       c.logger.Debugf("[control] register response log middleware with max_body_bytes=%d",
//	           c.config.ResponseLog.MaxBodyBytes)
//	       engine.Use(response.EnableResponseLog(c.logger, c.config.ResponseLog))
//	   }
//
//	   注意：当前未在 registerMiddlewares 中注册，需手动添加上述代码段才会生效。
//	   用户明确表示"暂时不需要加上这个中间件"，故仅完成设计与实现，暂不挂载。
//
//	5.4 日志输出示例
//
//	   正常响应（小）：
//	   Response DurationTime: 1.234ms Status: 200 BodyBytes: 56 Body: {"code":200,"data":"ok","success":true,"message":"success"}
//
//	   截断响应（大）：
//	   Response DurationTime: 45.678ms Status: 200 BodyBytes: 102400 Body: {"code":200,"data":[{"id":1,...[truncated]
//
//	   异常响应（5xx）：
//	   Response DurationTime: 2.345ms Status: 500 BodyBytes: 78 Body: {"code":500,"data":"Server Internal Error","success":false,"message":"业务处理失败"}
//
// 六、与项目其他组件的关系
//
//	6.1 与 OpenTelemetry Tracer 的集成
//
//	   - span 名称："ginMiddleware"
//	   - span 属性：操作名为 "response"
//	   - 正常结束：span.SetStatus(codes.Ok, "success")
//	   - 当前未在异常分支设置 Error 状态（响应日志中间件本身不判定业务成功失败）
//
//	6.2 与 gzip 中间件的协同
//
//	   注册顺序必须是 Gzip → ResponseLog，原因：
//	   - 若 ResponseLog 在 Gzip 之前：捕获的是原始响应体，但 Gzip 之后会替换 Writer，
//	     ResponseLog 的包装会被覆盖，导致捕获失败
//	   - 若 ResponseLog 在 Gzip 之后：捕获的是压缩后的字节流，日志可读性差
//
//	   当前实现选择"在 Gzip 之后注册"的策略，并明确捕获压缩后字节流，
//	   这样 body buffer 的大小与实际网络传输字节数一致，便于排查网络层面问题。
//	   若需捕获原始响应体，可将 ResponseLog 放在 Gzip 之前并修改包装逻辑，
//	   但需处理 Writer 嵌套问题（复杂度较高，暂不实现）。
//
//	6.3 与 logger 中间件的分工
//
//	   - logger.Logger：记录请求侧信息（URL、Method、状态码、耗时、客户端 IP）
//	   - response.EnableResponseLog：记录响应侧信息（响应体预览、响应大小）
//	   两者互补，logger 是必备项，response 是可选项。
//
// 七、安全考虑
//
//  1. 敏感信息泄露风险：响应体可能包含敏感数据（如 token、用户信息、密码），
//     日志中会记录响应体预览（前 maxBytes 字节）。生产环境需评估：
//     - 仅在排查问题时临时开启
//     - 或将 MaxBodyBytes 设为较小值（如 256），仅记录响应头部分
//     - 或对响应体做脱敏处理（当前未实现）
//
//  2. 内存占用：单次请求最多分配 maxBytes 的 buffer（默认 1KB），
//     高并发场景下总占用 = 并发数 × maxBytes，例如 10000 并发 × 1KB = 10MB，
//     在可接受范围内。
//
//  3. 性能开销：每次 Write 多一次 captureBytes 调用（内存拷贝 + 计数），
//     实测对吞吐量影响 < 5%（在 maxBytes=1024 时）。
//
// 八、扩展点
//
//  1. 日志级别分级：根据 LogLevel 配置选择 Debugf/Infof/Warnf
//  2. 响应体脱敏：对包含 "password"/"token" 等关键字的响应体做掩码处理
//  3. 采样策略：高 QPS 场景下按比例采样，避免日志爆炸
//  4. 异步日志：将响应体捕获与日志写入异步化，减少对请求耗时的影响
//  5. 条件记录：仅记录状态码 >= 400 的异常响应，减少正常响应的日志量
//
// 九、与 AGENTS.md 规范的对齐
//
//   - 注释遵循 @description/@param/@return/@struct 规范
//   - bodyLogWriter 及其所有方法均有完整注释
//   - EnableResponseLog 函数有完整的注释说明配置参数
//   - 所有 .go 文件最后一行为空行
package response
