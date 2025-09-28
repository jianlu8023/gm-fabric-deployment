# OTEL 链路追踪使用指南

## 概述

本文档介绍如何在项目中使用修改后的OTEL控制器实现全链路追踪功能，特别是针对HTTP请求的追踪。修改后的控制器支持在整个调用链中传递上下文(context)，从而实现完整的链路追踪。

## 修改内容概述

1. **Span方法修改**：现在支持传入外部context作为父上下文
2. **NewOTELControl方法增强**：添加了对nil上下文的处理，默认使用context.Background()
3. **提供HTTP中间件示例**：展示如何在HTTP请求中集成链路追踪

## 控制器初始化

### 初始化OTEL控制器

```go
import (
	"context"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/otel"
)

// 初始化日志控制器
loggerControl, err := logger.NewLoggerControl(loggerConfig)
if err != nil {
	// 处理错误
}

// 初始化OTEL配置
otelConfig := &config.OTELConfig{
	Enabled:             true,
	ExporterServiceName: "my-service",
	TracesExporter:      "stdout", // 可选: stdout, file, otlp, zipkin
}

// 初始化OTEL控制器
// 可以传入nil作为ctx参数，控制器会自动使用context.Background()
otelControl, err := otel.NewOTELControl(otelConfig, loggerControl, nil)
if err != nil {
	// 处理错误
}
```

## HTTP请求追踪

### 使用中间件集成追踪

项目中已提供了OTEL中间件示例，位于`pkg/control/http/otel_middleware.go`。您可以按照以下方式在HTTP服务器中集成此中间件：

```go
import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/http"
)

// 创建Gin引擎
ginRouter := gin.New()

// 添加OTEL中间件（注意：应在其他中间件之前添加，以确保完整追踪）
ginRouter.Use(http.OTELMiddleware(otelControl, logger))

// 注册路由和处理函数
// ...
```

### 在处理函数中使用追踪

中间件会自动为每个HTTP请求创建一个根span，并将包含追踪信息的上下文传递给处理函数。您可以在处理函数中获取这个上下文并用于创建子span：

```go
func MyHandler(c *gin.Context) {
	// 从请求中获取上下文
	ctx := c.Request.Context()

	// 使用上下文创建子span
	ctx, span := otelControl.Span(ctx, "handler", "MyHandler")
	defer span.End()

	// 执行业务逻辑
	// ...

	// 调用服务层时传递上下文
	result, err := service.DoSomething(ctx, params)
	// ...
}
```

## 服务层追踪

在服务层中，应该接收并使用传入的上下文创建新的span，以确保链路的连续性：

```go
func (s *Service) DoSomething(ctx context.Context, params Params) (Result, error) {
	// 创建子span，使用传入的上下文
	ctx, span := s.otelControl.Span(ctx, "service", "DoSomething")
	defer span.End()

	// 添加属性（可选）
	span.SetAttributes(
		attribute.String("param.id", params.ID),
		attribute.Int64("param.count", params.Count),
	)

	// 执行业务逻辑
	// ...

	// 调用其他服务时继续传递上下文
	data, err := s.repository.GetData(ctx, params.ID)
	if err != nil {
		// 记录错误
		span.RecordError(err)
		return Result{}, err
	}

	// 返回结果
	return Result{Data: data}, nil
}
```

## 最佳实践

1. **始终传递上下文**：在函数调用链中始终传递上下文，确保追踪链路的完整性

2. **正确命名span**：使用有意义的组件名和span名，格式为`componentName.spanName`

3. **使用defer确保span正确结束**：始终使用`defer span.End()`确保span被正确结束

4. **记录关键属性**：为span添加有用的属性，如请求参数、用户ID等

5. **记录错误信息**：使用`span.RecordError(err)`记录错误信息

6. **中间件顺序**：OTEL中间件应在其他中间件之前注册，以确保完整追踪请求处理过程

## 常见问题与解决方案

### 问题：追踪信息未在整个链路中传递

**解决方案**：检查是否在所有函数调用中正确传递了上下文。确保每个函数都接收并使用传入的上下文创建新的span。

### 问题：某些请求没有生成追踪信息

**解决方案**：检查OTEL配置是否正确，确保`Enabled`设置为`true`。另外，检查中间件是否正确注册。

### 问题：追踪信息过多或过少

**解决方案**：调整`TracesExporter`配置，选择合适的导出器（stdout、file、otlp或zipkin）。对于生产环境，建议使用otlp或zipkin。

## 代码优化建议

1. **使用接口抽象**：考虑为OTEL控制器定义接口，以便于测试和替换实现

2. **统一错误处理**：创建统一的错误处理函数，自动记录错误到span

3. **添加采样策略**：根据实际需求配置采样策略，避免生成过多的追踪数据

4. **集成日志**：将追踪ID集成到日志中，便于关联日志和追踪数据

5. **监控与告警**：设置基于追踪数据的监控和告警，及时发现性能瓶颈和错误

## 示例代码

完整的示例代码可以在以下文件中找到：

- OTEL控制器实现：`pkg/control/otel/control.go`
- HTTP中间件示例：`pkg/control/http/otel_middleware.go`

请参考这些文件了解更多实现细节和使用方法。