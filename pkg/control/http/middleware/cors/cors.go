package cors

import (
	"strings"
	"time"

	gincors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"go.uber.org/zap"
)

// EnableCors 启用CORS中间件
//
// @description 根据传入的 CORSConfig 构建安全的 gin-contrib/cors 中间件。
// 该函数会主动校验配置：若 AllowOrigins 含 "*" 且 AllowCredentials=true，
// 会记录 error 日志并强制将 AllowCredentials 降级为 false，避免运行时 panic
// 与浏览器拒绝带凭证的跨域请求。
// @param cfg *config.CORSConfig CORS配置，为 nil 时使用开发模式默认值
// @param logger *zap.SugaredLogger 日志实例，用于记录配置校验与降级信息
// @return gin.HandlerFunc Gin中间件函数
func EnableCors(cfg *config.CORSConfig, logger *zap.SugaredLogger) gin.HandlerFunc {
	libConfig := buildLibConfig(cfg, logger)
	return gincors.New(libConfig)
}

// buildLibConfig 将应用层 CORSConfig 转换为 gin-contrib/cors 的 Config，
// 并在转换过程中执行安全校验与默认值兜底。
//
// @description 转换规则见 design.go 中的"安全约束矩阵"；当配置缺失时按运行模式给出默认策略
// @param cfg *config.CORSConfig 应用层CORS配置
// @param logger *zap.SugaredLogger 日志实例
// @return gincors.Config gin-contrib/cors 库的配置
func buildLibConfig(cfg *config.CORSConfig, logger *zap.SugaredLogger) gincors.Config {
	// 处理 nil 或未启用：按运行模式给出默认策略
	if cfg == nil || !cfg.Enabled {
		return buildDefaultConfig()
	}

	// 校验 AllowOrigins 含 "*" 且 AllowCredentials=true 的违规组合
	allowOrigins := cfg.AllowOrigins
	allowCredentials := cfg.AllowCredentials
	if hasWildcardOrigin(allowOrigins) && allowCredentials {
		logger.Error("[http/cors] 配置违规：AllowOrigins 含 \"*\" 且 AllowCredentials=true，" +
			"违反 W3C CORS 规范，已强制将 AllowCredentials 降级为 false。" +
			"如需携带 Cookie，请改为显式 Origin 白名单或子域名通配符（allow_wildcard: true）")
		allowCredentials = false
	}

	// 若白名单为空，按运行模式兜底
	if len(allowOrigins) == 0 {
		if gin.Mode() == gin.DebugMode {
			logger.Warn("[http/cors] AllowOrigins 为空且处于开发模式，默认允许所有 Origin（不携带凭证）")
			allowOrigins = []string{wildcardOrigin}
		} else {
			logger.Warn("[http/cors] AllowOrigins 为空且处于发布模式，默认拒绝所有跨域请求；" +
				"请在配置文件中显式设置 allow_origins")
		}
	}

	libConfig := gincors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     defaultStringSlice(cfg.AllowMethods, splitCSV(defaultAllowMethods)),
		AllowHeaders:     defaultStringSlice(cfg.AllowHeaders, splitCSV(defaultAllowHeaders)),
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: allowCredentials,
		AllowWildcard:    cfg.AllowWildcard,
	}

	// 设置预检缓存时间
	if cfg.MaxAge > 0 {
		libConfig.MaxAge = time.Duration(cfg.MaxAge) * time.Second
	} else {
		libConfig.MaxAge = time.Duration(defaultMaxAgeSeconds) * time.Second
	}

	// 生产环境告警：允许所有 Origin + 关闭凭证是公开 API 模式，应确认是否符合预期
	if hasWildcardOrigin(allowOrigins) && gin.Mode() != gin.DebugMode {
		logger.Warn("[http/cors] 当前允许所有 Origin（*），仅适用于公开 API；" +
			"若需携带凭证请改为白名单模式")
	}

	return libConfig
}

// buildDefaultConfig 构造默认配置：开发模式允许所有 Origin 且不携带凭证
//
// @description 当 CORS 配置缺失或未启用时使用，确保开发环境可正常联调
// @return gincors.Config 默认CORS配置
func buildDefaultConfig() gincors.Config {
	return gincors.Config{
		AllowOrigins:     []string{wildcardOrigin},
		AllowMethods:     splitCSV(defaultAllowMethods),
		AllowHeaders:     splitCSV(defaultAllowHeaders),
		AllowCredentials: false,
		MaxAge:           time.Duration(defaultMaxAgeSeconds) * time.Second,
	}
}

// hasWildcardOrigin 判断 Origin 列表中是否包含通配符 "*"
//
// @param origins []string Origin 列表
// @return bool 含 "*" 返回 true
func hasWildcardOrigin(origins []string) bool {
	for _, o := range origins {
		if strings.TrimSpace(o) == wildcardOrigin {
			return true
		}
	}
	return false
}

// defaultStringSlice 当输入切片非空时返回输入，否则返回默认切片
//
// @param value []string 输入切片
// @param defaultValue []string 默认切片
// @return []string 最终结果
func defaultStringSlice(value, defaultValue []string) []string {
	if len(value) > 0 {
		return value
	}
	return defaultValue
}

// splitCSV 将逗号分隔的字符串切分为切片，元素已去除首尾空白
//
// @param csv 逗号分隔的字符串
// @return []string 切分后的切片
func splitCSV(csv string) []string {
	parts := strings.Split(csv, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
