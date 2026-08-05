package ratelimit

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"

	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// EnableRateLimit 创建并返回限流中间件（基于golang.org/x/time/rate库）
//
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @param trustedProxies 可信代理网段列表，用于正确解析客户端真实IP；为 nil 时仅使用 RemoteAddr
// @return gin.HandlerFunc Gin中间件函数
func EnableRateLimit(logger *zap.SugaredLogger, rps int64, burst int, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	// 获取管理器配置
	maxSize, expireTime := GetManagerConfig()
	manager := NewIPRateLimiterManager(context.Background(), maxSize, expireTime)

	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimitTime")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(ctx, trustedProxies)

		// 获取或创建限流器
		limiter := manager.GetOrCreate(clientIP, func() interface{} {
			return rate.NewLimiter(rate.Limit(rps), burst)
		})

		// 类型断言为 *rate.Limiter
		rl, ok := limiter.(*rate.Limiter)
		if !ok {
			// 理论上不会发生，防御性检查
			logger.Errorf("[RateLimit] Invalid limiter type for IP: %s", clientIP)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
			return
		}

		// 检查是否允许请求
		if rl.Allow() {
			logger.Debugf("[RateLimit] Allowed request from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		} else {
			logger.Warnf("[RateLimit] Too many requests from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Header("X-RateLimit-Type", "time")
			ctx.JSON(http.StatusTooManyRequests, commonhttp.BaseResponse{
				Code:    http.StatusTooManyRequests,
				Message: "业务处理失败",
				Data:    "Too many requests",
				Success: false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "Too many requests")
		}
	}
}
