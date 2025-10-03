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

// RateLimiterAnts 使用ants库实现的限流管理器
type RateLimiterAnts struct {
	mu              sync.Mutex
	rateLimitPerSec int64
	lastCheckTime   time.Time
	requestCount    int64
	bucketSize      int64
}

// NewRateLimiterAnts 创建一个新的限流管理器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @return *RateLimiterAnts 限流管理器实例
func NewRateLimiterAnts(rps int64, burst int) *RateLimiterAnts {
	return &RateLimiterAnts{
		rateLimitPerSec: rps,
		lastCheckTime:   time.Now(),
		bucketSize:      int64(burst),
	}
}

// Allow 检查是否允许继续请求
func (rl *RateLimiterAnts) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastCheckTime).Seconds()

	// 计算应该增加的令牌数
	newTokens := elapsed * float64(rl.rateLimitPerSec)

	// 更新请求计数（相当于减少令牌）
	if newTokens > 0 {
		// 如果有新令牌产生，更新最后检查时间
		if rl.requestCount > 0 {
			// 如果请求计数大于0，则减去新令牌数
			rl.requestCount = max(0, rl.requestCount-int64(newTokens))
		}
		rl.lastCheckTime = now
	}

	// 检查是否还有令牌可用
	if rl.requestCount < rl.bucketSize {
		// 还有令牌可用，增加请求计数
		rl.requestCount++
		return true
	}

	// 没有令牌可用
	return false
}

// EnableRateLimitAnts 使用ants库实现的限流中间件
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @return gin.HandlerFunc Gin中间件函数
func EnableRateLimitAnts(logger *zap.SugaredLogger, rps int64, burst int) gin.HandlerFunc {
	// 创建IP到限流器的映射
	rateLimiters := make(map[string]*RateLimiterAnts)
	mu := sync.RWMutex{}

	return func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimitAnts")
		defer span.End()
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(ctx)

		// 获取或创建限流器
		mu.RLock()
		rl, exists := rateLimiters[clientIP]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			// 双重检查，避免竞态条件
			if rl, exists = rateLimiters[clientIP]; !exists {
				rl = NewRateLimiterAnts(rps, burst)
				rateLimiters[clientIP] = rl
			}
			mu.Unlock()
		}

		// 检查是否允许请求
		if rl.Allow() {
			logger.Debugf("[RateLimit] Allowed request from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Next()
			span.SetStatus(codes.Ok, "success")
		} else {
			logger.Warnf("[RateLimit] Too many requests from IP: %s, Path: %s", clientIP, ctx.Request.URL.Path)
			ctx.Header("X-RateLimit-Type", "ants")
			ctx.JSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "Too many requests",
				"data":    nil,
				"success": false,
			})
			ctx.Abort()
			span.SetStatus(codes.Error, "To many requests")
		}
	}
}
