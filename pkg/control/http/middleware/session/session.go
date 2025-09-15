package session

import (
	// "github.com/gin-gonic/gin"
	// "github.com/gin-contrib/sessions/cookie"
	// "github.com/gin-contrib/sessions"
	// "github.com/gorilla/sessions"
	_ "fmt"
)

// https://segmentfault.com/a/1190000044565172

// EnableGinSession 启用gin的session功能
// func EnableGinSession(name string) gin.HandlerFunc {
// 	store := cookie.NewStore(sessionSecret)
// 	return sessions.Sessions(name, store)
// }

var (
// sessionSecret = []byte("sessionSecret")
// sessionMaxAge = 3600 * 24 * 7 // 7天
// sessionName   = "GSESSIONID"
// store         *sessions.CookieStore
)

// func init() {

// store = &sessions.CookieStore{
//	Codecs: securecookie.CodecsFromPairs(sessionSecret),
//	Options: &sessions.Options{
//		Path:     "/",
//		Secure:   true,
//		HttpOnly: true,
//		MaxAge:   sessionMaxAge,
//	},
// }

// store = sessions.NewCookieStore(sessionSecret)
// store.MaxAge(sessionMaxAge)
// }

// func EnableGorillaSession() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		if ctx.Request.URL.Path == "/login" {
// 			ctx.Next()
// 			return
// 		}
// 		_, err := store.Get(ctx.Request, sessionName)
// 		if err != nil {
// 			ctx.AbortWithStatusJSON(401, gin.H{
// 				"msg": "未登录",
// 			})
// 			return
// 		}
// 		ctx.Next()
// 	}
// }
