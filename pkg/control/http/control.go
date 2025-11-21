package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	// "gitee.com/zhaochuninhefei/gmgo/gmtls"
	// gmx509 "gitee.com/zhaochuninhefei/gmgo/x509"
	// "github.com/hxx258456/ccgo/gmtls"
	// gmx509 "github.com/hxx258456/ccgo/x509"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/cors"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/gzip"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/ipblacklist"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/ipwhitelist"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/language"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/ratelimit"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/recovery"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/requestid"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/secure"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/tjfoc/gmsm/gmtls"
	gmx509 "github.com/tjfoc/gmsm/x509"

	"github.com/gin-gonic/gin"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"go.uber.org/zap"
)

type Control struct {
	config         *config.HttpServerConfig
	server         *http.Server
	ginRouter      *gin.Engine
	tlsConfig      *tls.Config
	gmTlsConfig    *gmtls.Config
	ctx            context.Context
	logger         *zap.SugaredLogger
	sessionManager jwt.SessionManager
	once           sync.Once
	tracerControl  *tracer.Control
	routerGroups   concurrent.Map[string, commonhttp.GroupRouterHandler] // 使用 并发安全的 map
	// routerGroups   map[string]commonhttp.GroupRouterHandler
	// routerMutex    sync.RWMutex
}

