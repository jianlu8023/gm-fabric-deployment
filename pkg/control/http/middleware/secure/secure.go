package secure

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/unrolled/secure"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// EnableTLSProtection 创建并返回TLS保护中间件
// @description 提供TLS安全相关的HTTP头设置，增强Web应用安全性
// @param logger 日志记录器
// @param isDevelopment 是否为开发环境
// @param sslRedirect 是否启用SSL重定向
// @return gin.HandlerFunc Gin中间件函数
func EnableTLSProtection(logger *zap.SugaredLogger, isDevelopment bool, sslRedirect bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "secure")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 设置HSTS头，强制客户端使用HTTPS
		if !isDevelopment && sslRedirect {
			// 315360000秒 = 10年
			ctx.Header("Strict-Transport-Security", "max-age=315360000; includeSubDomains")
		}

		// 设置X-Frame-Options，防止点击劫持
		ctx.Header("X-Frame-Options", "DENY")

		// 设置X-Content-Type-Options，防止MIME类型嗅探
		ctx.Header("X-Content-Type-Options", "nosniff")

		// 设置X-XSS-Protection，启用XSS过滤
		ctx.Header("X-XSS-Protection", "1; mode=block")

		// 设置内容安全策略(CSP)
		// 这里使用相对宽松的策略，实际应用中应根据需求进行调整
		ctx.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")

		// 对于非GET请求，记录TLS连接信息
		if ctx.Request.Method != http.MethodGet {
			if ctx.Request.TLS != nil {
				logger.Debugf("[TLS] Connection from %s using TLS version %s, cipher suite: %04x",
					ctx.ClientIP(),
					ctx.Request.TLS.Version,
					ctx.Request.TLS.CipherSuite,
				)
			}
		}

		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}

// RedirectToHTTPS 创建并返回HTTP重定向到HTTPS的中间件
// @description 将所有HTTP请求重定向到HTTPS
// @param httpsPort HTTPS服务端口
// @param sslRedirect 是否启用SSL重定向
// @return gin.HandlerFunc Gin中间件函数
func RedirectToHTTPS(httpsPort int, sslRedirect bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "secure")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 检查是否为HTTPS连接，并且启用了SSL重定向
		if ctx.Request.TLS == nil && sslRedirect {
			// 构建HTTPS URL
			host := ctx.Request.Host
			// 如果主机名包含端口号，替换为HTTPS端口
			if _, _, err := net.SplitHostPort(host); err == nil {
				host = strings.Split(host, ":")[0]
			}
			newURL := fmt.Sprintf("https://%s:%d%s", host, httpsPort, ctx.Request.RequestURI)

			// 301永久重定向到HTTPS
			ctx.Redirect(http.StatusMovedPermanently, newURL)
			ctx.Abort()
			return
		}
		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}

// EnableUnrolledTLS 实现类似unrolled/secure包的TLS安全中间件功能
// @description 提供与unrolled/secure包类似的TLS安全增强功能，但不依赖外部包
// @param logger 日志记录器
// @param isDevelopment 是否为开发环境
// @param sslHost SSL主机名
// @param sslRedirect 是否启用SSL重定向
// @return gin.HandlerFunc Gin中间件函数
func EnableUnrolledTLS(logger *zap.SugaredLogger, isDevelopment bool, sslHost string, sslRedirect bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "secure")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)
		// 设置HSTS头，强制客户端使用HTTPS
		if !isDevelopment && sslRedirect {
			// 315360000秒 = 10年
			ctx.Header("Strict-Transport-Security", "max-age=315360000; includeSubDomains")
		}

		// 设置X-Frame-Options，防止点击劫持
		ctx.Header("X-Frame-Options", "DENY")

		// 设置X-Content-Type-Options，防止MIME类型嗅探
		ctx.Header("X-Content-Type-Options", "nosniff")

		// 设置X-XSS-Protection，启用XSS过滤
		ctx.Header("X-XSS-Protection", "1; mode=block")

		// 设置内容安全策略(CSP)
		ctx.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")

		// 记录TLS连接信息
		if ctx.Request.TLS != nil && logger != nil {
			logger.Debugf("[TLS] Connection to %s from %s using TLS version %s",
				sslHost,
				ctx.ClientIP(),
				ctx.Request.TLS.Version,
			)
		}

		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}

// EnableSecurePackageTLS 使用unrolled/secure包实现TLS安全中间件
// @description 使用成熟的unrolled/secure包提供TLS安全增强功能
// @param sslHost SSL主机名
// @param isDevelopment 是否为开发环境
// @param sslRedirect 是否启用SSL重定向
// @return gin.HandlerFunc Gin中间件函数
func EnableSecurePackageTLS(sslHost string, isDevelopment bool, sslRedirect bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		savedCtx := ctx.Request.Context()
		defer func() {
			ctx.Request = ctx.Request.WithContext(savedCtx)
		}()
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "secure")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)

		// 处理sslHost，确保正确解析主机名和端口
		host := ctx.Request.Host

		// 如果sslHost为空，则直接使用请求的主机
		if stringer.IsBlank(sslHost) {
			sslHost = host
		} else {
			// 解析当前请求的主机名和端口
			requestHost, requestPort, err := net.SplitHostPort(host)
			if err != nil {
				// 如果无法分割，说明没有端口号或者格式异常
				requestHost = host
				requestPort = ""
			}

			// 解析sslHost参数
			sslHostWithoutPort, sslPort, sslErr := net.SplitHostPort(sslHost)
			if sslErr != nil {
				// 如果sslHost无法分割，说明不包含端口号
				sslHostWithoutPort = sslHost
				sslPort = ""
			}

			// 如果sslHost只有端口号（如":8080"），则使用请求的主机名
			if stringer.IsBlank(sslHostWithoutPort) && !stringer.IsBlank(sslPort) {
				if !stringer.IsBlank(requestHost) {
					sslHost = net.JoinHostPort(requestHost, sslPort)
				}
			}
			// 如果sslHost有主机名但没有端口，而请求有端口，则保留请求的端口
			if !stringer.IsBlank(sslHostWithoutPort) &&
				stringer.IsBlank(sslPort) &&
				!stringer.IsBlank(requestPort) {
				sslHost = net.JoinHostPort(sslHostWithoutPort, requestPort)
			}
		}

		secureMiddleware := secure.New(secure.Options{
			SSLRedirect:           sslRedirect,
			SSLHost:               sslHost,
			STSSeconds:            315360000,
			FrameDeny:             true,
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:",
			IsDevelopment:         isDevelopment,
		})
		// 如果SSL重定向已经发送，不再继续处理请求
		if err := secureMiddleware.Process(ctx.Writer, ctx.Request); err != nil {
			ctx.Abort()
			span.SetStatus(codes.Error, err.Error())
			return
		}
		ctx.Next()
		span.SetStatus(codes.Ok, "success")
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
