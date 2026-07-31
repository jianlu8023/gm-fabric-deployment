package http

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/cors"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/gzip"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/ipblacklist"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/iphelper"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/ipwhitelist"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/language"
	middlewarelogger "github.com/jianlu8023/golang-example/pkg/control/http/middleware/logger"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/ratelimit"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/recovery"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/requestid"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/secure"
	"github.com/tjfoc/gmsm/gmtls"
	gmx509 "github.com/tjfoc/gmsm/x509"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// setupTLSConfig 配置TLS（GM TLS或标准TLS）
// @description 根据配置启用不同的TLS模式（GM TLS或标准TLS），并加载相应的证书和配置
// @return error 配置过程中可能产生的错误
func (c *Control) setupTLSConfig() error {
	if !c.config.TlsEnabled {
		return nil
	}

	if c.config.TlsGM {
		return c.setupGMTLSConfig()
	}
	return c.setupStandardTLSConfig()
}

// setupGMTLSConfig 配置GM TLS
// @description 配置国密TLS，支持单证书和双证书模式，加载证书和根证书配置
// @return error 配置过程中可能产生的错误
func (c *Control) setupGMTLSConfig() error {
	// 配置 gm secure
	gmTLSConfig := &gmtls.Config{
		GMSupport: &gmtls.GMSupport{
			WorkMode: gmtls.ModeGMSSLOnly,
		},
		MinVersion: gmtls.VersionGMSSL,
		MaxVersion: gmtls.VersionGMSSL,
		CipherSuites: []uint16{
			gmtls.GMTLS_SM2_WITH_SM1_SM3,
			gmtls.GMTLS_SM2_WITH_SM4_SM3,
			gmtls.GMTLS_ECDHE_SM2_WITH_SM4_SM3,
			gmtls.GMTLS_ECDHE_SM2_WITH_SM1_SM3,
			//gmtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			//gmtls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			//gmtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			//gmtls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
		SessionTicketsDisabled: false, // 启用会话票据
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
	//根据HTTP/2配置决定是否启用HTTP/2协议协商
	if c.config.Http2Enabled {
		gmTLSConfig.NextProtos = []string{"h2", "http/1.1"}
		c.logger.Infof("[control] HTTP/2 enabled for TLS connections")
	} else {
		gmTLSConfig.NextProtos = []string{"http/1.1"}
		c.logger.Infof("[control] HTTP/2 disabled, using HTTP/1.1 only")
	}

	// 根据配置决定使用单证书还是双证书模式
	if c.config.TlsGMSingleCert {
		// 单证书模式
		c.logger.Info("[control] GM TLS using single certificate mode")
		return c.loadGMSingleCert(gmTLSConfig)
	}

	// 双证书模式（默认）
	c.logger.Info("[control] GM TLS using dual certificate mode")
	if err := c.loadGMDualCert(gmTLSConfig); err != nil {
		return err
	}

	// 配置根证书
	if err := c.setupGMRootCA(gmTLSConfig); err != nil {
		return err
	}

	c.gmTlsConfig = gmTLSConfig
	return nil
}

// loadGMSingleCert 加载GM单证书模式
// @description 加载国密TLS单证书模式，需要提供一个证书和一个密钥文件
// @param gmTLSConfig *gmtls.Config GM TLS配置对象
// @return error 加载过程中可能产生的错误
func (c *Control) loadGMSingleCert(gmTLSConfig *gmtls.Config) error {
	// 去除文件路径的空格
	certFile := strings.TrimSpace(c.config.TlsCertFile)
	keyFile := strings.TrimSpace(c.config.TlsKeyFile)

	// 验证文件路径不为空
	if stringer.IsBlank(certFile) {
		c.logger.Errorf("[control] GM TLS certificate file path is empty")
		return errors.New("GM TLS certificate file path is empty")
	}
	if stringer.IsBlank(keyFile) {
		c.logger.Errorf("[control] GM TLS key file path is empty")
		return errors.New("GM TLS key file path is empty")
	}

	// 加载单证书
	cert, err := gmtls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		c.logger.Errorf("[control] failed to load GM TLS certificate: %v", err)
		return fmt.Errorf("加载GM TLS证书失败: %v", err)
	}

	gmTLSConfig.Certificates = []gmtls.Certificate{cert}
	c.logger.Infof("[control] 成功加载GM模式单证书: %s -> %s", certFile, keyFile)

	// 配置根证书
	//if err := c.setupGMRootCA(gmTLSConfig); err != nil {
	//	return err
	//}

	//c.gmTlsConfig = gmTLSConfig
	return nil
}

// loadGMDualCert 加载GM双证书模式
// @description 加载国密TLS双证书模式，需要提供两套证书和密钥文件（签名和加密）
// @param gmTLSConfig *gmtls.Config GM TLS配置对象
// @return error 加载过程中可能产生的错误
func (c *Control) loadGMDualCert(gmTLSConfig *gmtls.Config) error {
	// GM模式需要两套keypair：一个签名，一个加密
	// 使用逗号分割证书和密钥文件路径
	certFiles := strings.Split(c.config.TlsCertFile, ",")
	keyFiles := strings.Split(c.config.TlsKeyFile, ",")

	// 检查是否提供了两套证书和密钥
	if len(certFiles) != 2 || len(keyFiles) != 2 {
		c.logger.Errorf("[control] GM双证书模式必须提供两套keypair（签名和加密），当前证书文件数量: %d, 密钥文件数量: %d", len(certFiles), len(keyFiles))
		return errors.New("GM双证书模式必须提供两套keypair，请在 tls_cert_file 和 tls_key_file 中使用逗号分割两套证书和密钥文件路径")
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
			c.logger.Errorf("[control] 第%d个证书文件路径为空", i+1)
			return fmt.Errorf("第%d个证书文件路径为空", i+1)
		}
	}
	for i, file := range keyFiles {
		if stringer.IsBlank(file) {
			c.logger.Errorf("[control] 第%d个密钥文件路径为空", i+1)
			return fmt.Errorf("第%d个密钥文件路径为空", i+1)
		}
	}

	// 加载两套keypair
	var certificates []gmtls.Certificate
	for i := 0; i < 2; i++ {
		cert, err := gmtls.LoadX509KeyPair(certFiles[i], keyFiles[i])
		if err != nil {
			c.logger.Errorf("[control] 加载第%d套GM TLS证书失败: %v", i+1, err)
			return fmt.Errorf("加载第%d套GM TLS证书失败: %v", i+1, err)
		}
		certificates = append(certificates, cert)
		c.logger.Debugf("[control] 成功加载第%d套GM TLS证书: %s -> %s", i+1, certFiles[i], keyFiles[i])
	}

	// 设置证书到GM TLS配置
	gmTLSConfig.Certificates = certificates
	c.logger.Infof("[control] 成功加载GM模式两套keypair，签名证书: %s，加密证书: %s", certFiles[0], certFiles[1])
	//c.gmTlsConfig = gmTLSConfig
	return nil
}