// NewWebServerControl 创建Web服务器控制器
// @param config *config.HttpServerConfig HTTP服务器配置
// @param loggerControl *logger.Control 日志控制器
// @param opts ...Option 可选的配置选项
// @return *Control Web服务器控制器实例
func NewWebServerControl(serverConfig *config.HttpServerConfig, loggerControl *logger.Control, opts ...Option) (*Control, error) {
	if serverConfig == nil {
		serverConfig = getDefaultConfig()
	}
	if !serverConfig.Enabled {
		return nil, errors.New("http is not enabled")
	}

	webLogger := loggerControl.GenLogger(logger.ModuleWeb)
	webLogger.Info("[control] start new http server control...")
	gin.SetMode(serverConfig.RunMode)
	// 强制彩色输出
	gin.ForceConsoleColor()

	ctx := context.Background()

	webLogger.Debug("[control] generate gin engine...")
	engine := gin.Default()

	srv := &http.Server{
		Addr:         serverConfig.Address,
		Handler:      engine.Handler(),
		ReadTimeout:  30 * time.Second,  // 设置读取超时
		WriteTimeout: 60 * time.Second,  // 设置写入超时
		IdleTimeout:  120 * time.Second, // 设置空闲超时
	}

	// 如果启用HTTP/2且非TLS模式，使用h2c支持HTTP/2 over cleartext
	if serverConfig.Http2Enabled && !serverConfig.TlsEnabled {
		webLogger.Info("[control] HTTP/2 enabled for cleartext connections (h2c)")
		h2s := &http2.Server{}
		srv.Handler = h2c.NewHandler(engine, h2s)
	}

	webLogger.Debug("[control] generate http control...")
	control := &Control{
		config:       serverConfig,
		server:       srv,
		ctx:          ctx,
		logger:       webLogger,
		ginRouter:    engine,
		routerGroups: concurrentmap.NewRWMap[string, commonhttp.GroupRouterHandler](),
		// routerGroups: make(map[string]commonhttp.GroupRouterHandler),
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(control)
	}

	if control.sessionManager == nil {
		// 创建会话管理器
		webLogger.Debug("[control] create session manager...")
		control.sessionManager = jwt.NewMemorySessionManager(webLogger, ctx)
	}

	webLogger.Info("[control] register gin middleware...")

	// 0. Tracer中间件 - 用于请求追踪
	if control.tracerControl != nil {
		engine.Use(otelgin.Middleware(control.tracerControl.GetServiceName(), otelgin.WithTracerProvider(control.tracerControl.TracerProvider())))
	}

	// 1. 恢复中间件（Recovery Middleware）- 应在最前面注册，捕获所有后续中间件的panic
	engine.Use(recovery.EnableRecovery(control.logger, true))

	// 添加语言支持中间件
	engine.Use(language.EnableLanguageSupport(serverConfig.Language))

	// 2. 请求ID中间件 - 为每个请求生成唯一标识
	engine.Use(requestid.EnableRequestID(webLogger))

	// 3. IP白名单中间件（如果启用）- 尽早过滤非白名单IP
	if serverConfig.IPWhiteList.Enabled && len(serverConfig.IPWhiteList.IPs) > 0 {
		webLogger.Debugf("[control] register IP white list middleware with %d IPs", len(serverConfig.IPWhiteList.IPs))
		engine.Use(ipwhitelist.EnableIPWhiteList(webLogger, serverConfig.IPWhiteList.IPs))
	}

	// 4. IP黑名单中间件（如果启用）- 尽早拒绝黑名单IP
	if serverConfig.IPBlackList.Enabled && len(serverConfig.IPBlackList.IPs) > 0 {
		webLogger.Debugf("[control] register IP black list middleware with %d IPs", len(serverConfig.IPBlackList.IPs))
		engine.Use(ipblacklist.EnableIPBlackList(webLogger, serverConfig.IPBlackList.IPs))
	}

	if !serverConfig.TlsGM {
		// TODO gm模式下 会出现一直301的情况
		// 6. TLS安全中间件 - 安全检查，在基础过滤和追踪后执行
		// 判断是否为开发环境（根据Gin模式）
		isDevelopment := gin.Mode() == gin.DebugMode
		// 使用成熟的unrolled/secure包实现的TLS安全中间件
		// 推荐在生产环境使用，提供完整的TLS安全保护功能
		engine.Use(secure.EnableSecurePackageTLS(serverConfig.Address, isDevelopment))
	}

	// 如果需要使用不依赖外部包的版本，可以取消注释下面这行
	// engine.Use(secure.EnableUnrolledTLS(webLogger, isDevelopment, serverConfig.Address))

	// 7. CORS中间件 - 跨域处理
	engine.Use(cors.EnableCors())

	// 8. 限流中间件（如果启用）- 在业务逻辑前执行
	if serverConfig.RateLimit.Enabled && serverConfig.RateLimit.RPS > 0 {
		// 创建限流配置
		rateLimitConfig := ratelimit.Config{
			Type:  ratelimit.Type(serverConfig.RateLimit.Type), // 使用配置文件中的限流类型
			RPS:   serverConfig.RateLimit.RPS,
			Burst: serverConfig.RateLimit.Burst,
		}
		// 使用工厂函数创建限流中间件
		engine.Use(ratelimit.NewRateLimitMiddleware(webLogger, rateLimitConfig))
	}

	// 9. 压缩中间件 - 性能优化，在响应前执行
	engine.Use(gzip.EnableGzip())

	// 调试模式，开启 pprof 包，便于开发阶段分析程序性能
	// gin.DefaultWriter = io.MultiWriter(os.Stdout, io.Discard)
	// gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, io.Discard)
	// engine = gin.Default()
	// 调试模式下开启pprof
	// pprof.Register(engine)

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

	if serverConfig.TlsEnabled {
		if serverConfig.TlsGM {
			// 配置 gm secure
			gmTLSConfig := &gmtls.Config{
				GMSupport: &gmtls.GMSupport{
					WorkMode: gmtls.ModeGMSSLOnly,
				},
				// MinVersion: gmtls.VersionTLS12,
				// MaxVersion: gmtls.VersionGMSSL,
				// CipherSuites: []uint16{
				// 	gmtls.GMTLS_SM2_WITH_SM1_SM3,
				// 	gmtls.GMTLS_SM2_WITH_SM4_SM3,
				// 	gmtls.GMTLS_ECDHE_SM2_WITH_SM4_SM3,
				// 	gmtls.GMTLS_ECDHE_SM2_WITH_SM1_SM3,
				// 	gmtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				// 	gmtls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				// 	gmtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				// 	gmtls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				// },
				// SessionTicketsDisabled: false, // 启用会话票据
			}
			// TODO 目前 gmhserver.key 是加密的key 在使用 emmansun/gmsm 解密后 出现 secure: sm2 private key does not match public key
			//
			// certificates, err := gmtls.LoadX509KeyPair(serverConfig.TlsCertFile, serverConfig.TlsKeyFile)
			// if err != nil {
			// 	webLogger.Errorf("[control] failed to load GM TLS certificate: %v", err)
			// 	return nil, err
			// } else {
			// 	gmTLSConfig.Certificates = []gmtls.Certificate{certificates}
			// }

			// TODO 目前 没找到支持http2的方法 暂时注释掉
			// 根据HTTP/2配置决定是否启用HTTP/2协议协商
			// if serverConfig.Http2Enabled {
			// 	gmTLSConfig.NextProtos = []string{"h2", "http/1.1"}
			// 	webLogger.Infof("[control] HTTP/2 enabled for TLS connections")
			// } else {
			// 	gmTLSConfig.NextProtos = []string{"http/1.1"}
			// 	webLogger.Infof("[control] HTTP/2 disabled, using HTTP/1.1 only")
			// }

			// GM模式需要两套keypair：一个签名，一个加密
			// 使用逗号分割证书和密钥文件路径
			certFiles := strings.Split(serverConfig.TlsCertFile, ",")
			keyFiles := strings.Split(serverConfig.TlsKeyFile, ",")

			// 检查是否提供了两套证书和密钥
			if len(certFiles) != 2 || len(keyFiles) != 2 {
				webLogger.Errorf("[control] GM模式必须提供两套keypair（签名和加密），当前证书文件数量: %d, 密钥文件数量: %d", len(certFiles), len(keyFiles))
				return nil, errors.New("GM模式必须提供两套keypair，请在 tls_cert_file 和 tls_key_file 中使用逗号分割两套证书和密钥文件路径")
			}

			// 去除文件路径的空格
			for i := range certFiles {
				certFiles[i] = strings.TrimSpace(certFiles[i])
			}
			for i := range keyFiles {
				keyFiles[i] = strings.TrimSpace(keyFiles[i])
			}

			// 验证文件路径不为空
			for i, file := range certFiles {
				if stringer.IsBlank(file) {
					webLogger.Errorf("[control] 第%d个证书文件路径为空", i+1)
					return nil, fmt.Errorf("第%d个证书文件路径为空", i+1)
				}
			}
			for i, file := range keyFiles {
				if stringer.IsBlank(file) {
					webLogger.Errorf("[control] 第%d个密钥文件路径为空", i+1)
					return nil, fmt.Errorf("第%d个密钥文件路径为空", i+1)
				}
			}

			// 加载两套keypair
			var certificates []gmtls.Certificate
			for i := 0; i < 2; i++ {
				cert, err := gmtls.LoadX509KeyPair(certFiles[i], keyFiles[i])
				if err != nil {
					webLogger.Errorf("[control] 加载第%d套GM TLS证书失败: %v", i+1, err)
					return nil, fmt.Errorf("加载第%d套GM TLS证书失败: %v", i+1, err)
				}
				certificates = append(certificates, cert)
				webLogger.Debugf("[control] 成功加载第%d套GM TLS证书: %s -> %s", i+1, certFiles[i], keyFiles[i])
			}

			// 设置证书到GM TLS配置
			gmTLSConfig.Certificates = certificates
			webLogger.Infof("[control] 成功加载GM模式两套keypair，签名证书: %s，加密证书: %s", certFiles[0], certFiles[1])

			// 如果配置了根证书，则启用客户端证书验证
			rootCaCertFile := serverConfig.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
				caCertPool := gmx509.NewCertPool()
				caCert, err := os.ReadFile(rootCaCertFile)
				if err == nil {
					if ok := caCertPool.AppendCertsFromPEM(caCert); ok {
						gmTLSConfig.ClientCAs = caCertPool
						gmTLSConfig.ClientAuth = gmtls.VerifyClientCertIfGiven // 根据需要验证客户端证书
						webLogger.Debugf("[control] client certificate verification enabled with GM CA cert: %s", rootCaCertFile)
					}
				} else {
					webLogger.Warnf("[control] failed to read GM CA cert file: %v", err)
				}
			}

			control.gmTlsConfig = gmTLSConfig
		} else {
			// 配置	标准tls
			tlsConfig := &tls.Config{
				MinVersion: tls.VersionTLS12, // 设置最低TLS版本
				MaxVersion: tls.VersionTLS13, // 设置最高TLS版本
				CipherSuites: []uint16{
					// tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					// tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256, // secure 1.3
					tls.TLS_AES_128_GCM_SHA256,       // secure 1.3
					tls.TLS_AES_256_GCM_SHA384,       // secure 1.3
				},
				CurvePreferences: []tls.CurveID{
					tls.X25519, // 优先使用X25519椭圆曲线
					tls.CurveP256,
				},
				// PreferServerCipherSuites: true,  // 优先使用服务端加密套件
				SessionTicketsDisabled: false, // 启用会话票据
			}

			// 根据HTTP/2配置决定是否启用HTTP/2协议协商
			if serverConfig.Http2Enabled {
				tlsConfig.NextProtos = []string{"h2", "http/1.1"}
				webLogger.Infof("[control] HTTP/2 enabled for TLS connections")
			} else {
				tlsConfig.NextProtos = []string{"http/1.1"}
				webLogger.Infof("[control] HTTP/2 disabled, using HTTP/1.1 only")
			}
			certificates, err := tls.LoadX509KeyPair(serverConfig.TlsCertFile, serverConfig.TlsKeyFile)
			if err != nil {
				webLogger.Errorf("[control] failed to load TLS certificate: %v", err)
				return nil, err
			} else {
				tlsConfig.Certificates = []tls.Certificate{certificates}
			}
			// 如果配置了根证书，则启用客户端证书验证
			rootCaCertFile := serverConfig.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
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
			control.tlsConfig = tlsConfig
		}
	}

	// control.logger.Debugf("[control] register 404 405 handler...")
	// control.registerDefaultRouter()

	control.logger.Warnf("[control] current not setting router please call control.RegisterRouter to register router...")
	// 在这里不调用
	// control.initRouters()
	return control, nil
}

