package ipwhitelist

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnableIPWhiteList 创建并返回IP白名单中间件
// @param logger 日志记录器
// @param whiteList IP白名单列表
// @return gin.HandlerFunc Gin中间件函数
func EnableIPWhiteList(logger *zap.SugaredLogger, whiteList []string) gin.HandlerFunc {
	// 创建白名单集合，提高查找效率
	whiteListMap := make(map[string]bool)
	for _, ip := range whiteList {
		if ip != "" {
			whiteListMap[strings.TrimSpace(ip)] = true
		}
	}

	return func(c *gin.Context) {
		// 如果白名单为空，则不进行过滤
		if len(whiteListMap) == 0 {
			c.Next()
			return
		}

		// 获取客户端IP
		clientIP := getClientIP(c)

		// 检查IP是否在白名单中
		if whiteListMap[clientIP] {
			logger.Debugf("[IP WhiteList] Allowed IP: %s, Path: %s", clientIP, c.Request.URL.Path)
			c.Next()
		} else {
			logger.Warnf("[IP WhiteList] Forbidden IP: %s, Path: %s", clientIP, c.Request.URL.Path)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Access forbidden: IP not in white list",
				"data":    nil,
			})
			c.Abort()
		}
	}
}

// getClientIP 获取客户端真实IP地址
// 优先从X-Forwarded-For头获取，其次是X-Real-IP，最后是RemoteAddr
func getClientIP(c *gin.Context) string {
	// 从X-Forwarded-For头获取IP，通常由代理服务器添加
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For格式可能是多个IP，逗号分隔，第一个是原始客户端IP
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 从X-Real-IP头获取，通常由Nginx等代理服务器设置
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 直接从连接中获取RemoteAddr
	remoteAddr := c.Request.RemoteAddr
	// 移除端口部分
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		remoteAddr = remoteAddr[:idx]
	}
	// 移除可能的[]括号（IPv6地址）
	remoteAddr = strings.TrimPrefix(strings.TrimSuffix(remoteAddr, "]"), "[")

	return remoteAddr
}
