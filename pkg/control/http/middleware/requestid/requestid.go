package requestid

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func EnableRequestID(webLogger *zap.SugaredLogger) gin.HandlerFunc {
	return requestid.New(
		requestid.WithGenerator(func() string {
			return uuid.GetUUID()
		}),
		// requestid.WithCustomHeaderStrKey("golang-example-request-id"),
		requestid.WithHandler(func(ctx *gin.Context, requestID string) {
			_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "requestId")
			defer span.End()
			webLogger.Debugf("clientIp: %s requestProto %v requestURL: %s requestMethod: %s agent: %v requestID %v",
				ctx.ClientIP(), ctx.Request.Proto, ctx.Request.URL.String(), ctx.Request.Method, ctx.Request.UserAgent(), requestID)
			span.SetStatus(codes.Ok, "success")
		}),
	)
}
