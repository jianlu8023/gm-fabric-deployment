package response

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func EnableResponseLog(webLogger *zap.SugaredLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "response")
		defer span.End()
		blw := &bodyLogWriter{body: bytes.NewBufferString(""),
			ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		startTime := time.Now()
		ctx.Next()
		duration := time.Since(startTime).String()
		// ctx.Header("X-Response-Time", duration)
		// webLogger.Info("Response",
		// 	zap.String("DurationTime", duration),
		// 	zap.Any("ResponseBody", blw.body.String()),
		// )
		webLogger.Debugf("Response DurationTime: %v responseBody: %v", duration, blw.body.String())
		span.SetStatus(codes.Ok, "success")
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
