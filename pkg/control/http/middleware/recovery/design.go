// Package recovery 提供 HTTP 请求 panic 恢复中间件的实现与设计文档。
//
// ====================================================================
// 设计文档 (design.go)
// ====================================================================
//
// 一、设计目标
//
// 本包旨在提供一个健壮、可观测的 panic 恢复中间件，解决以下核心问题：
//
//  1. 服务稳定性：业务代码或下游中间件的 panic 不应导致整个进程崩溃，
//     recovery 中间件位于中间件链最前端，确保任何后续环节的 panic 都能被捕获。
//  2. 可观测性：panic 发生时需完整记录请求上下文（URL、Method、Headers）、
//     错误信息、可选的 goroutine 堆栈，并通过 OpenTelemetry span 上报错误状态。
//  3. 安全降级：对 broken pipe、connection reset by peer 等常见网络异常
//     做特殊处理——这些属于客户端主动断开，不应产生堆栈日志或误报为系统故障。
//  4. 安全类型转换：recover() 返回 interface{}，直接 err.(error) 断言会导致
//     二次 panic（见历史问题 8），需通过 safeError() 统一安全转换。
//
// 历史问题修复：
//   - 问题 8：err.(error) 类型断言可能 panic → 引入 safeError() 使用 type switch 安全转换
//   - 问题 24：recovery.go 内大量注释掉的日志代码块 → 已清理
//
// 二、设计原则
//
//  1. 最小侵入 (Minimal Intrusion)
//     recovery 中间件只做三件事：捕获 panic、记录日志、返回 500 响应。
//     不引入业务逻辑，不修改请求/响应体，保持中间件的单一职责。
//
//  2. 可观测性优先 (Observability First)
//     每次请求（包括 panic 恢复）都在 OpenTelemetry 中创建 span：
//     - 正常请求：span 状态为 Ok
//     - panic 恢复：span 状态为 Error，标记为 "recovery" 操作
//     日志格式统一为结构化日志（SugaredLogger），包含完整的请求 URL、
//     dump 后的 HTTP 请求、错误信息，便于在 ELK/Loki 等系统中检索。
//
//  3. 环境适配 (Environment-Adaptive)
//     通过 stack 参数控制是否输出 goroutine 堆栈：
//     - 开发环境 (stack=true)：输出完整堆栈，便于定位问题
//     - 生产环境 (stack=false)：仅记录错误信息和请求摘要，避免日志膨胀
//     堆栈信息包含大量内存地址和 goroutine 状态，生产环境仅在必要时开启。
//
//  4. 安全转换 (Safe Conversion)
//     recover() 返回值为 interface{}，可能是任何类型：
//     - string（如 panic("something")）
//     - error（如 panic(fmt.Errorf("..."))）
//     - 自定义结构体（如 panic(MyError{Code: 500})）
//     - 基本类型（如 panic(42)）
//     safeError() 使用 type switch，已实现 error 接口的直接返回，
//     其余类型用 fmt.Errorf("%v", err) 安全包装，确保不会二次 panic。
//
//  5. 连接异常降级 (Connection Exception Downgrade)
//     检测 broken pipe / connection reset by peer：
//     - 这两种异常由客户端断开连接导致，服务端无法写入响应
//     - 处理方式：记录简要日志、ctx.Abort() 终止请求、标记 span Error
//     - 不返回 JSON 响应体（连接已断开，写入会失败并产生新的 panic）
//     - 不输出 goroutine 堆栈（这不是代码 bug）
//
//     三、架构与调用链路
//
//     3.1 中间件注册位置
//
//     在 http.Control.registerMiddlewares 中，recovery 是第二个注册的中间件
//     （紧接在 otelgin Tracer 之后），确保所有后续中间件（logger、security、
//     CORS、限流、业务 Handler）的 panic 都能被捕获。
//
//     注册顺序：
//     otelgin Tracer → Recovery → Logger → Language → RequestID →
//     IPWhiteList → IPBlackList → Secure → CORS → RateLimit → Gzip → Handler
//
//     3.2 调用流程图
//
//     请求进入
//     │
//     ▼
//     otelgin.Middleware (创建根 span)
//     │
//     ▼
//     Recovery.Middleware (创建 recovery span)
//     │
//     ├── defer recover() 注册
//     ├── 创建子 span，设置到 ctx.Request
//     │
//     ▼
//     ctx.Next() ─────────────────────────┐
//     │                                 │
//     │   panic? ──────────────────► defer 触发
//     │                                 │
//     │                           ┌─────┴─────┐
//     │                           │ 是 broken? │
//     │                           └─────┬─────┘
//     │                           ┌─────┴─────┐
//     │                           Y           N
//     │                     记录简要日志   记录完整日志
//     │                     ctx.Error()   ctx.JSON(500)
//     │                     ctx.Abort()   ctx.AbortWithStatus(500)
//     │                     span.Error    span.Error
//     │                     return        │
//     │                                    │
//     ▼                                    ▼
//     span.SetStatus(Ok)               span.End()
//     span.End()
//
//     3.3 请求上下文处理
//
//     recovery 中间件需要在 panic 恢复后保持原始的 context.Context 链：
//
//     savedCtx := ctx.Request.Context()   // 保存原始 context
//     // ... 创建子 span ...
//     ctx.Request = ctx.Request.WithContext(tCtx)
//     defer func() {
//     ctx.Request = ctx.Request.WithContext(savedCtx)  // 恢复原始 context
//     }()
//
//     这样确保 panic 恢复后，下游依赖 context 的逻辑（如日志、tracer）
//     不会因为 context 丢失而崩溃。
//
//     四、核心实现解析
//
//     4.1 双重 defer 模式
//
//     recovery 中间件使用两个 defer：
//     - 第一个 defer：恢复 context 链，确保即使 panic 也能正确还原
//     - 第二个 defer：recover() panic，执行日志记录和响应返回
//
//     执行顺序：
//
//  1. ctx.Next() 正常执行或触发 panic
//
//  2. 第二个 defer（recover）先执行——捕获 panic、记录日志、返回 500
//
//  3. 第一个 defer（context 恢复）后执行——恢复原始 context
//
//  4. span.End() 被触发（因为 panic 已被恢复）
//
//     4.2 broken pipe 检测
//
//     broken pipe 检测链：
//     recover() → *net.OpError → *os.SyscallError → 检查 error 字符串
//
//     判断条件（任一满足即视为 broken pipe）：
//     - se.Error() 包含 "broken pipe"（不区分大小写）
//     - se.Error() 包含 "connection reset by peer"（不区分大小写）
//
//     处理策略：
//     - 仅记录简要错误日志，不记录堆栈
//     - 调用 ctx.Abort() 终止请求链（连接已断开）
//     - 不调用 ctx.JSON()（写入会失败）
//
//     4.3 安全类型转换 safeError()
//
//     历史上使用 err.(error) 直接断言，当 panic 传入非 error 类型时
//     （如 panic("error message")），断言本身会 panic，导致原始 panic
//     信息丢失且无法恢复。
//
//     safeError() 实现：
//
//     func safeError(err interface{}) error {
//     switch v := err.(type) {
//     case error:
//     return v
//     default:
//     return fmt.Errorf("%v", err)
//     }
//     }
//
//     使用 type switch 而非 comma-ok 断言的优势：
//     - 更清晰的控制流
//     - 自然处理 default 分支
//     - Go 编译器对 type switch 有更好的优化
//
//     五、配置与扩展性
//
//     5.1 当前配置
//
//     本包采用配置驱动设计，通过 RecoveryConfig 结构体注入配置：
//
//     type RecoveryConfig struct {
//     Enabled       bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`
//     EnableStack   bool   `json:"enable_stack,omitempty" yaml:"enable_stack,omitempty" mapstructure:"enable_stack"`
//     LogLevel      string `json:"log_level,omitempty" yaml:"log_level,omitempty" mapstructure:"log_level"`
//     CustomMessage string `json:"custom_message,omitempty" yaml:"custom_message,omitempty" mapstructure:"custom_message"`
//     }
//
//     EnableRecovery(webLogger *zap.SugaredLogger, cfg *config.RecoveryConfig)
//
//     配置字段说明：
//     - Enabled: 是否启用recovery中间件，建议始终为true
//     - EnableStack: 是否在日志中输出goroutine堆栈，开发环境true，生产环境false
//     - LogLevel: panic日志级别（error | warn | fatal），暂未使用，预留扩展
//     - CustomMessage: 自定义响应消息，为空时使用默认"Server Internal Error"
//
//     5.2 配置示例
//
//     // configs/server-win.yaml
//     http:
//     recovery:
//     enabled: true                  # 是否启用recovery中间件，建议始终为true
//     enable_stack: true             # 开发环境开启，生产环境关闭避免日志膨胀
//     log_level: "error"             # 日志级别
//     custom_message: ""             # 自定义响应消息，为空时使用默认值
//
//     5.3 扩展点
//
//     以下扩展可在不破坏现有接口的前提下添加：
//
//  1. 自定义错误响应：允许传入自定义响应结构体，替换默认的
//     commonhttp.BaseResponse
//
//  2. 错误通知钩子：panic 发生时可触发告警（钉钉/邮件/短信），
//     实现在线故障快速响应
//
//  3. 分级堆栈：根据 panic 类型决定是否输出堆栈（如业务 panic 输出、
//     系统 panic 不输出）
//
//  4. 采样策略：高 QPS 场景下，相同错误可做日志采样，避免日志爆炸
//
//  5. 请求体记录：可选记录请求 body（需限制大小），辅助 panic 复现
//
//     六、与项目其他组件的关系
//
//     6.1 与 OpenTelemetry Tracer 的集成
//
//     recovery 中间件内部通过 tracer.StartSpan 创建子 span：
//     - span 名称："ginMiddleware"
//     - span 属性：操作名为 "recovery"
//     - 正常结束：span.SetStatus(codes.Ok, "success")
//     - panic 恢复：span.SetStatus(codes.Error, "")
//
//     这样在 Jaeger/Zipkin 等追踪系统中可以清晰看到 recovery 中间件的
//     执行状态，以及 panic 发生时的错误标记。
//
//     6.2 与 commonhttp 的集成
//
//     panic 恢复后的响应使用项目统一的 commonhttp.BaseResponse 封装：
//     {
//     "code": 500,
//     "message": "业务处理失败",
//     "success": false,
//     "data": "Server Internal Error"
//     }
//
//     确保客户端收到统一格式的错误响应。
//
//     6.3 与日志系统的集成
//
//     使用 *zap.SugaredLogger 记录结构化日志：
//     - 日志级别：Error
//     - 日志内容：请求 URL + dump 后的 HTTP 请求 + 错误信息
//     - 可选：goroutine 堆栈（debug.Stack()）
//
//     七、安全考虑
//
//  1. 敏感信息泄露风险：httputil.DumpRequest 会 dump 请求头和路径，
//     可能包含 Authorization、Cookie 等敏感信息。当前未做脱敏处理，
//     生产环境需评估风险或在 DumpRequest 前对敏感头做掩码。
//
//  2. 堆栈信息泄露：debug.Stack() 输出的堆栈可能包含文件路径、
//     函数名等内部信息。生产环境建议关闭 stack 或在响应中不返回堆栈。
//
//  3. 错误消息定制：当前响应消息为硬编码的"业务处理失败"和
//     "Server Internal Error"，应根据环境调整（开发环境返回更多细节，
//     生产环境返回通用消息）。
//
//     八、与 AGENTS.md 规范的对齐
//
//     - 注释遵循 @description/@param/@return 规范
//     - safeError 函数有完整的注释说明设计思路
//     - EnableRecovery 函数有完整的注释
//     - 错误处理完善：safeError() 防止二次 panic
//     - 所有 .go 文件最后一行为空行（design.go 最后一行）
package recovery
