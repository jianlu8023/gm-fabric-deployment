package secure

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
	"go.uber.org/zap"
)

// EnableTLSProtection 创建并返回TLS保护中间件
// @description 提供TLS安全相关的HTTP头设置，增强Web应用安全性
// @param logger 日志记录器
// @param isDevelopment 是否为开发环境
// @return gin.HandlerFunc Gin中间件函数
func EnableTLSProtection(logger *zap.SugaredLogger, isDevelopment bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置HSTS头，强制客户端使用HTTPS
		if !isDevelopment {
			// 315360000秒 = 10年
			c.Header("Strict-Transport-Security", "max-age=315360000; includeSubDomains")
		}

		// 设置X-Frame-Options，防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 设置X-Content-Type-Options，防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 设置X-XSS-Protection，启用XSS过滤
		c.Header("X-XSS-Protection", "1; mode=block")

		// 设置内容安全策略(CSP)
		// 这里使用相对宽松的策略，实际应用中应根据需求进行调整
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")

		// 对于非GET请求，记录TLS连接信息
		if c.Request.Method != http.MethodGet {
			if c.Request.TLS != nil {
				logger.Debugf("[TLS] Connection from %s using TLS version %s, cipher suite: %04x",
					c.ClientIP(),
					c.Request.TLS.Version,
					c.Request.TLS.CipherSuite,
				)
			}
		}

		c.Next()
	}
}

// RedirectToHTTPS 创建并返回HTTP重定向到HTTPS的中间件
// @description 将所有HTTP请求重定向到HTTPS
// @param httpsPort HTTPS服务端口
// @return gin.HandlerFunc Gin中间件函数
func RedirectToHTTPS(httpsPort int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为HTTPS连接
		if c.Request.TLS == nil {
			// 构建HTTPS URL
			host := c.Request.Host
			// 如果主机名包含端口号，替换为HTTPS端口
			if _, _, err := net.SplitHostPort(host); err == nil {
				host = strings.Split(host, ":")[0]
			}
			newURL := fmt.Sprintf("https://%s:%d%s", host, httpsPort, c.Request.RequestURI)

			// 301永久重定向到HTTPS
			c.Redirect(http.StatusMovedPermanently, newURL)
			c.Abort()
			return
		}
		c.Next()
	}
}

// EnableUnrolledTLS 实现类似unrolled/secure包的TLS安全中间件功能
// @description 提供与unrolled/secure包类似的TLS安全增强功能，但不依赖外部包
// @param logger 日志记录器
// @param isDevelopment 是否为开发环境
// @param sslHost SSL主机名
// @return gin.HandlerFunc Gin中间件函数
func EnableUnrolledTLS(logger *zap.SugaredLogger, isDevelopment bool, sslHost string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置HSTS头，强制客户端使用HTTPS
		if !isDevelopment {
			// 315360000秒 = 10年
			c.Header("Strict-Transport-Security", "max-age=315360000; includeSubDomains")
		}

		// 设置X-Frame-Options，防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 设置X-Content-Type-Options，防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 设置X-XSS-Protection，启用XSS过滤
		c.Header("X-XSS-Protection", "1; mode=block")

		// 设置内容安全策略(CSP)
		c.Header("Content-Security-Policy", "default-src 'self'")

		// 记录TLS连接信息
		if c.Request.TLS != nil && logger != nil {
			logger.Debugf("[TLS] Connection to %s from %s using TLS version %s",
				sslHost,
				c.ClientIP(),
				c.Request.TLS.Version,
			)
		}

		c.Next()
	}
}

// EnableSecurePackageTLS 使用unrolled/secure包实现TLS安全中间件
// @description 使用成熟的unrolled/secure包提供TLS安全增强功能
// @param sslHost SSL主机名
// @param isDevelopment 是否为开发环境
// @return gin.HandlerFunc Gin中间件函数
func EnableSecurePackageTLS(sslHost string, isDevelopment bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect:           true,
			SSLHost:               sslHost,
			STSSeconds:            315360000,
			FrameDeny:             true,
			ContentSecurityPolicy: "default-src 'self'",
			IsDevelopment:         isDevelopment,
		})
		err := secureMiddleware.Process(ctx.Writer, ctx.Request)
		// 如果SSL重定向已经发送，不再继续处理请求
		if err != nil {
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

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
