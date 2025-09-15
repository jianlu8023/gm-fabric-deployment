package response

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func EnableResponseLog(webLogger *zap.SugaredLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""),
			ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		startTime := time.Now()
		ctx.Next()
		duration := time.Since(startTime).String()
		// ctx.Header("X-Response-Time", duration)
		webLogger.Info("Response",
			zap.String("DurationTime", duration),
			zap.Any("ResponseBody", blw.body.String()),
		)

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
