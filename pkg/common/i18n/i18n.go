package i18n

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/common/lang"
)

// 支持的语言
const (
	LanguageZh = lang.LanguageZh
	LanguageEn = lang.LanguageEn
)

// 默认语言
const defaultLanguage = lang.LanguageZh

// 翻译键值映射
var translations = map[string]map[string]string{
	LanguageZh: {
		// 通用消息
		"success":             "业务处理成功",
		"failed":              "业务处理失败",
		"invalid_parameter":   "请求参数无效，请检查参数",
		"unauthorized":        "未授权，请先登录",
		"forbidden":           "禁止访问，权限不足",
		"too_many_requests":   "请求过于频繁，请稍后再试",
		"internal_error":      "服务器内部错误，请稍后再试",
		"not_found":           "资源未找到",
		"method_not_allowed":  "请求方法不被允许",
		"service_unavailable": "服务不可用，请稍后再试",

		// 文件相关消息
		"file_not_found":    "文件未找到",
		"file_too_large":    "文件过大，请上传较小的文件",
		"file_type_error":   "文件类型错误，请上传正确的文件类型",
		"file_write_error":  "文件写入失败，请检查磁盘空间或权限",
		"file_read_error":   "文件读取失败，请检查文件是否损坏",
		"file_delete_error": "文件删除失败，请检查权限",

		// 数据库相关消息
		"database_error":            "数据库操作失败，请稍后再试",
		"record_not_found":          "记录未找到",
		"duplicate_record":          "记录已存在",
		"database_timeout":          "数据库操作超时，请稍后再试",
		"database_connection_error": "数据库连接失败，请检查数据库配置",

		// 用户相关消息
		"user_not_found":          "用户不存在",
		"invalid_password":        "密码错误",
		"user_disabled":           "用户已被禁用",
		"username_already_exists": "用户名已存在",
		"email_already_exists":    "邮箱已存在",
		"invalid_token":           "无效的 Token",
		"token_expired":           "Token 已过期，请重新登录",
		"password_reset_failed":   "密码重置失败，请稍后再试",
		"authentication_failed":   "认证失败，请检查用户名和密码",
		"permission_denied":       "权限不足，无法访问该资源",
		"invalid_credentials":     "无效的凭证",
		"session_expired":         "会话已过期，请重新登录",

		// 网络相关消息
		"network_error":      "网络错误，请检查网络连接",
		"timeout_error":      "请求超时，请稍后再试",
		"connection_refused": "连接被拒绝，请检查服务是否启动",

		// 业务逻辑消息
		"business_logic_error": "业务逻辑错误",
		"invalid_input":        "无效的输入",
		"insufficient_funds":   "余额不足",
		"resource_unavailable": "资源不可用",
		"data_conflict":        "数据冲突，请稍后再试",

		// 验证码相关消息
		"invalid_captcha":           "验证码不正确",
		"captcha_expired":           "验证码已过期，请刷新",
		"captcha_generation_failed": "验证码生成失败，请稍后再试",

		// 第三方服务消息
		"third_party_service_error": "第三方服务错误，请稍后再试",
		"third_party_api_failed":    "调用第三方 API 失败，请检查配置",
	},
	LanguageEn: {
		// 通用消息
		"success":             "Operation successful",
		"failed":              "Operation failed",
		"invalid_parameter":   "Invalid request parameters, please check",
		"unauthorized":        "Unauthorized, please login first",
		"forbidden":           "Access forbidden, insufficient permissions",
		"too_many_requests":   "Too many requests, please try again later",
		"internal_error":      "Internal server error, please try again later",
		"not_found":           "Resource not found",
		"method_not_allowed":  "Request method not allowed",
		"service_unavailable": "Service unavailable, please try again later",

		// 文件相关消息
		"file_not_found":    "File not found",
		"file_too_large":    "File too large, please upload a smaller file",
		"file_type_error":   "File type error, please upload the correct file type",
		"file_write_error":  "File write failed, please check disk space or permissions",
		"file_read_error":   "File read failed, please check if the file is corrupted",
		"file_delete_error": "File delete failed, please check permissions",

		// 数据库相关消息
		"database_error":            "Database operation failed, please try again later",
		"record_not_found":          "Record not found",
		"duplicate_record":          "Record already exists",
		"database_timeout":          "Database operation timeout, please try again later",
		"database_connection_error": "Database connection failed, please check database configuration",

		// 用户相关消息
		"user_not_found":          "User not found",
		"invalid_password":        "Invalid password",
		"user_disabled":           "User has been disabled",
		"username_already_exists": "Username already exists",
		"email_already_exists":    "Email already exists",
		"invalid_token":           "Invalid token",
		"token_expired":           "Token has expired, please login again",
		"password_reset_failed":   "Password reset failed, please try again later",
		"authentication_failed":   "Authentication failed, please check username and password",
		"permission_denied":       "Insufficient permissions to access this resource",
		"invalid_credentials":     "Invalid credentials",
		"session_expired":         "Session has expired, please login again",

		// 网络相关消息
		"network_error":      "Network error, please check your connection",
		"timeout_error":      "Request timeout, please try again later",
		"connection_refused": "Connection refused, please check if the service is running",

		// 业务逻辑消息
		"business_logic_error": "Business logic error",
		"invalid_input":        "Invalid input",
		"insufficient_funds":   "Insufficient funds",
		"resource_unavailable": "Resource unavailable",
		"data_conflict":        "Data conflict, please try again later",

		// 验证码相关消息
		"invalid_captcha":           "Invalid captcha",
		"captcha_expired":           "Captcha has expired, please refresh",
		"captcha_generation_failed": "Captcha generation failed, please try again later",

		// 第三方服务消息
		"third_party_service_error": "Third party service error, please try again later",
		"third_party_api_failed":    "Third party API call failed, please check configuration",
	},
}

// Translate 翻译消息
func Translate(lang, key string) string {
	// 获取指定语言的翻译映射
	langMap, exists := translations[lang]
	if !exists {
		// 如果语言不存在，使用默认语言
		langMap = translations[defaultLanguage]
	}

	// 获取翻译文本
	text, exists := langMap[key]
	if !exists {
		// 如果键不存在，返回键本身
		return key
	}

	return text
}

// TranslateWithDefaultLanguage 使用默认语言翻译消息
func TranslateWithDefaultLanguage(key string) string {
	return Translate(defaultLanguage, key)
}

// TranslateAuto 从上下文中自动获取语言并翻译消息
func TranslateAuto(ctx *gin.Context, key string) string {
	return Translate(lang.GetLanguageFromContext(ctx), key)
}
