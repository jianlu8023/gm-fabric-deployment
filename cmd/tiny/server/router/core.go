package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/handle/core"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
	"github.com/jianlu8023/golang-example/cmd/tiny/server/middleware"
)

func loadCoreRoute(r *gin.Engine) *gin.RouterGroup {
	fileG := r.Group(conf.FileGroupPrefix, middleware.CheckLogin, middleware.CheckLevel)
	{
		fileG.GET("/*filename", core.Downloader)
		fileG.POST("/upload", core.Uploader)
	}
	return fileG
}