// StartUp 启动HTTP服务器
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Info("[control] starting up http server...")

			c.logger.Debug("[control] starting define router...")
			c.registerDefaultRouter()
			c.initRouters()

			if c.config.TlsEnabled {
				if c.config.TlsGM {
					c.logger.Infof("[control] start gm https server on %v", c.config.Address)
					go func() {
						listener, err := gmtls.Listen("tcp", c.config.Address, c.gmTlsConfig)
						if err != nil {
							c.logger.Errorf("[control] failed to create gm TLS listener: %v", err)
							if failedFunc != nil {
								failedFunc(err)
							}
							return
						}

						// 注意：GM TLS不支持HTTP/2，因为HTTP/2需要的ALPN协议协商和GM TLS不兼容
						if c.config.Http2Enabled {
							c.logger.Warnf("[control] HTTP/2 is not supported with GM TLS, falling back to HTTP/1.1")
							// if err := http2.ConfigureServer(c.server, &http2.Server{}); err != nil {
							// 	c.logger.Errorf("[control] failed to configure HTTP/2 server: %v", err)
							// 	if failedFunc != nil {
							// 		failedFunc(err)
							// 	}
							// 	return
							// }
							// c.logger.Info("[control] HTTP/2 server configured successfully")
						}

						defer func(listener net.Listener) {
							if err := listener.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
								c.logger.Errorf("[control] failed to close gm TLS listener: %v", err)
							}
						}(listener)
						// 使用srv.Serve启动服务器
						if err := c.server.Serve(listener); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
							c.logger.Errorf("[control] HTTPS server error: %v", err)
							if failedFunc != nil {
								failedFunc(err)
							}
							return
						}
					}()
				} else {
					c.logger.Infof("[control] start https server on %v", c.config.Address)
					go func() {

						// 使用tls.Listen创建监听器
						listener, err := tls.Listen("tcp", c.config.Address, c.tlsConfig)
						if err != nil {
							c.logger.Errorf("[control] failed to create TLS listener: %v", err)
							if failedFunc != nil {
								failedFunc(err)
							}
							return
						}

						// 为HTTPS服务器启用HTTP/2支持
						if c.config.Http2Enabled {
							if err := http2.ConfigureServer(c.server, &http2.Server{}); err != nil {
								c.logger.Errorf("[control] failed to configure HTTP/2 server: %v", err)
								if failedFunc != nil {
									failedFunc(err)
								}
								return
							}
							c.logger.Info("[control] HTTP/2 server configured successfully")
						}

						defer func(listener net.Listener) {
							if err := listener.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
								c.logger.Errorf("[control] failed to close TLS listener: %v", err)
							}
						}(listener)

						// 使用srv.Serve启动服务器
						if err := c.server.Serve(listener); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
							c.logger.Errorf("[control] HTTPS server error: %v", err)
							if failedFunc != nil {
								failedFunc(err)
							}
							return
						}
					}()
				}

			} else {
				c.logger.Infof("[control] start http server on %v", c.config.Address)
				go func() {
					// 使用普通TCP监听器
					listener, err := net.Listen("tcp", c.config.Address)
					if err != nil {
						c.logger.Errorf("[control] failed to create TCP listener: %v", err)
						if failedFunc != nil {
							failedFunc(err)
						}
						return
					}
					defer func(listener net.Listener) {
						if err := listener.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
							c.logger.Errorf("[control] failed to close TCP listener: %v", err)
						}
					}(listener)

					// 使用srv.Serve启动服务器
					if err := c.server.Serve(listener); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
						c.logger.Errorf("[control] HTTP server error: %v", err)
						if failedFunc != nil {
							failedFunc(err)
						}
						return
					}
				}()
			}
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
	if err := c.ClearUploadCache(); err != nil {
		c.logger.Errorf("[control] failed to clear upload cache: %v", err)
		return err
	}
	return nil
}

