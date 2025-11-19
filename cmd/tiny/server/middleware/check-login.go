package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
)

func CheckLogin(c *gin.Context) {
	// 检查 session，
	// 有则放行
	// 无则跳转登录页面
	if !conf.Config.IsSecure {
		return
	}

	// if session := sessions.Default(c); session.Get("login") == conf.Config.SessionVal {
	// 	c.Next()
	// 	return
	// } else {
	// 	c.Redirect(http.StatusTemporaryRedirect, "/login")
	// 	return
	// }
	
}
