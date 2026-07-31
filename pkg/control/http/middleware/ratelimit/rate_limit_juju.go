package ratelimit

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/juju/ratelimit"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// RateLimiterJuju 基于github.com/juju/ratelimit库的限流器
// 这个限流器会为每个客户端IP维护一个令牌桶
type RateLimiterJuju struct {
	// limiters map[string]*ratelimit.Bucket // 存储每个IP的令牌桶
	limiters concurrent.Map[string, *ratelimit.Bucket] // 存储每个IP的令牌桶
	rps      int64                                     // 每秒生成的令牌数
	burst    int                                       // 令牌桶的容量
}

// NewRateLimiterJuju 创建一个新的RateLimiterJuju实例
// @param rps 每秒生成的令牌数
// @param burst 令牌桶的容量
// @return *RateLimiterJuju 限流器实例
func NewRateLimiterJuju(rps int64, burst int) *RateLimiterJuju {
	return &RateLimiterJuju{
		// limiters: make(map[string]*ratelimit.Bucket),
		limiters: concurrentmap.NewRWMap[string, *ratelimit.Bucket](),
		rps:      rps,
		burst:    burst,
	}
}

// Allow 检查是否允许请求通过
// @param ip 客户端IP
// @return bool 是否允许通过
func (rl *RateLimiterJuju) Allow(ip string) bool {
	// 为IP获取或创建令牌桶
	// bucket, exists := rl.limiters[ip]
	bucket, exists := rl.limiters.Get(ip)
	if !exists {
		// 创建一个新的令牌桶，每秒钟添加rps个令牌，容量为burst
		// ratelimit.NewBucketWithRate会创建一个令牌桶，参数是每秒添加的令牌数和容量
		bucket = ratelimit.NewBucketWithRate(float64(rl.rps), int64(rl.burst))
		// 保存令牌桶，供后续请求使用
		// rl.limiters[ip] = bucket
		rl.limiters.Put(ip, bucket)
	}

	// 尝试获取一个令牌，如果成功则返回true，否则返回false
	return bucket.TakeAvailable(1) == 1
}

// EnableRateLimitJuju 启用基于github.com/juju/ratelimit库的限流中间件
// 该中间件使用令牌桶算法，为每个客户端IP维护独立的限流计数
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @param trustedProxies 可信代理网段列表，用于正确解析客户端真实IP；为 nil 时仅使用 RemoteAddr
// @return gin.HandlerFunc Gin中间件函数
func EnableRateLimitJuju(logger *zap.SugaredLogger, rps int64, burst int, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	// 创建一个限流器实例
	limiter := NewRateLimiterJuju(rps, burst)

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

		// 检查是否允许请求通过
		if !limiter.Allow(clientIP) {
			// 如果不允许，记录日志并返回429状态码
			logger.Warnf("[RateLimitJuju] Request from %s denied due to rate limiting", clientIP)
			ctx.Header("X-RateLimit-Type", "juju")
			ctx.JSON(http.StatusTooManyRequests, commonhttp.BaseResponse{
				Code:    http.StatusTooManyRequests,
				Message: "业务处理失败",
				Data:    "Too many requests",
				Success: false,
			})
			span.SetStatus(codes.Error, "Too many requests")
			ctx.Abort()
			return
		}

		// 允许请求继续处理
		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}
