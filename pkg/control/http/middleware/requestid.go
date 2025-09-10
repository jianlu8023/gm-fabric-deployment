package middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/random/uuid"
	"go.uber.org/zap"
)

func EnableRequestID(webLogger *zap.SugaredLogger) gin.HandlerFunc {
	return requestid.New(
		requestid.WithGenerator(func() string {
			return uuid.GetUUID()
		}),
		requestid.WithCustomHeaderStrKey("your-customer-key"),
		requestid.WithHandler(func(ctx *gin.Context, requestID string) {
			webLogger.Debugf("clientIp: %s requestProto %v requestURL: %s requestMethod: %s agent: %v requestID %v",
				ctx.ClientIP(), ctx.Request.Proto, ctx.Request.URL.String(), ctx.Request.Method, ctx.Request.UserAgent(), requestID)
		}),
	)
}
