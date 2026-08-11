package recovery

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// 默认值常量
const (
	// defaultEnableStack 默认是否输出堆栈（开发环境true，生产环境应配置为false）
	defaultEnableStack = true
	// defaultMessage panic恢复后返回给客户端的默认响应消息
	defaultMessage = "Server Internal Error"
	// defaultLogLevel 默认日志级别
	defaultLogLevel = "error"
)

// EnableRecovery recover掉项目可能出现的panic
//
// @description 捕获HTTP请求处理过程中的panic，防止整个服务崩溃。
//
//	支持两种模式：带堆栈信息的详细日志和不带堆栈的简洁日志。
//	对broken pipe等常见网络异常做特殊处理，避免不必要的堆栈输出。
//	配置驱动：通过 cfg 参数控制是否输出堆栈、响应消息等，配置为 nil 时使用默认值。
//
// @param webLogger *zap.SugaredLogger 用于记录panic信息的日志器
// @param cfg *config.RecoveryConfig recovery配置，为nil时使用默认值（enableStack=true）
// @return gin.HandlerFunc Gin中间件函数
func EnableRecovery(webLogger *zap.SugaredLogger, cfg *config.RecoveryConfig) gin.HandlerFunc {
	// 解析配置，nil 时使用默认值
	enableStack := defaultEnableStack
	customMessage := defaultMessage
	if cfg != nil {
		enableStack = cfg.EnableStack
		if cfg.CustomMessage != "" {
			customMessage = cfg.CustomMessage
		}
	}

	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "recovery")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(ctx.Request, false)
				if brokenPipe {
					webLogger.Errorf("[http/Recovery] request url %v with request %v failed: %v",
						ctx.Request.URL.Path,
						string(httpRequest),
						err,
					)
					// webLogger.Error(ctx.Request.URL.Path,
					// 	zap.Any("error", err),
					// 	zap.String("request", string(httpRequest)),
					// )
					// If the connection is dead, we can't write a status to it.
					_ = ctx.Error(safeError(err)) // nolint: errcheck
					ctx.Abort()
					span.SetStatus(codes.Error, "")
					return
				}

				if enableStack {
					webLogger.Errorf("[http/Recovery] request url %v with request %v with stack %v failed: %v",
						ctx.Request.URL.Path,
						string(httpRequest),
						string(debug.Stack()),
						err,
					)
					// webLogger.Error("[http/Recovery]",
					// 	zap.Any("error", err),
					// 	zap.String("request", string(httpRequest)),
					// 	zap.String("stack", string(debug.Stack())),
					// )
				} else {
					// webLogger.Error("[http/Recovery]",
					// 	zap.Any("error", err),
					// 	zap.String("request", string(httpRequest)),
					// )
					webLogger.Errorf("[http/Recovery] request url %v with request %v failed: %v",
						ctx.Request.URL.Path,
						string(httpRequest),
						err,
					)
				}
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, commonhttp.BaseResponse{
					Code:    http.StatusInternalServerError,
					Message: "业务处理失败",
					Success: false,
					Data:    customMessage,
				})
				span.SetStatus(codes.Error, "")
			}
		}()
		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}

// safeError 将 recover() 返回的任意类型安全转换为 error 接口
//
// @description recover() 返回 interface{}，可能是 string、int、自定义结构体等任意类型，
//
//	直接调用 err.(error) 类型断言会导致二次panic。本函数使用 type switch
//	对已实现 error 接口的类型直接返回，其余类型用 fmt.Errorf 包装为 error。
//
// @param err interface{} recover() 返回的原始值
// @return error 转换后的 error 对象
func safeError(err interface{}) error {
	switch v := err.(type) {
	case error:
		return v
	default:
		return fmt.Errorf("%v", err)
	}
}
