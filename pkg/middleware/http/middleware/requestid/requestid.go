package requestid

import (
	"fmt"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

func EnableRequestID() gin.HandlerFunc {
	return requestid.New(
		requestid.WithGenerator(func() string {
			return fmt.Sprintf("%d", time.Now().UnixMilli())
		}),
		requestid.WithCustomHeaderStrKey("your-customer-key"),
		requestid.WithHandler(func(c *gin.Context, requestID string) {
			log.Printf("RequestID: %s", requestID)
		}),
	)
}
