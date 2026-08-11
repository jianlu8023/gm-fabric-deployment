package ratelimit

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"

	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// SimpleRateLimiter 简易令牌桶限流器（实现简单计数器算法）
//
// @description 使用浮点令牌数避免低RPS下的令牌截断问题
type SimpleRateLimiter struct {
	mu              sync.Mutex
	rateLimitPerSec float64 // 每秒生成的令牌数（浮点）
	lastCheckTime   time.Time
	tokens          float64 // 当前令牌数（浮点）
	bucketSize      float64 // 令牌桶容量
}

// NewSimpleRateLimiter 创建简易限流器
//
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @return *SimpleRateLimiter 限流器实例
func NewSimpleRateLimiter(rps int64, burst int) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		rateLimitPerSec: float64(rps),
		lastCheckTime:   time.Now(),
		tokens:          float64(burst), // 初始时桶是满的
		bucketSize:      float64(burst),
	}
}

// Allow 检查是否允许继续请求
//
// @description 使用浮点运算避免令牌截断问题：
//  1. 根据 elapsed 时间计算应增加的令牌数
//  2. 使用 float64 维护令牌数，避免 int64 截断
//  3. 令牌消耗后可能为小数，但 Accumulate 足够快时不会丢失
//
// @return bool true表示允许请求通过，false表示拒绝请求
func (rl *SimpleRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastCheckTime).Seconds()

	// 计算应该增加的令牌数（使用float64）
	if elapsed > 0 {
		newTokens := elapsed * rl.rateLimitPerSec
		// 添加令牌，但不超过桶容量
		rl.tokens = min(rl.bucketSize, rl.tokens+newTokens)
		rl.lastCheckTime = now
	}

	// 检查是否有足够的令牌
	if rl.tokens >= 1.0 {
		// 消耗一个令牌
		rl.tokens -= 1.0
		return true
	}

	// 没有令牌可用
	return false
}

// EnableSimpleRateLimit 使用简易令牌桶算法的限流中间件
//
// @description 使用浮点令牌数避免低RPS下的令牌截断问题
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @param trustedProxies 可信代理网段列表，用于正确解析客户端真实IP；为 nil 时仅使用 RemoteAddr
// @return gin.HandlerFunc Gin中间件函数
func EnableSimpleRateLimit(logger *zap.SugaredLogger, rps int64, burst int, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	// 获取管理器配置
	maxSize, expireTime := GetManagerConfig()
	manager := NewIPRateLimiterManager(context.Background(), maxSize, expireTime)

	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimitSimple")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(ctx, trustedProxies)

		// 获取或创建限流器
		limiter := manager.GetOrCreate(clientIP, func() interface{} {
			return NewSimpleRateLimiter(rps, burst)
		})

		// 类型断言为 *SimpleRateLimiter
		rl, ok := limiter.(*SimpleRateLimiter)
		if !ok {
			// 理论上不会发生，防御性检查
			logger.Errorf("[http/Simple] Invalid limiter type for IP: %s", clientIP)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
			return
		}

		// 检查是否允许请求
		if rl.Allow() {
			logger.Debugf("[http/Simple] Allowed request from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		} else {
			logger.Warnf("[http/Simple] Too many requests from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Header("X-RateLimit-Type", "simple")
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

// EnableRateLimitAnts 使用ants库实现的限流中间件（保留兼容）
//
// @description 此函数已重命名为 EnableSimpleRateLimit，保留此函数以兼容旧代码。
//
//	内部实现未使用ants库，实为自定义令牌桶算法。
//
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @param trustedProxies 可信代理网段列表
// @return gin.HandlerFunc Gin中间件函数
// @deprecated 请使用 EnableSimpleRateLimit 替代
func EnableRateLimitAnts(logger *zap.SugaredLogger, rps int64, burst int, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	return EnableSimpleRateLimit(logger, rps, burst, trustedProxies)
}
