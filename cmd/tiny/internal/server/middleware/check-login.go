package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
)

func CheckLogin(c *gin.Context) {
	// 检查 cookie，
	// 有则放行
	// 无则跳转登录页面
	if !conf.Config.IsSecure {
		return
	}

	// 使用cookie验证登录状态
	loginCookie, err := c.Cookie("login")
	if err != nil || loginCookie != conf.Config.SessionVal {
		// 对于API请求返回JSON错误，对于普通请求重定向到登录页
		if c.Request.Header.Get("Accept") == "application/json" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "请先登录"})
		} else {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		c.Abort()
		return
	}
	c.Next()
}
