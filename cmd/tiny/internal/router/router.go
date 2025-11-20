package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/templates"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

func GenRouters() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		&commonhttp.MyRouter{
			Name:   "/",
			Uri:    "/",
			Method: http.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				ctx.Redirect(http.StatusPermanentRedirect, "/file/")
			},
			Enabled:         true,
			Desc:            "index",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:   "favicon.ico",
			Uri:    "/favicon.ico",
			Method: http.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {
				file, _ := templates.FS.ReadFile("favicon.ico")
				ctx.Data(http.StatusOK, "image/x-icon", file)
			},
			Enabled:         true,
			Desc:            "favicon.ico",
			EnableJWtVerify: false,
		},
		&commonhttp.MyRouter{
			Name:   "/file/*filename",
			Uri:    "/file/*filename",
			Method: http.MethodGet,
			HandlerFunc: func(ctx *gin.Context) {

			},
			Enabled:         true,
			EnableJWtVerify: false,
		},
	}
}
