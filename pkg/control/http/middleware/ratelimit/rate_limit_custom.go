package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/iphelper"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// CustomRateLimiter 自定义令牌桶限流器
type CustomRateLimiter struct {
	mutex          sync.Mutex
	rate           int64     // 每秒生成的令牌数
	capacity       int64     // 令牌桶容量
	tokens         float64   // 当前令牌数
	lastRefillTime time.Time // 上次填充令牌的时间
}

// NewCustomRateLimiter 创建自定义令牌桶限流器
// @param rate 每秒生成的令牌数
// @param capacity 令牌桶容量
// @return *CustomRateLimiter 限流器实例
func NewCustomRateLimiter(rate int64, capacity int64) *CustomRateLimiter {
	return &CustomRateLimiter{
		rate:           rate,
		capacity:       capacity,
		tokens:         float64(capacity), // 初始时桶是满的
		lastRefillTime: time.Now(),
	}
}

// Allow 检查是否允许请求通过
// @return bool true表示允许请求通过，false表示拒绝请求
func (rl *CustomRateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// 计算从上次调用到现在应该生成的令牌数
	now := time.Now()
	duration := now.Sub(rl.lastRefillTime).Seconds()
	newTokens := duration * float64(rl.rate)

	// 如果有新的令牌生成，更新令牌数和最后填充时间
	if newTokens > 0 {
		// 确保令牌数不超过桶容量
		rl.tokens = min(float64(rl.capacity), rl.tokens+newTokens)
		rl.lastRefillTime = now
	}

	// 检查是否有足够的令牌
	if rl.tokens >= 1 {
		// 消耗一个令牌
		rl.tokens--
		return true
	}

	// 没有足够的令牌
	return false
}

// EnableCustomRateLimit 使用自定义令牌桶算法的限流中间件
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @return gin.HandlerFunc Gin中间件函数
func EnableCustomRateLimit(logger *zap.SugaredLogger, rps int64, burst int) gin.HandlerFunc {
	// 创建全局限流器管理器
	limiters := make(map[string]*CustomRateLimiter)
	limitersMutex := sync.RWMutex{}

	return func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimitCustom")
		defer span.End()
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(ctx)

		// 获取或创建该IP的限流器
		limitersMutex.RLock()
		limiter, exists := limiters[clientIP]
		limitersMutex.RUnlock()

		if !exists {
			limitersMutex.Lock()
			// 双重检查，避免竞态条件
			if limiter, exists = limiters[clientIP]; !exists {
				limiter = NewCustomRateLimiter(rps, int64(burst))
				limiters[clientIP] = limiter
			}
			limitersMutex.Unlock()
		}

		// 检查是否允许请求
		if limiter.Allow() {
			logger.Debugf("[CustomRateLimit] Allowed request from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		} else {
			logger.Warnf("[CustomRateLimit] Too many requests from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Header("X-RateLimit-Type", "custom")
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "Too many requests",
				"data":    nil,
				"success": false,
			})
			span.SetStatus(codes.Error, "Too may requests")
			ctx.Abort()
		}
	}
}
