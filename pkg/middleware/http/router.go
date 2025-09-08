package http

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type MyRouter struct {
	Name        string                 `yaml:"name"`
	Uri         string                 `yaml:"uri"`
	Method      string                 `yaml:"method"`
	HandlerFunc func(ctx *gin.Context) `yaml:"handlerFunc"`
	Enabled     bool                   `yaml:"enabled"`
	Desc        string                 `yaml:"desc"`
}

func NewRouter() []MyRouter {
	routers := map[string][]MyRouter{
		"base": {
			{
				Name:   "ping",
				Uri:    "/ping",
				Method: http.MethodGet,
				HandlerFunc: func(ctx *gin.Context) {
					id := requestid.Get(ctx)
					SuccessResponse(ctx, gin.H{
						"request_id": id,
						"message":    "pong",
					})
				},
				Enabled: true,
				Desc:    "ping",
			},
			{
				Name:   "health",
				Uri:    "/health",
				Method: http.MethodGet,
				HandlerFunc: func(ctx *gin.Context) {
					SuccessResponse(ctx, "ok")
				},
				Enabled: true,
				Desc:    "health check",
			},
			{
				Name:   "sse",
				Uri:    "/sse",
				Method: http.MethodGet,
				HandlerFunc: func(ctx *gin.Context) {
					sse.Encode(ctx.Writer, sse.Event{
						Event: "message",
						Data:  "some data\nmore data",
					})

					// also a complex type, like a map, a struct or a slice
					sse.Encode(ctx.Writer, sse.Event{
						Id:    "124",
						Event: "message",
						Data: map[string]interface{}{
							"user":    "manu",
							"date":    time.Now().Unix(),
							"content": "hi!",
						},
					})
				},
				Enabled: true,
				Desc:    "sse",
			},
		},
		"auth": {},
	}
	result := make([]MyRouter, 0, 64)

	for _, v := range routers {
		result = append(result, v...)
	}
	return result
}