// setupGMRootCA 配置GM根证书
// @description 为GM TLS配置根证书，启用客户端证书验证
// @param gmTLSConfig *gmtls.Config GM TLS配置对象
// @return error 读取或解析CA证书失败时返回错误
func (c *Control) setupGMRootCA(gmTLSConfig *gmtls.Config) error {
	rootCaCertFile := c.config.TlsRCACertFile
	if stringer.IsBlank(rootCaCertFile) {
		return nil
	}

	caCert, err := os.ReadFile(rootCaCertFile)
	if err != nil {
		c.logger.Errorf("[control] failed to read GM CA cert file: %v", err)
		return fmt.Errorf("读取GM CA证书失败: %v", err)
	}

	caCertPool := gmx509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		c.logger.Errorf("[control] failed to parse GM CA cert file: %s", rootCaCertFile)
		return fmt.Errorf("解析GM CA证书失败: %s", rootCaCertFile)
	}

	gmTLSConfig.ClientCAs = caCertPool
	gmTLSConfig.ClientAuth = gmtls.VerifyClientCertIfGiven // 根据需要验证客户端证书
	c.logger.Debugf("[control] client certificate verification enabled with GM CA cert: %s", rootCaCertFile)
	return nil
}

// setupStandardTLSConfig 配置标准TLS
// @description 配置标准TLS，支持HTTP/2协议，加载证书和根证书配置
// @return error 配置过程中可能产生的错误
func (c *Control) setupStandardTLSConfig() error {
	// 配置标准TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // 设置最低TLS版本
		MaxVersion: tls.VersionTLS13, // 设置最高TLS版本
		CipherSuites: []uint16{
			// TLS 1.2 - RSA 证书
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			// TLS 1.2 - ECDSA 证书
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			// TLS 1.3 套件（Go 运行时自动选择，此处仅作参考）
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519, // 优先使用X25519椭圆曲线
			tls.CurveP256,
		},
		//PreferServerCipherSuites: true,  // 优先使用服务端加密套件 Deprecated
		SessionTicketsDisabled: false, // 启用会话票据
	}

	// 根据HTTP/2配置决定是否启用HTTP/2协议协商
	if c.config.Http2Enabled {
		tlsConfig.NextProtos = []string{"h2", "http/1.1"}
		c.logger.Infof("[control] HTTP/2 enabled for TLS connections")
	} else {
		tlsConfig.NextProtos = []string{"http/1.1"}
		c.logger.Infof("[control] HTTP/2 disabled, using HTTP/1.1 only")
	}

	// 加载证书
	certificates, err := tls.LoadX509KeyPair(c.config.TlsCertFile, c.config.TlsKeyFile)
	if err != nil {
		c.logger.Errorf("[control] failed to load TLS certificate: %v", err)
		return err
	}
	tlsConfig.Certificates = []tls.Certificate{certificates}

	// 配置根证书
	if err := c.setupStandardRootCA(tlsConfig); err != nil {
		return err
	}

	c.tlsConfig = tlsConfig
	return nil
}

