package ipblacklist

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/iphelper"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// EnableIPBlackList 创建并返回IP黑名单中间件（支持精确IP和CIDR范围）
// @param logger 日志记录器
// @param blackList IP黑名单列表，支持精确IP和CIDR格式的IP范围
// @return gin.HandlerFunc Gin中间件函数
func EnableIPBlackList(logger *zap.SugaredLogger, blackList []string) gin.HandlerFunc {
	// 如果黑名单为空，则不进行过滤
	if len(blackList) == 0 {
		return func(ctx *gin.Context) {
			_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipBlack")
			defer span.End()
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		}
	}

	// 创建CIDR列表（一次性解析，提高性能）
	cidrList, err := iphelper.NewCIDRList(blackList)
	if err != nil {
		logger.Errorf("[IP BlackList] Failed to create CIDR list: %v", err)
		// 如果解析失败，默认拒绝所有请求
		return func(ctx *gin.Context) {
			_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipBlack")
			defer span.End()
			logger.Warnf("[IP BlackList] Blocked due to invalid CIDR configuration, Path: %s", ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Access forbidden: Invalid IP black list configuration",
				"data":    nil,
				"success": false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: Invalid IP black list configuration")
		}
	}

	return func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipBlack")
		defer span.End()
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(ctx)
		parsedIP := net.ParseIP(clientIP)
		if parsedIP == nil {
			logger.Warnf("[IP BlackList] Invalid IP address: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Access forbidden: Invalid IP address",
				"data":    nil,
				"success": false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: Invalid IP address")
			return
		}

		// 使用CIDRList检查IP是否在黑名单中
		if cidrList.Contains(parsedIP) {
			logger.Warnf("[IP BlackList] Blocked IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Access forbidden: IP in black list",
				"data":    nil,
				"success": false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: IP in black list")
		} else {
			logger.Debugf("[IP BlackList] Allowed IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		}
	}
}
