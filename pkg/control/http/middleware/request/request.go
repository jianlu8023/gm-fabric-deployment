package request

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func EnableRequestLog(webLogger *zap.SugaredLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body != nil {
			webLogger.Debugf("clientIp: %s requestURL: %s requestMethod: %s", ctx.ClientIP(), ctx.Request.URL.String(), ctx.Request.Method)

		}
		ctx.Next()
	}
}