// setupStandardRootCA 配置标准TLS根证书
// @description 为标准TLS配置根证书，启用客户端证书验证
// @param tlsConfig *tls.Config 标准TLS配置对象
// @return error 读取或解析CA证书失败时返回错误
func (c *Control) setupStandardRootCA(tlsConfig *tls.Config) error {
	rootCaCertFile := c.config.TlsRCACertFile
	if stringer.IsBlank(rootCaCertFile) {
		return nil
	}

	caCert, err := os.ReadFile(rootCaCertFile)
	if err != nil {
		c.logger.Errorf("[control] failed to read CA cert file: %v", err)
		return fmt.Errorf("读取CA证书失败: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		c.logger.Errorf("[control] failed to parse CA cert file: %s", rootCaCertFile)
		return fmt.Errorf("解析CA证书失败: %s", rootCaCertFile)
	}

	tlsConfig.ClientCAs = caCertPool
	tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven // 根据需要验证客户端证书
	c.logger.Debugf("[control] client certificate verification enabled with CA cert: %s", rootCaCertFile)
	return nil
}

// registerMiddlewares 注册Gin中间件
// @description 注册所有HTTP服务器的中间件，包括日志、恢复、IP过滤、安全、限流等
// @param engine *gin.Engine Gin引擎实例
func (c *Control) registerMiddlewares(engine *gin.Engine) {
	c.logger.Info("[control] register gin middleware...")

	// 解析可信代理配置，仅当 RemoteAddr 命中可信代理网段时才信任 X-Forwarded-For / X-Real-IP 头
	// 未配置时 trustedProxiesCIDRList 为 nil，GetClientIP 将仅使用 RemoteAddr，防止 XFF 头被伪造
	var trustedProxiesCIDRList *iphelper.CIDRList
	if c.config.TrustedProxies != nil &&
		c.config.TrustedProxies.Enabled &&
		len(c.config.TrustedProxies.IPs) > 0 {
		cidrList, err := iphelper.NewCIDRList(c.config.TrustedProxies.IPs)
		if err != nil {
			c.logger.Errorf("[control] failed to parse trusted proxies config: %v, X-Forwarded-For will be ignored", err)
		} else {
			trustedProxiesCIDRList = cidrList
			c.logger.Infof("[control] trusted proxies enabled with %d entries", len(c.config.TrustedProxies.IPs))
		}
	}

	// 0. Tracer中间件 - 用于请求追踪
	if c.tracerControl != nil {
		engine.Use(otelgin.Middleware(c.tracerControl.GetServiceName(), otelgin.WithTracerProvider(c.tracerControl.TracerProvider())))
	}

	// 0. 输出请求信息
	engine.Use(middlewarelogger.Logger())

	// 1. 恢复中间件（Recovery Middleware）- 应在最前面注册，捕获所有后续中间件的panic
	engine.Use(recovery.EnableRecovery(c.logger, true))

	// 添加语言支持中间件
	engine.Use(language.EnableLanguageSupport(c.config.Language))

	// 2. 请求ID中间件 - 为每个请求生成唯一标识
	engine.Use(requestid.EnableRequestID(c.logger))

	// 3. IP白名单中间件（如果启用）- 尽早过滤非白名单IP
	if c.config.IPWhiteList != nil &&
		c.config.IPWhiteList.Enabled &&
		len(c.config.IPWhiteList.IPs) > 0 {
		c.logger.Debugf("[control] register IP white list middleware with %d IPs", len(c.config.IPWhiteList.IPs))
		engine.Use(ipwhitelist.EnableIPWhiteList(c.logger, c.config.IPWhiteList.IPs, trustedProxiesCIDRList))
	}

	// 4. IP黑名单中间件（如果启用）- 尽早拒绝黑名单IP
	if c.config.IPBlackList != nil &&
		c.config.IPBlackList.Enabled &&
		len(c.config.IPBlackList.IPs) > 0 {
		c.logger.Debugf("[control] register IP black list middleware with %d IPs", len(c.config.IPBlackList.IPs))
		engine.Use(ipblacklist.EnableIPBlackList(c.logger, c.config.IPBlackList.IPs, trustedProxiesCIDRList))
	}

	if !c.config.TlsGM {
		// TODO gm模式下 会出现一直301的情况
		// 6. TLS安全中间件 - 安全检查，在基础过滤和追踪后执行
		// 判断是否为开发环境（根据Gin模式）
		isDevelopment := gin.Mode() == gin.DebugMode
		// 使用成熟的unrolled/secure包实现的TLS安全中间件
		// 推荐在生产环境使用，提供完整的TLS安全保护功能
		// 只有在启用TLS时才启用SSL重定向
		engine.Use(secure.EnableSecurePackageTLS(c.config.Address, isDevelopment, c.config.TlsEnabled))
	}

	// 如果需要使用不依赖外部包的版本，可以取消注释下面这行
	// engine.Use(secure.EnableUnrolledTLS(webLogger, isDevelopment, serverConfig.Address))

	// 7. CORS中间件 - 跨域处理
	engine.Use(cors.EnableCors())

	// 8. 限流中间件（如果启用）- 在业务逻辑前执行
	if c.config.RateLimit != nil &&
		c.config.RateLimit.Enabled &&
		c.config.RateLimit.RPS > 0 {
		// 创建限流配置
		rateLimitConfig := ratelimit.Config{
			Type:  ratelimit.Type(c.config.RateLimit.Type), // 使用配置文件中的限流类型
			RPS:   c.config.RateLimit.RPS,
			Burst: c.config.RateLimit.Burst,
		}
		// 使用工厂函数创建限流中间件
		engine.Use(ratelimit.NewRateLimitMiddleware(c.logger, rateLimitConfig, trustedProxiesCIDRList))
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
}
