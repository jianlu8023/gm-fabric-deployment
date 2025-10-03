package recovery

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// EnableRecovery recover掉项目可能出现的panic
func EnableRecovery(webLogger *zap.SugaredLogger, stack bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "recovery")
		defer span.End()
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
					webLogger.Errorf("request url %v with request %v failed: %v",
						ctx.Request.URL.Path,
						string(httpRequest),
						err,
					)
					// webLogger.Error(ctx.Request.URL.Path,
					// 	zap.Any("error", err),
					// 	zap.String("request", string(httpRequest)),
					// )
					// If the connection is dead, we can't write a status to it.
					_ = ctx.Error(err.(error)) // nolint: errcheck
					ctx.Abort()
					span.SetStatus(codes.Error, "")
					return
				}

				if stack {
					webLogger.Errorf("request url %v with request %v with stack %v failed: %v",
						ctx.Request.URL.Path,
						string(httpRequest),
						string(debug.Stack()),
						err,
					)
					// webLogger.Error("[Recovery from panic]",
					// 	zap.Any("error", err),
					// 	zap.String("request", string(httpRequest)),
					// 	zap.String("stack", string(debug.Stack())),
					// )
				} else {
					// webLogger.Error("[Recovery from panic]",
					// 	zap.Any("error", err),
					// 	zap.String("request", string(httpRequest)),
					// )
					webLogger.Errorf("request url %v with request %v failed: %v",
						ctx.Request.URL.Path,
						string(httpRequest),
						err,
					)
				}
				ctx.AbortWithStatus(http.StatusInternalServerError)
				span.SetStatus(codes.Error, "")
			}
		}()
		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}
