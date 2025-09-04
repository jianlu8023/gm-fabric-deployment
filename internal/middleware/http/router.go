package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
					ctx.JSON(http.StatusOK, gin.H{
						"msg": "pong",
					})
				},
				Enabled: true,
				Desc:    "ping",
			},
		},
	}
	result := make([]MyRouter, 0, 64)

	for _, v := range routers {
		result = append(result, v...)
	}
	return result
}
