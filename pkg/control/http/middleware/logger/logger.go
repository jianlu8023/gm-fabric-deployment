package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if params.IsOutputColor() {
				statusColor = params.StatusCodeColor()
				methodColor = params.MethodColor()
				resetColor = params.ResetColor()
			}
			if params.Latency > time.Minute {
				// Truncate in a golang < 1.8 safe way
				params.Latency = params.Latency - params.Latency%time.Second
			}
			return fmt.Sprintf("%v %s %3d %s %13v | %15s |%s %-7s %s %#v\n%s",
				params.TimeStamp.Format("2006-01-02 15:04:05.000"),
				statusColor, params.StatusCode, resetColor,
				params.Latency,
				params.ClientIP,
				methodColor, params.Method, resetColor,
				params.Path,
				params.ErrorMessage,
			)
		},
	})
}
