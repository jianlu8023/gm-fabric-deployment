package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"sync"
	"time"

	str "github.com/jianlu8023/go-tools/v2/pkg/helper/stringer"

	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/cors"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/gzip"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/requestid"
	"github.com/jianlu8023/golang-example/pkg/control/logger"

	"github.com/gin-gonic/gin"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"go.uber.org/zap"
)

type Control struct {
	serverConfig   *config.HttpServerConfig
	server         *http.Server
	ginRouter      *gin.Engine
	ctx            context.Context
	logger         *zap.SugaredLogger
	routers        []commonhttp.RouterHandler
	routerMutex    sync.RWMutex
	sessionManager jwt.SessionManager
	once           sync.Once
}

// NewWebServerControl 创建Web服务器控制器
// @param serverConfig *config.HttpServerConfig HTTP服务器配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control Web服务器控制器实例
func NewWebServerControl(serverConfig *config.HttpServerConfig, loggerControl *logger.Control) *Control {
	webLogger := loggerControl.GenLogger(logger.ModuleWeb)
	webLogger.Infof("[control] start new http server control...")
	gin.SetMode(serverConfig.RunMode)
	// 强制彩色输出
	gin.ForceConsoleColor()

	ctx := context.Background()

	webLogger.Debugf("[control] generate gin engine...")
	engine := gin.Default()
	webLogger.Debugf("[control] register gin middleware...")
	engine.Use(requestid.EnableRequestID(webLogger))
	engine.Use(cors.EnableCors())
	engine.Use(gzip.EnableGzip())

	// 调试模式，开启 pprof 包，便于开发阶段分析程序性能
	// gin.DefaultWriter = io.MultiWriter(os.Stdout, io.Discard)
	// gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, io.Discard)
	// engine = gin.Default()
	// 调试模式下开启pprof
	// pprof.Register(engine)

	// 创建会话管理器
	webLogger.Debugf("[control] create session manager...")
	sessionManager := jwt.NewMemorySessionManager(webLogger, ctx)

	// 注册JWT中间件
	// webLogger.Debugf("[control] register JWT middleware...")
	// engine.Use(middleware.EnableJWT(webLogger, sessionManager))

	// engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 	// 你的自定义格式
	// 127.0.0.1 - [2025-09-08 19:57:12.078] "GET /example/ping HTTP/2.0 200 50.066µs "curl/7.68.0" "
	// "%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 		param.ClientIP,
	// 		param.TimeStamp.Format("2006-01-02 15:04:05.000"),
	// 		param.Method,
	// 		param.Path,
	// 		param.Request.Proto,
	// 		param.StatusCode,
	// 		param.Latency,
	// 		param.Request.UserAgent(),
	// 		param.ErrorMessage,
	// 	)
	// }))
	webLogger.Debugf("[control] generate http server...")

	// 初始化TLS配置
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // 设置最低TLS版本
		MaxVersion: tls.VersionTLS13, // 设置最高TLS版本
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_AES_128_GCM_SHA256,       // tls 1.3
			tls.TLS_AES_256_GCM_SHA384,       // tls 1.3
			tls.TLS_CHACHA20_POLY1305_SHA256, // tls 1.3
		},
		CurvePreferences: []tls.CurveID{
			tls.CurveP256, tls.X25519,
		},
		PreferServerCipherSuites: true,  // 优先使用服务端加密套件
		SessionTicketsDisabled:   false, // 启用会话票据

		// SessionTickets:           &tls.SessionTicketMemoryStorage{}, // 使用内存存储会话票据
	}

	// 如果配置了根证书，则启用客户端证书验证
	rootCaCertFile := serverConfig.TlsRCACertFile
	if str.IsBlank(rootCaCertFile) {
		caCertPool := x509.NewCertPool()
		caCert, err := os.ReadFile(rootCaCertFile)
		if err == nil {
			if ok := caCertPool.AppendCertsFromPEM(caCert); ok {
				tlsConfig.ClientCAs = caCertPool
				tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven // 根据需要验证客户端证书
				webLogger.Debugf("[control] client certificate verification enabled with CA cert: %s", rootCaCertFile)
			}
		} else {
			webLogger.Warnf("[control] failed to read CA cert file: %v", err)
		}
	}

	srv := &http.Server{
		Addr:         serverConfig.Address,
		Handler:      engine,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,  // 设置读取超时
		WriteTimeout: 60 * time.Second,  // 设置写入超时
		IdleTimeout:  120 * time.Second, // 设置空闲超时
	}
	webLogger.Debugf("[control] generate http control...")
	control := &Control{
		serverConfig:   serverConfig,
		server:         srv,
		ctx:            ctx,
		logger:         webLogger,
		ginRouter:      engine,
		sessionManager: sessionManager,
	}

	// control.logger.Debugf("[control] register 404 405 handler...")
	// control.registerDefaultRouter()

	control.logger.Warnf("[control] current not setting router please call control.RegisterRouter to register router...")
	// 在这里不调用
	// control.initRouters()
	return control
}