// registerDefaultRouter 注册默认路由处理器
// @description 注册404(路径不存在)和405(方法不允许)的默认处理函数
func (c *Control) registerDefaultRouter() {
	c.logger.Debug("[control] register default router (NoRoute and NoMethod)...")
	{
		// 首先注册NoMethod处理器（方法不允许）
		// NoMethod应该在NoRoute之前注册，以确保当路径存在但方法不支持时能正确返回405
		c.ginRouter.NoMethod(func(ctx *gin.Context) {
			_, span := tracer.StartSpan(ctx.Request.Context(), "ginRouter", "noMethod")
			defer span.End()
			c.logger.Warnf("[control] 405 Method Not Allowed: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
			ctx.JSON(http.StatusMethodNotAllowed, commonhttp.BaseResponse{
				Code:    http.StatusMethodNotAllowed,
				Message: "业务处理失败",
				Data:    "Method Not Allowed",
				Success: false,
			})
		})
	}

	{
		// 然后注册NoRoute处理器（路径不存在）
		c.ginRouter.NoRoute(func(ctx *gin.Context) {
			_, span := tracer.StartSpan(ctx.Request.Context(), "ginRouter", "noRouter")
			defer span.End()

			var routeInfo struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			}

			for _, r := range c.ginRouter.Routes() {
				if stringer.CompareIgnoreCase(r.Path, ctx.Request.URL.Path) {
					routeInfo.Path = r.Path
					routeInfo.Method = r.Method
				}
			}
			if stringer.IsBlank(routeInfo.Path) {
				// 路由不存在
				c.logger.Warnf("[control] 404 Not Found: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
				msg := fmt.Sprintf("Router %s Not Found", ctx.Request.URL.Path)
				ctx.JSON(http.StatusNotFound, commonhttp.BaseResponse{
					Code:    http.StatusNotFound,
					Data:    msg,
					Success: false,
					Message: "业务处理失败",
				})
			} else {
				// 路由存在，但请求方法不匹配
				c.logger.Warnf("[control] 405 Method Not Allowed: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
				msg := fmt.Sprintf("Router %s Not Allow %s Method", ctx.Request.URL.Path, ctx.Request.Method)
				ctx.JSON(http.StatusMethodNotAllowed, commonhttp.BaseResponse{
					Code:    http.StatusMethodNotAllowed,
					Data:    msg,
					Success: false,
					Message: "业务处理失败",
				})
			}
		})
	}

	{
		// 定义路由
		allRouterUri := fmt.Sprintf("%s", "routers")
		c.RegisterRouter([]commonhttp.RouterHandler{
			&commonhttp.MyRouter{
				Name:   "router",
				Uri:    allRouterUri,
				Method: http.MethodGet,
				HandlerFunc: func(ctx *gin.Context) {
					_, span := tracer.StartSpan(ctx.Request.Context(), "ginRouter", "routers")
					defer span.End()
					commonhttp.SuccessResponse(ctx, gin.H{
						"routers": c.routerGroups.Values(),
					})
				},
				Enabled:         true,
				Desc:            "获取全部路由信息",
				EnableJWtVerify: false,
			},
		})
	}

	// 如果启用了pprof，则注册pprof路由
	if c.config.Pprof {
		c.logger.Info("[control] pprof enabled, registering pprof routes")
		pprofUri := fmt.Sprintf("%s", "debug/pprof")
		c.RegisterGroupedRouter(&commonhttp.MyGroupRouter{
			Group: pprofUri,
			Routers: []commonhttp.RouterHandler{
				&commonhttp.MyRouter{
					Name:            "pprof",
					Uri:             "/",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Index),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprof",
				},
				&commonhttp.MyRouter{
					Name:            "pprofCmdLine",
					Uri:             "/cmdline",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Cmdline),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofCmdLine",
				},
				&commonhttp.MyRouter{
					Name:            "pprofProfile",
					Uri:             "/profile",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Profile),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofProfile",
				},
				&commonhttp.MyRouter{
					Name:            "pprofSymbol",
					Uri:             "/symbol",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Symbol),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofSymbol",
				},
				&commonhttp.MyRouter{
					Name:            "pprofSymbol",
					Uri:             "/symbol",
					Method:          http.MethodPost,
					HandlerFunc:     gin.WrapF(pprof.Symbol),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofSymbol",
				},
				&commonhttp.MyRouter{
					Name:            "pprofTrace",
					Uri:             "/trace",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Trace),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofTrace",
				},
				&commonhttp.MyRouter{
					Name:            "pprofAllocs",
					Uri:             "/allocs",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Handler("allocs").ServeHTTP),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofAllocs",
				},
				&commonhttp.MyRouter{
					Name:            "pprofBlock",
					Uri:             "/block",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Handler("block").ServeHTTP),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofBlock",
				},
				&commonhttp.MyRouter{
					Name:            "pprofGoroutine",
					Uri:             "/goroutine",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Handler("goroutine").ServeHTTP),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofGoroutine",
				},
				&commonhttp.MyRouter{
					Name:            "pprofHeap",
					Uri:             "/heap",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Handler("heap").ServeHTTP),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofHeap",
				},
				&commonhttp.MyRouter{
					Name:            "pprofMutex",
					Uri:             "/mutex",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Handler("mutex").ServeHTTP),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofMutex",
				},
				&commonhttp.MyRouter{
					Name:            "pprofThreadcreate",
					Uri:             "/threadcreate",
					Method:          http.MethodGet,
					HandlerFunc:     gin.WrapF(pprof.Handler("threadcreate").ServeHTTP),
					EnableJWtVerify: false,
					Enabled:         true,
					Desc:            "pprofThreadcreate",
				},
			},
			MiddlewaresFunc: nil,
		})
	}

	c.logger.Debug("[control] default router registered successfully")
}

