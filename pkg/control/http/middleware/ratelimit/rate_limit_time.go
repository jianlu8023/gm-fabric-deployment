package ratelimit

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/iphelper"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimiter 限流管理器
type RateLimiter struct {
	limiters map[string]*rate.Limiter // 存储每个IP的限流器
	mutex    sync.Mutex               // 用于保护limiters的并发访问
	rps      rate.Limit               // 每秒请求数
	burst    int                      // 令牌桶突发大小
}

// NewRateLimiter 创建一个新的限流管理器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @return *RateLimiter 限流管理器实例
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// getLimiter 根据IP获取或创建限流器
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		// 为新IP创建限流器
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		// 存储限流器
		rl.limiters[ip] = limiter
	}

	return limiter
}

// Allow 检查IP是否允许继续请求
func (rl *RateLimiter) Allow(ip string) bool {
	limiter := rl.getLimiter(ip)
	return limiter.Allow()
}

// EnableRateLimit 创建并返回限流中间件
// @param logger 日志记录器
// @param rps 每秒请求数限制
// @param burst 令牌桶突发大小
// @return gin.HandlerFunc Gin中间件函数
func EnableRateLimit(logger *zap.SugaredLogger, rps int64, burst int) gin.HandlerFunc {
	// 创建全局限流器实例
	rateLimiter := NewRateLimiter(float64(rps), burst)

	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := iphelper.GetClientIP(c)

		// 检查是否允许请求
		if rateLimiter.Allow(clientIP) {
			logger.Debugf("[RateLimit] Allowed request from IP: %s, Path: %s", clientIP, c.Request.URL.Path)
			c.Next()
		} else {
			logger.Warnf("[RateLimit] Too many requests from IP: %s, Path: %s", clientIP, c.Request.URL.Path)
			c.Header("X-RateLimit-Type", "time")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    http.StatusTooManyRequests,
				"message": "Too many requests",
				"data":    nil,
			})
			c.Abort()
		}
	}
}
