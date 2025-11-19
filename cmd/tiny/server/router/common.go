package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/templates"
)

func load404Route(app *gin.Engine) {
	app.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "404 Page Not Found", nil)
	})
}

func loadIndexRoute(app *gin.Engine) {
	app.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/file/")
	})
}

func loadICORoute(app *gin.Engine) {
	app.GET("favicon.ico", func(c *gin.Context) {
		file, _ := templates.FS.ReadFile("favicon.ico")
		c.Data(http.StatusOK, "image/x-icon", file)
	})
}
