package secure

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/hash/md5"
	"github.com/jianlu8023/golang-example/cmd/tiny/internal/conf"
)

// LoginGet 接收 GET /login 请求,返回登录页面
func LoginGet(c *gin.Context) {
	c.HTML(http.StatusOK, "login.tpl", nil)
}

// LoginPost 接收 POST /login 请求,验证帐号密码,验证通过则生成 session
func LoginPost(c *gin.Context) {
	// 检查帐号密码
	// 通过则生成session，跳转首页
	// 不通过则返回登录页

	if c.PostForm("username") == md5.SumStringHex(conf.Config.Username) &&
		c.PostForm("password") == md5.SumStringHex(conf.Config.Password) {
		// session := sessions.Default(c)
		// session.Set("login", conf.Config.SessionVal)
		// session.Save()
		c.JSON(http.StatusOK, gin.H{"code": 1, "message": "登录成功"})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "登录失败"})
	}
}
