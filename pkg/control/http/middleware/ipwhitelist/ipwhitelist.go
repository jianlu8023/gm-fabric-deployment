package ipwhitelist

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/iphelper"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// EnableIPWhiteList 创建并返回IP白名单中间件（支持精确IP和CIDR范围）
// @param logger 日志记录器
// @param whiteList IP白名单列表，支持精确IP和CIDR格式的IP范围
// @return gin.HandlerFunc Gin中间件函数
func EnableIPWhiteList(logger *zap.SugaredLogger, whiteList []string) gin.HandlerFunc {
	// 如果白名单为空，则不进行过滤
	if len(whiteList) == 0 {
		return func(ctx *gin.Context) {
			savedCtx := ctx.Request.Context()
			defer func() {
				ctx.Request = ctx.Request.WithContext(savedCtx)
			}()
			tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipWhite")
			defer span.End()
			ctx.Request = ctx.Request.WithContext(tCtx)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		}
	}

	// 创建CIDR列表（一次性解析，提高性能）
	cidrList, err := iphelper.NewCIDRList(whiteList)
	if err != nil {
		logger.Errorf("[IP WhiteList] Failed to create CIDR list: %v", err)
		// 如果解析失败，默认拒绝所有请求
		return func(ctx *gin.Context) {
			savedCtx := ctx.Request.Context()
			defer func() {
				ctx.Request = ctx.Request.WithContext(savedCtx)
			}()
			tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipWhite")
			defer span.End()
			ctx.Request = ctx.Request.WithContext(tCtx)
			logger.Warnf("[IP WhiteList] Blocked due to invalid CIDR configuration, Path: %s", ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, commonhttp.BaseResponse{
				Code:    http.StatusForbidden,
				Message: "业务处理失败",
				Data:    "Access forbidden: Invalid IP white list configuration",
				Success: false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: Invalid IP white list configuration")
		}
	}

	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "ipWhite")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(ctx)
		parsedIP := net.ParseIP(clientIP)
		if parsedIP == nil {
			logger.Warnf("[IP WhiteList] Invalid IP address: %s, Path: %s", clientIP, ctx.Request.URL.Path)
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

		// 使用CIDRList检查IP是否在白名单中
		if cidrList.Contains(parsedIP) {
			logger.Debugf("[IP WhiteList] Allowed IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		} else {
			logger.Warnf("[IP WhiteList] Forbidden IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.JSON(http.StatusForbidden, commonhttp.BaseResponse{
				Code:    http.StatusForbidden,
				Message: "业务处理失败",
				Data:    "Access forbidden: IP not in white list",
				Success: false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Access forbidden: IP not in white list")
		}
	}
}
