package ratelimit

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/juju/ratelimit"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// EnableRateLimitJuju 启用基于github.com/juju/ratelimit库的限流中间件
//
// @description 使用juju/ratelimit库的令牌桶实现，为每个客户端IP维护独立的限流计数
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @param trustedProxies 可信代理网段列表，用于正确解析客户端真实IP；为 nil 时仅使用 RemoteAddr
// @return gin.HandlerFunc Gin中间件函数
func EnableRateLimitJuju(logger *zap.SugaredLogger, rps int64, burst int, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	// 获取管理器配置
	maxSize, expireTime := GetManagerConfig()
	manager := NewIPRateLimiterManager(context.Background(), maxSize, expireTime)

	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimitJuju")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 获取客户端IP地址
		clientIP := iphelper.GetClientIP(ctx, trustedProxies)

		// 获取或创建限流器
		limiter := manager.GetOrCreate(clientIP, func() interface{} {
			return ratelimit.NewBucketWithRate(float64(rps), int64(burst))
		})

		// 类型断言为 *ratelimit.Bucket
		bucket, ok := limiter.(*ratelimit.Bucket)
		if !ok {
			// 理论上不会发生，防御性检查
			logger.Errorf("[RateLimitJuju] Invalid limiter type for IP: %s", clientIP)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
			return
		}

		// 检查是否允许请求通过
		if bucket.TakeAvailable(1) == 1 {
			logger.Debugf("[RateLimitJuju] Allowed request from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		} else {
			logger.Warnf("[RateLimitJuju] Too many requests from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Header("X-RateLimit-Type", "juju")
			ctx.JSON(http.StatusTooManyRequests, commonhttp.BaseResponse{
				Code:    http.StatusTooManyRequests,
				Message: "业务处理失败",
				Data:    "Too many requests",
				Success: false,
			})
			span.SetStatus(codes.Error, "Too many requests")
			ctx.Abort()
		}
	}
}
