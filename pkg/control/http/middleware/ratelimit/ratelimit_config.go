package ratelimit

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.uber.org/zap"
)

// Type 限流类型
type Type string

const (
	// TypeTime 使用golang.org/x/time/rate库实现的限流
	TypeTime Type = "time"
	// TypeUlule 使用ulule/limiter库实现的限流
	TypeUlule Type = "ulule"
	// TypeAnts 使用ants库实现的限流
	TypeAnts Type = "ants"
	// TypeCustom 使用自定义令牌桶算法实现的限流
	TypeCustom Type = "custom"
	// TypeJuju 使用github.com/juju/ratelimit库实现的限流
	TypeJuju Type = "juju"
)

// Config 限流配置
type Config struct {
	Type  Type  // 限流类型
	RPS   int64 // 每秒请求数限制
	Burst int   // 令牌桶突发大小
}

// NewRateLimitMiddleware 创建限流中间件工厂函数
// 根据配置选择不同的限流实现方式
// @param logger 日志记录器
// @param config 限流配置
// @param trustedProxies 可信代理网段列表，用于正确解析客户端真实IP；为 nil 时仅使用 RemoteAddr
// @return gin.HandlerFunc Gin中间件函数
func NewRateLimitMiddleware(logger *zap.SugaredLogger, config Config, trustedProxies *iphelper.CIDRList) gin.HandlerFunc {
	// 验证配置
	if config.RPS <= 0 {
		logger.Warnf("[RateLimit] Invalid RPS value, rate limiting disabled")
		// 返回一个空中间件，不做任何处理
		return func(ctx *gin.Context) {
			savedCtx := ctx.Request.Context()
			defer func() {
				ctx.Request = ctx.Request.WithContext(savedCtx)
			}()
			tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "rateLimit")
			defer span.End()
			ctx.Request = ctx.Request.WithContext(tCtx)
			ctx.Next()
		}
	}

	// 如果未设置burst，则默认为RPS的2倍
	burst := config.Burst
	if burst <= 0 {
		burst = int(config.RPS * 2)
	}

	// 根据类型选择不同的限流实现
	logger.Infof("[RateLimit] Creating %s rate limiter with RPS: %d, burst: %d", config.Type, config.RPS, burst)

	switch config.Type {
	case TypeTime:
		return EnableRateLimit(logger, config.RPS, burst, trustedProxies)
	case TypeUlule:
		return EnableRateLimitUlule(logger, config.RPS, burst, trustedProxies)
	case TypeAnts:
		return EnableRateLimitAnts(logger, config.RPS, burst, trustedProxies)
	case TypeCustom:
		return EnableCustomRateLimit(logger, config.RPS, burst, trustedProxies)
	case TypeJuju:
		return EnableRateLimitJuju(logger, config.RPS, burst, trustedProxies)
	default:
		logger.Warnf("[RateLimit] Unknown rate limit type: %s, using default (time/rate)", config.Type)
		return EnableRateLimit(logger, config.RPS, burst, trustedProxies)
	}
}