// StartUp 启动HTTP服务器
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting up http server...")

		c.logger.Debugf("[control] starting define router...")
		c.initRouters()
		c.registerDefaultRouter()

		if c.serverConfig.TlsEnabled {
			c.logger.Infof("[control] start https server on %v", c.serverConfig.Address)
			go func() {
				if err := c.server.ListenAndServeTLS(c.serverConfig.TlsCertFile, c.serverConfig.TlsKeyFile); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
					if failedFunc != nil {
						failedFunc(err)
					}
				}
			}()
		} else {
			c.logger.Infof("[control] start http server on %v", c.serverConfig.Address)
			go func() {
				if err := c.server.ListenAndServe(); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
					if failedFunc != nil {
						failedFunc(err)
					}
				}
			}()
		}
	})
}

// Shutdown 关闭HTTP服务器
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	c.logger.Infof("[control] shutdown http server...")
	if err := c.server.Shutdown(c.ctx); err != nil {
		c.logger.Errorf("[control] shutdown http server failed: %v", err)
		return err
	}
	return nil
}

// registerDefaultRouter 注册默认路由处理器
// @description 注册404(路径不存在)和405(方法不允许)的默认处理函数
func (c *Control) registerDefaultRouter() {
	c.logger.Debugf("[control] register default router (NoRoute and NoMethod)...")
	{
		// 首先注册NoMethod处理器（方法不允许）
		// NoMethod应该在NoRoute之前注册，以确保当路径存在但方法不支持时能正确返回405
		c.ginRouter.NoMethod(func(ctx *gin.Context) {
			c.logger.Warnf("[control] 405 Method Not Allowed: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
			ctx.JSON(http.StatusMethodNotAllowed, gin.H{
				"code":    http.StatusMethodNotAllowed,
				"message": "Method Not Allowed",
				"data":    nil,
			})
		})
	}

	{
		// 然后注册NoRoute处理器（路径不存在）
		c.ginRouter.NoRoute(func(ctx *gin.Context) {
			c.logger.Warnf("[control] 404 Not Found: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
			ctx.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "Not Found",
				"data":    nil,
			})

		})
	}

	{
		// 定义路由
		allRouterUri := fmt.Sprintf("%s/%s", c.serverConfig.ContextPath, "routers")

		c.ginRouter.GET(allRouterUri, func(ctx *gin.Context) {
			commonhttp.SuccessResponse(ctx, gin.H{
				"routers": c.routers,
			})
		})
	}

	// 如果启用了pprof，则注册pprof路由
	if c.serverConfig.Pprof {
		c.logger.Infof("[control] pprof enabled, registering pprof routes")
		pprofUri := fmt.Sprintf("%s/%s", c.serverConfig.ContextPath, "debug/pprof")
		pprofRouter := c.ginRouter.Group(pprofUri)
		{
			pprofRouter.GET("/", gin.WrapF(pprof.Index))
			pprofRouter.GET("/cmdline", gin.WrapF(pprof.Cmdline))
			pprofRouter.GET("/profile", gin.WrapF(pprof.Profile))
			pprofRouter.POST("/symbol", gin.WrapF(pprof.Symbol))
			pprofRouter.GET("/symbol", gin.WrapF(pprof.Symbol))
			pprofRouter.GET("/trace", gin.WrapF(pprof.Trace))
			pprofRouter.GET("/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))
			pprofRouter.GET("/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
			pprofRouter.GET("/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
			pprofRouter.GET("/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
			pprofRouter.GET("/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
			pprofRouter.GET("/threadcreate", gin.WrapF(pprof.Handler("threadcreate").ServeHTTP))
		}
	}

	c.logger.Debugf("[control] default router registered successfully")
}

// GetSessionManager 获取会话管理器
// @return jwt.SessionManager 会话管理器
func (c *Control) GetSessionManager() jwt.SessionManager {
	return c.sessionManager
}

// RegisterRouter 注册HTTP路由
// @param routers []commonhttp.RouterHandler 路由处理器列表
func (c *Control) RegisterRouter(routers []commonhttp.RouterHandler) {
	c.logger.Infof("[control] register router...")
	c.routerMutex.Lock()
	c.routers = append(c.routers, routers...)
	c.routerMutex.Unlock()
	c.logger.Debugf("[control] deduplicate routers...")
	c.deduplicateRouters()

	c.logger.Infof("[control] register router success...")
}

func (c *Control) deduplicateRouters() {
	c.logger.Debugf("[control] starting deduplicate routers...")
	c.routerMutex.RLock()
	defer c.routerMutex.RUnlock()
	after := make([]commonhttp.RouterHandler, 0, len(c.routers))
	seen := make(map[string]struct{}, len(c.routers))

	for _, router := range c.routers {
		if _, ok := seen[router.GetUri()]; ok {
			// 已经存在
			continue
		}
		// 还不存在
		seen[router.GetUri()] = struct{}{}
		after = append(after, router)
	}
	c.logger.Debugf("[control] finished deduplicate routers...")
	c.routers = after
}

// initRouters 初始化所有注册的路由
func (c *Control) initRouters() {
	c.logger.Infof("[control] start init routers...")

	// 预创建认证和非认证的路由组，避免每次循环重新创建
	authGroup := c.ginRouter.Group("/", jwt.EnableJWT(c.logger, c.sessionManager))
	noAuthGroup := c.ginRouter.Group("/")

	for _, router := range c.routers {
		if !router.IsEnabled() {
			continue // 跳过禁用的路由
		}

		// 构建URL路径
		var url string
		if strings.HasPrefix(router.GetUri(), "/") {
			// 避免重复添加前缀
			url = fmt.Sprintf("%s%s", c.serverConfig.ContextPath, router.GetUri())
		} else {
			url = fmt.Sprintf("%s/%s", c.serverConfig.ContextPath, router.GetUri())
		}

		handlerFunc := router.GetHandlerFunc()
		httpMethod := router.GetMethod()

		c.logger.Debugf("[control] register router uri %s method %s", url, httpMethod)

		// 根据是否需要JWT验证和HTTP方法选择合适的路由组和注册方法
		if router.GetEnableJWtVerify() {
			switch httpMethod {
			case http.MethodGet:
				authGroup.GET(url, handlerFunc)
			case http.MethodPost:
				authGroup.POST(url, handlerFunc)
			case http.MethodPut:
				authGroup.PUT(url, handlerFunc)
			case http.MethodDelete:
				authGroup.DELETE(url, handlerFunc)
			case http.MethodPatch:
				authGroup.PATCH(url, handlerFunc)
			case http.MethodOptions:
				authGroup.OPTIONS(url, handlerFunc)
			case http.MethodHead:
				authGroup.HEAD(url, handlerFunc)
			default:
				// 默认使用GET方法
				authGroup.GET(url, handlerFunc)
			}
		} else {
			switch httpMethod {
			case http.MethodGet:
				noAuthGroup.GET(url, handlerFunc)
			case http.MethodPost:
				noAuthGroup.POST(url, handlerFunc)
			case http.MethodPut:
				noAuthGroup.PUT(url, handlerFunc)
			case http.MethodDelete:
				noAuthGroup.DELETE(url, handlerFunc)
			case http.MethodPatch:
				noAuthGroup.PATCH(url, handlerFunc)
			case http.MethodOptions:
				noAuthGroup.OPTIONS(url, handlerFunc)
			case http.MethodHead:
				noAuthGroup.HEAD(url, handlerFunc)
			default:
				// 默认使用GET方法
				noAuthGroup.GET(url, handlerFunc)
			}
		}
	}
}
