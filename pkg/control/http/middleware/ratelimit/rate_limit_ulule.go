package ratelimit

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"

	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// EnableRateLimitUlule 使用ulule/limiter库实现的限流中间件（仅使用内存存储）
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @param trustedProxies 可信代理网段列表，用于正确解析客户端真实IP；为 nil 时仅使用 RemoteAddr
// @return gin.HandlerFunc Gin中间件函数
func EnableRateLimitUlule(logger *zap.SugaredLogger, rps int64, burst int, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	// 使用格式化的速率字符串，支持更灵活的限流配置
	// 格式示例: "10-S" (每秒10个), "100-M" (每分钟100个), "500-H" (每小时500个)
	rateStr := fmt.Sprintf("%d-S", rps)
	rate, err := limiter.NewRateFromFormatted(rateStr)
	if err != nil {
		logger.Errorf("[RateLimit] Failed to create rate: %v", err)
		// 回退到基本配置
		rate = limiter.Rate{Period: time.Second, Limit: rps}
	}

	// 只使用内存存储
	store := memory.NewStore()
	logger.Info("[RateLimit] Using memory store for rate limiting")

	// 创建限流器实例
	limiterInstance := limiter.New(store, rate)

	// 使用ulule提供的中间件
	middleware := mgin.NewMiddleware(limiterInstance, mgin.WithLimitReachedHandler(func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimitUlule")
		defer span.End()
		ctx.Header("X-RateLimit-Type", "ulule")
		ctx.JSON(http.StatusTooManyRequests, commonhttp.BaseResponse{
			Code:    http.StatusTooManyRequests,
			Message: "业务处理失败",
			Data:    "Too many requests",
			Success: false,
		})
		span.SetStatus(codes.Error, "Too many requests")
	}))

	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimitUlule")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)

		// 调用ulule中间件
		// 注意：ulule 内部使用 gin.ClientIP() 获取客户端IP，不再修改请求头以避免污染后续中间件与上游服务
		middleware(ctx)

		// 如果请求被限流，记录日志
		if ctx.IsAborted() && ctx.Writer.Status() == http.StatusTooManyRequests {
			clientIP := iphelper.GetClientIP(ctx, trustedProxies)
			logger.Warnf("[RateLimit] Too many requests from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			span.SetStatus(codes.Error, "Too many requests")
			return
		}
		span.SetStatus(codes.Ok, "success")
	}
}
