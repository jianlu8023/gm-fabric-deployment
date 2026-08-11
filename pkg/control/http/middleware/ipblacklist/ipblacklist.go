package ipblacklist

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"

	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// EnableIPBlackList 创建并返回IP黑名单中间件（支持精确IP和CIDR范围）
// @param logger 日志记录器
// @param blackList IP黑名单列表，支持精确IP和CIDR格式的IP范围
// @param trustedProxies 可信代理网段列表，用于正确解析客户端真实IP；为 nil 时仅使用 RemoteAddr
// @return gin.HandlerFunc Gin中间件函数
func EnableIPBlackList(logger *zap.SugaredLogger, blackList []string, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	// 如果黑名单为空，则不进行过滤，直接返回空中间件
	if len(blackList) == 0 {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}

	// 创建CIDR列表（一次性解析，提高性能）
	cidrList, err := iphelper.NewCIDRList(blackList)
	if err != nil {
		logger.Errorf("[http/IpBlackList] Failed to create CIDR list: %v", err)
		// 如果解析失败，默认拒绝所有请求
		return func(ctx *gin.Context) {
			savedCtx := ctx.Request.Context()
			defer func() {
				ctx.Request = ctx.Request.WithContext(savedCtx)
			}()
			tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipBlack")
			defer span.End()
			ctx.Request = ctx.Request.WithContext(tCtx)
			logger.Warnf("[http/IpBlackList] Blocked due to invalid CIDR configuration, Path: %s", ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, commonhttp.BaseResponse{
				Code:    http.StatusForbidden,
				Message: "业务处理失败",
				Data:    "Access forbidden: Invalid IP black list configuration",
				Success: false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: Invalid IP black list configuration")
		}
	}

	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipBlack")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(ctx, trustedProxies)
		parsedIP := net.ParseIP(clientIP)
		if parsedIP == nil {
			logger.Warnf("[http/IpBlackList] Invalid IP address: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, commonhttp.BaseResponse{
				Code:    http.StatusForbidden,
				Message: "业务处理失败",
				Data:    "Access forbidden: Invalid IP address",
				Success: false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: Invalid IP address")
			return
		}

		// 使用CIDRList检查IP是否在黑名单中
		if cidrList.Contains(parsedIP) {
			logger.Warnf("[http/IpBlackList] Blocked IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, commonhttp.BaseResponse{
				Code:    http.StatusForbidden,
				Message: "业务处理失败",
				Data:    "Access forbidden: IP in black list",
				Success: false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: IP in black list")
		} else {
			logger.Debugf("[http/IpBlackList] Allowed IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		}
	}
}
