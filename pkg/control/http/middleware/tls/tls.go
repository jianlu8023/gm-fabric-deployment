package tls

import (
	// "github.com/gin-gonic/gin"
	// "github.com/jianlu8023/template/internal/addr"
	// "github.com/unrolled/secure"
	_ "fmt"
)


// https://liuqh.icu/2021/06/22/go/gin/integrated/2-zap/

// https://blog.csdn.net/qq_33591055/article/details/113955016

// https://blog.mulinbiao.com/post/gin-global-err/



// func EnableUnrolledTLS() gin.HandlerFunc {
// 	return func(ctx *gin.Context) {
// 		secureMiddleware := secure.New(secure.Options{
// 			SSLRedirect:           true,
// 			SSLHost:               addr.WebServiceAddr(),
// 			STSSeconds:            315360000,
// 			FrameDeny:             true,
// 			ContentSecurityPolicy: "default-src 'self'",
// 			IsDevelopment:         false,
// 		})
// 		err := secureMiddleware.Process(ctx.Writer, ctx.Request)
// 		if err != nil {
// 			return
// 		}
// 		ctx.Next()
// 	}
// }