// GetSessionManager 获取会话管理器
// @return jwt.SessionManager 会话管理器
func (c *Control) GetSessionManager() jwt.SessionManager {
	return c.sessionManager
}

// validateRouters 验证路由处理器列表的有效性
func (c *Control) validateRouters(routers []commonhttp.RouterHandler) error {
	if len(routers) == 0 {
		return nil
	}

	// 检查是否有nil路由
	for i, router := range routers {
		if router == nil {
			return fmt.Errorf("router at index %d is nil", i)
		}

		// 检查路由基本信息是否完整
		if stringer.IsBlank(router.GetUri()) {
			return fmt.Errorf("router at index %d has empty URI", i)
		}
		if stringer.IsBlank(router.GetMethod()) {
			return fmt.Errorf("router at index %d has empty method", i)
		}
		if router.GetHandlerFunc() == nil {
			return fmt.Errorf("router at index %d has nil handler function", i)
		}
	}

	return nil
}

// RegisterRouter 注册HTTP路由
// @param routers []commonhttp.RouterHandler 路由处理器列表
func (c *Control) RegisterRouter(routers []commonhttp.RouterHandler) {
	// 验证路由有效性
	if err := c.validateRouters(routers); err != nil {
		c.logger.Errorf("[control] invalid routers: %v", err)
		return
	}

	c.logger.Infof("[control] registering %d router(s)...", len(routers))
	c.RegisterGroupedRouter(&commonhttp.MyGroupRouter{
		Group:           "default",
		Routers:         routers,
		MiddlewaresFunc: make([]gin.HandlerFunc, 0),
	})
	c.logger.Info("[control] routers registered successfully")
}

