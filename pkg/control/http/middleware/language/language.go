package language

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/common/lang"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
)

// ContextLanguageKey 语言上下文键
const ContextLanguageKey = lang.ContextLanguageKey

// EnableLanguageSupport 启用语言支持中间件
func EnableLanguageSupport(defaultLang string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "language")
		defer span.End()

		// 从请求中获取语言设置
		langCode := lang.GetLanguageFromContextOrDefault(ctx, defaultLang)

		// 将语言设置存储在上下文中
		ctx.Set(ContextLanguageKey, langCode)

		// 继续处理请求
		ctx.Next()

		span.SetStatus(codes.Ok, "success")
	}
}

// GetLanguageFromContext 从Gin上下文中获取语言设置
func GetLanguageFromContext(ctx *gin.Context) string {
	if langCode, exists := ctx.Get(ContextLanguageKey); exists {
		if langStr, ok := langCode.(string); ok {
			return langStr
		}
	}

	// 如果上下文中没有语言设置，直接从请求中获取
	return lang.GetLanguageFromContext(ctx)
}
