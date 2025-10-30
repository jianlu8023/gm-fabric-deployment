package lang

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// 支持的语言
const (
	LanguageZh = "zh"
	LanguageEn = "en"
)

// 默认语言
const defaultLanguage = LanguageZh

// ContextLanguageKey 语言上下文键
const ContextLanguageKey = "language"

// GetLanguageFromContext 从Gin上下文中获取语言设置
func GetLanguageFromContext(ctx *gin.Context) string {
	// 首先检查URL参数
	lang := ctx.Query("lang")
	if !stringer.IsBlank(lang) {
		return normalizeLanguage(lang)
	}

	// 然后检查Accept-Language头
	acceptLang := ctx.GetHeader("Accept-Language")
	if !stringer.IsBlank(acceptLang) {
		// 简单解析Accept-Language头，取第一个语言
		parts := strings.Split(acceptLang, ",")
		if len(parts) > 0 {
			lang = strings.TrimSpace(parts[0])
			return normalizeLanguage(lang)
		}
	}

	// 最后返回默认语言
	return defaultLanguage
}

// GetLanguageFromContextOrDefault 从Gin上下文中获取语言设置，如果未找到则返回默认语言
func GetLanguageFromContextOrDefault(ctx *gin.Context, defaultLang string) string {
	// 首先检查URL参数
	lang := ctx.Query("lang")
	if !stringer.IsBlank(lang) {
		return normalizeLanguage(lang)
	}

	// 然后检查Accept-Language头
	acceptLang := ctx.GetHeader("Accept-Language")
	if !stringer.IsBlank(acceptLang) {
		// 简单解析Accept-Language头，取第一个语言
		parts := strings.Split(acceptLang, ",")
		if len(parts) > 0 {
			lang = strings.TrimSpace(parts[0])
			return normalizeLanguage(lang)
		}
	}

	// 最后返回默认语言
	return defaultLang
}

// normalizeLanguage 标准化语言标识符
func normalizeLanguage(lang string) string {
	// 转换为小写
	lang = strings.ToLower(lang)

	// 处理带地区标识的语言代码，如zh-CN, en-US等
	if strings.Contains(lang, "-") {
		parts := strings.Split(lang, "-")
		lang = parts[0]
	}

	// 验证是否为支持的语言 (这里我们只检查是否为已知的语言代码)
	switch lang {
	case LanguageZh, LanguageEn:
		return lang
	default:
		// 不支持的语言返回默认语言
		return defaultLanguage
	}
}