// RegisterGroupedRouter 注册支持路由组的HTTP路由
// @param groupRouter commonhttp.GroupRouterHandler 路由组处理器
func (c *Control) RegisterGroupedRouter(groupRouter commonhttp.GroupRouterHandler) {
	if groupRouter == nil {
		c.logger.Error("[control] groupRouter is nil")
		return
	}

	// c.routerMutex.Lock()
	// defer c.routerMutex.Unlock()

	groupName := groupRouter.GetGroup()
	if stringer.IsBlank(groupName) {
		groupName = "default"
	}

	routers := groupRouter.GetRouterHandler()
	middlewares := groupRouter.GetMiddlewares()

	// 验证路由有效性
	if err := c.validateRouters(routers); err != nil {
		c.logger.Errorf("[control] invalid routers in group %s: %v", groupName, err)
		return
	}

	c.logger.Infof("[control] registering %d router(s) and %d middleware(s) for group '%s'...", len(routers), len(middlewares), groupName)

	// 首先判断 routerGroups 是否有 groupName
	// if existingGroup, exists := c.routerGroups[groupName]; exists {
	if existingGroup, exists := c.routerGroups.Get(groupName); exists {
		// 存在则合并 Routers 和 MiddlewaresFunc
		c.logger.Debugf("[control] merging with existing router group: %s", groupName)

		// 预分配足够容量的切片，减少内存分配
		existingRouters := existingGroup.GetRouterHandler()
		existingMiddlewares := existingGroup.GetMiddlewares()

		// 合并路由
		mergedRouters := make([]commonhttp.RouterHandler, 0, len(existingRouters)+len(routers))
		mergedRouters = append(mergedRouters, existingRouters...)
		mergedRouters = append(mergedRouters, routers...)

		// 合并中间件
		mergedMiddlewares := make([]gin.HandlerFunc, 0, len(existingMiddlewares)+len(middlewares))
		mergedMiddlewares = append(mergedMiddlewares, existingMiddlewares...)
		mergedMiddlewares = append(mergedMiddlewares, middlewares...)

		// 创建新的组路由器
		mergedGroupRouter := &commonhttp.MyGroupRouter{
			Group:           groupName,
			Routers:         mergedRouters,
			MiddlewaresFunc: mergedMiddlewares,
		}

		// 更新路由组
		// c.routerGroups[groupName] = mergedGroupRouter
		c.routerGroups.Put(groupName, mergedGroupRouter)
		c.logger.Debugf("[control] merged router group '%s': total %d routers, %d middlewares", groupName, len(mergedRouters), len(mergedMiddlewares))
	} else {
		// 不存在则直接添加
		c.logger.Debugf("[control] adding new router group: %s", groupName)
		// c.routerGroups[groupName] = groupRouter
		c.routerGroups.Put(groupName, groupRouter)
	}

	c.logger.Infof("[control] router group '%s' registered successfully", groupName)
}

func (c *Control) deduplicateRouters() {
	c.logger.Debug("[control] starting deduplicate routers...")

	// routerGroups 每个组都进行去重

	iter := c.routerGroups.Iterator()
	defer func() {
		_ = iter.Close()
	}()
	for iter.HasNext() {
		value := iter.Value()
		groupName := value.Key
		groupRouter := value.Value
		c.logger.Debugf("[control] deduplicating routers in group: %s", groupName)

		// 获取该组的所有路由
		routers := groupRouter.GetRouterHandler()
		routerCount := len(routers)

		// 如果没有路由，直接跳过
		if routerCount == 0 {
			continue
		}

		// 检查是否有重复路由
		seenRouters := make(map[string]bool)
		hasDuplicates := false

		// 第一次遍历：检查是否有重复
		for _, router := range routers {
			key := fmt.Sprintf("%s:%s", router.GetUri(), router.GetMethod())
			if seenRouters[key] {
				hasDuplicates = true
				break
			}
			seenRouters[key] = true
		}

		// 如果没有重复路由，直接跳过后续处理
		if !hasDuplicates {
			c.logger.Debugf("[control] no duplicate routers found in group: %s", groupName)
			continue
		}

		// 第二次遍历：创建唯一路由列表
		uniqueRouters := make([]commonhttp.RouterHandler, 0, routerCount)
		seenRouters = make(map[string]bool) // 重置map

		for _, router := range routers {
			key := fmt.Sprintf("%s:%s", router.GetUri(), router.GetMethod())
			if !seenRouters[key] {
				seenRouters[key] = true
				uniqueRouters = append(uniqueRouters, router)
				c.logger.Debugf("[control] added unique router: %s %s", router.GetMethod(), router.GetUri())
			} else {
				c.logger.Debugf("[control] skipped duplicate router: %s %s", router.GetMethod(), router.GetUri())
			}
		}

		// 创建新的组路由器替换原有的
		dupCount := routerCount - len(uniqueRouters)
		c.logger.Debugf("[control] removed %d duplicate routers from group: %s", dupCount, groupName)

		newGroupRouter := &commonhttp.MyGroupRouter{
			Group:           groupRouter.GetGroup(),
			Routers:         uniqueRouters,
			MiddlewaresFunc: groupRouter.GetMiddlewares(),
		}

		// 更新路由组
		c.routerGroups.Put(groupName, newGroupRouter)
	}
	// for groupName, groupRouter := range c.routerGroups {
	// 	c.logger.Debugf("[control] deduplicating routers in group: %s", groupName)
	//
	// 	// 获取该组的所有路由
	// 	routers := groupRouter.GetRouterHandler()
	// 	routerCount := len(routers)
	//
	// 	// 如果没有路由，直接跳过
	// 	if routerCount == 0 {
	// 		continue
	// 	}
	//
	// 	// 检查是否有重复路由
	// 	seenRouters := make(map[string]bool)
	// 	hasDuplicates := false
	//
	// 	// 第一次遍历：检查是否有重复
	// 	for _, router := range routers {
	// 		key := fmt.Sprintf("%s:%s", router.GetUri(), router.GetMethod())
	// 		if seenRouters[key] {
	// 			hasDuplicates = true
	// 			break
	// 		}
	// 		seenRouters[key] = true
	// 	}
	//
	// 	// 如果没有重复路由，直接跳过后续处理
	// 	if !hasDuplicates {
	// 		c.logger.Debugf("[control] no duplicate routers found in group: %s", groupName)
	// 		continue
	// 	}
	//
	// 	// 第二次遍历：创建唯一路由列表
	// 	uniqueRouters := make([]commonhttp.RouterHandler, 0, routerCount)
	// 	seenRouters = make(map[string]bool) // 重置map
	//
	// 	for _, router := range routers {
	// 		key := fmt.Sprintf("%s:%s", router.GetUri(), router.GetMethod())
	// 		if !seenRouters[key] {
	// 			seenRouters[key] = true
	// 			uniqueRouters = append(uniqueRouters, router)
	// 			c.logger.Debugf("[control] added unique router: %s %s", router.GetMethod(), router.GetUri())
	// 		} else {
	// 			c.logger.Debugf("[control] skipped duplicate router: %s %s", router.GetMethod(), router.GetUri())
	// 		}
	// 	}
	//
	// 	// 创建新的组路由器替换原有的
	// 	dupCount := routerCount - len(uniqueRouters)
	// 	c.logger.Debugf("[control] removed %d duplicate routers from group: %s", dupCount, groupName)
	//
	// 	newGroupRouter := &commonhttp.MyGroupRouter{
	// 		Group:           groupRouter.GetGroup(),
	// 		Routers:         uniqueRouters,
	// 		MiddlewaresFunc: groupRouter.GetMiddlewares(),
	// 	}
	//
	// 	// 更新路由组
	// 	c.routerGroups[groupName] = newGroupRouter
	// }

	c.logger.Debug("[control] finished deduplicate routers...")
}

