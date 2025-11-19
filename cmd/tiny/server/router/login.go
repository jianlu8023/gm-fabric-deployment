package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/handle/secure"
)

func loadLoginRoute(r *gin.Engine) {
	r.GET("/login", secure.LoginGet)
	r.POST("/login", secure.LoginPost)
}