// registerRouter 注册单个路由到指定的gin路由组
func (c *Control) registerRouter(ginGroup *gin.RouterGroup, router commonhttp.RouterHandler) {
	if !router.IsEnabled() {
		return // 跳过禁用的路由
	}

	// 构建URL路径
	// var url string
	// if strings.HasPrefix(router.GetUri(), "/") {
	// 	url = router.GetUri()
	// } else {
	// 	url = "/" + router.GetUri()
	// }
	url := check.IF[string](strings.HasPrefix(router.GetUri(), "/"),
		router.GetUri(),
		"/"+router.GetUri())

	handlerFunc := router.GetHandlerFunc()
	httpMethod := router.GetMethod()

	// 根据HTTP方法注册路由
	switch httpMethod {
	case http.MethodGet:
		ginGroup.GET(url, handlerFunc)
	case http.MethodPost:
		ginGroup.POST(url, handlerFunc)
	case http.MethodPut:
		ginGroup.PUT(url, handlerFunc)
	case http.MethodDelete:
		ginGroup.DELETE(url, handlerFunc)
	case http.MethodPatch:
		ginGroup.PATCH(url, handlerFunc)
	case http.MethodOptions:
		ginGroup.OPTIONS(url, handlerFunc)
	case http.MethodHead:
		ginGroup.HEAD(url, handlerFunc)
	default:
		// 默认使用GET方法
		ginGroup.GET(url, handlerFunc)
	}
}

// initRouters 初始化所有注册的路由
func (c *Control) initRouters() {
	c.logger.Info("[control] start init routers...")

	// 创建用于存储gin路由组的映射
	// key: 组名, value: 该组的主路由组
	ginRouterGroups := make(map[string]*gin.RouterGroup)

	// 遍历所有路由组
	iter := c.routerGroups.Iterator()
	defer func() {
		_ = iter.Close()
	}()
	for iter.HasNext() {
		next := iter.Value()
		groupName := next.Key
		groupRouter := next.Value
		// 获取该组的所有路由
		routers := groupRouter.GetRouterHandler()

		// 获取该组的中间件
		middlewares := groupRouter.GetMiddlewares()

		c.logger.Debugf("[control] processing router group: %s with %d routers and %d middlewares", groupName, len(routers), len(middlewares))

		// 获取或创建该组的主路由组
		var mainGroup *gin.RouterGroup
		if existingGroup, exists := ginRouterGroups[groupName]; exists {
			mainGroup = existingGroup
		} else {
			// 根据组名决定路由组的基础路径
			// basePath := "/"
			// if !stringer.CompareIgnoreCase(groupName, "default") {
			// 	basePath = "/" + groupName
			// }

			basePath := check.IF[string](!stringer.CompareIgnoreCase(groupName, "default"),
				"/"+groupName,
				"/")

			// 添加上下文路径
			fullPath := ""
			if !stringer.CompareIgnoreCase(c.config.ContextPath, "") {
				// 确保ContextPath以/开头
				// if !strings.HasPrefix(c.config.ContextPath, "/") {
				// 	fullPath = "/" + c.config.ContextPath
				// } else {
				// 	fullPath = c.config.ContextPath
				// }
				fullPath = check.IF[string](!strings.HasPrefix(c.config.ContextPath, "/"),
					"/"+c.config.ContextPath,
					c.config.ContextPath)
				// 添加basePath（如果不是根路径）
				if !stringer.CompareIgnoreCase(basePath, "/") {
					// 确保basePath前没有重复的/
					// if strings.HasSuffix(fullPath, "/") {
					// 	fullPath += basePath[1:]
					// } else {
					// 	fullPath += basePath
					// }
					fullPath = check.IF[string](strings.HasSuffix(fullPath, "/"),
						fullPath+basePath[1:],
						fullPath+basePath)
				}
			} else {
				fullPath = basePath
			}

			// 创建新的gin路由组
			mainGroup = c.ginRouter.Group(fullPath)
			// 应用组级别中间件
			for _, middleware := range middlewares {
				mainGroup.Use(middleware)
			}
			ginRouterGroups[groupName] = mainGroup
			c.logger.Debugf("[control] created new gin router group: %s with full path: %s", groupName, fullPath)
		}

		// 为当前组创建认证和非认证子组
		authGroup := mainGroup.Group("/")
		authGroup.Use(jwt.EnableJWT(c.logger, c.sessionManager))

		noAuthGroup := mainGroup.Group("/")

		// 注册该组下的所有路由
		for _, router := range routers {
			if router.GetEnableJWtVerify() {
				c.logger.Debugf("[control] register router uri %s method %s in %s auth group", router.GetUri(), router.GetMethod(), groupName)
				c.registerRouter(authGroup, router)
			} else {
				c.logger.Debugf("[control] register router uri %s method %s in %s no-auth group", router.GetUri(), router.GetMethod(), groupName)
				c.registerRouter(noAuthGroup, router)
			}
		}
	}
	// 遍历所有路由组
	// for groupName, groupRouter := range c.routerGroups {
	// 	// 获取该组的所有路由
	// 	routers := groupRouter.GetRouterHandler()
	//
	// 	// 获取该组的中间件
	// 	middlewares := groupRouter.GetMiddlewares()
	//
	// 	c.logger.Debugf("[control] processing router group: %s with %d routers and %d middlewares", groupName, len(routers), len(middlewares))
	//
	// 	// 获取或创建该组的主路由组
	// 	var mainGroup *gin.RouterGroup
	// 	if existingGroup, exists := ginRouterGroups[groupName]; exists {
	// 		mainGroup = existingGroup
	// 	} else {
	// 		// 根据组名决定路由组的基础路径
	// 		basePath := "/"
	// 		if !stringer.CompareIgnoreCase(groupName, "default") {
	// 			basePath = "/" + groupName
	// 		}
	//
	// 		// 添加上下文路径
	// 		fullPath := ""
	// 		if !stringer.CompareIgnoreCase(c.config.ContextPath, "") {
	// 			// 确保ContextPath以/开头
	// 			if !strings.HasPrefix(c.config.ContextPath, "/") {
	// 				fullPath = "/" + c.config.ContextPath
	// 			} else {
	// 				fullPath = c.config.ContextPath
	// 			}
	// 			// 添加basePath（如果不是根路径）
	// 			if !stringer.CompareIgnoreCase(basePath, "/") {
	// 				// 确保basePath前没有重复的/
	// 				if strings.HasSuffix(fullPath, "/") {
	// 					fullPath += basePath[1:]
	// 				} else {
	// 					fullPath += basePath
	// 				}
	// 			}
	// 		} else {
	// 			fullPath = basePath
	// 		}
	//
	// 		// 创建新的gin路由组
	// 		mainGroup = c.ginRouter.Group(fullPath)
	// 		// 应用组级别中间件
	// 		for _, middleware := range middlewares {
	// 			mainGroup.Use(middleware)
	// 		}
	// 		ginRouterGroups[groupName] = mainGroup
	// 		c.logger.Debugf("[control] created new gin router group: %s with full path: %s", groupName, fullPath)
	// 	}
	//
	// 	// 为当前组创建认证和非认证子组
	// 	authGroup := mainGroup.Group("/")
	// 	authGroup.Use(jwt.EnableJWT(c.logger, c.sessionManager))
	//
	// 	noAuthGroup := mainGroup.Group("/")
	//
	// 	// 注册该组下的所有路由
	// 	for _, router := range routers {
	// 		if router.GetEnableJWtVerify() {
	// 			c.logger.Debugf("[control] register router uri %s method %s in %s auth group", router.GetUri(), router.GetMethod(), groupName)
	// 			c.registerRouter(authGroup, router)
	// 		} else {
	// 			c.logger.Debugf("[control] register router uri %s method %s in %s no-auth group", router.GetUri(), router.GetMethod(), groupName)
	// 			c.registerRouter(noAuthGroup, router)
	// 		}
	// 	}
	// }
}

func (c *Control) GetUploadDir() string {
	c.logger.Debugf("[control] get upload dir...")
	if stringer.IsBlank(c.config.UploadDir) {
		return ""
	}
	return filepath.Clean(c.config.UploadDir)
}

func (c *Control) GetUploadCacheDir() string {
	c.logger.Debugf("[control] get upload cache dir...")
	return filepath.Clean(filepath.Join(c.GetUploadDir(), "temp"))
}

func (c *Control) ClearUploadCache() error {
	c.logger.Debugf("[control] clear upload cache...")
	if err := path.ClearDir(c.GetUploadCacheDir()); err != nil {
		c.logger.Errorf("[control] clear upload cache failed: %v", err)
		return err
	}
	return nil
}
