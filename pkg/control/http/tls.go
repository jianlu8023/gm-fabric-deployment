package http

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"golang.org/x/net/http2"
	"net"
	"os"
	"strings"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/tjfoc/gmsm/gmtls"
	gmx509 "github.com/tjfoc/gmsm/x509"
)

// http2Settings 构建HTTP/2服务端显式调优参数
//
// @description 基于 golang.org/x/net/http2 显式配置单连接并发流上限、流控窗口、帧大小与空闲回收，
// 替代 http2.Server 零值默认；TLS 场景经 http2.ConfigureServer 注册后获得 h2 GOAWAY 优雅关闭能力，
// h2c 明文场景由 h2c.NewHandler 直接使用
// @return *http2.Server HTTP/2服务端配置
func http2Settings() *http2.Server {
	return &http2.Server{
		// MaxConcurrentStreams 单连接最大并发流数，HTTP/2 规范建议值
		MaxConcurrentStreams: h2MaxConcurrentStreams,
		// MaxReadFrameSize 可接受的单帧最大字节数（合法范围 16KB~16MB）
		MaxReadFrameSize: h2MaxBufferSize,
		// MaxUploadBufferPerConnection 连接级流控窗口
		MaxUploadBufferPerConnection: h2MaxBufferSize,
		// MaxUploadBufferPerStream 单流流控窗口
		MaxUploadBufferPerStream: h2MaxBufferSize,
		// IdleTimeout 空闲连接回收，避免长连接长期占用句柄，与 http.Server.IdleTimeout 保持一致
		IdleTimeout: http2IdleTimeout,
		// NewWriteScheduler 优先级调度器，与 Go 运行时内置实现保持一致
		NewWriteScheduler: func() http2.WriteScheduler {
			return http2.NewPriorityWriteScheduler(nil)
		},
	}
}

// newTCPListener 创建带KeepAlive探测的TCP监听器
//
// @description 使用 net.ListenConfig 显式指定 KeepAlive 间隔，accept 出连接时会据此开启
// SO_KEEPALIVE 并设置 TCP_KEEPIDLE/TCP_KEEPINTVL（同时默认开启 TCP_NODELAY），
// 以便及时回收异常断电/断网留下的半开连接，避免句柄与 goroutine 堆积；
// TLS/GM TLS 场景在该裸监听器上再包装 tls.NewListener/gmtls.NewListener
// @return net.Listener TCP监听器
// @return error 监听失败原因
func (c *Control) newTCPListener() (net.Listener, error) {
	lc := net.ListenConfig{KeepAlive: tcpKeepAlive}
	return lc.Listen(c.ctx, "tcp", c.config.Address)
}

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
			// gmtls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			// gmtls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			// gmtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			// gmtls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
		CurvePreferences: []gmtls.CurveID{
			gmtls.X25519, gmtls.CurveP521, gmtls.CurveP384, gmtls.CurveP256,
		},
		SessionTicketsDisabled: false, // 启用会话票据
		// 对端（客户端）证书链校验通过后的审计回调：记录 GM mTLS 对端证书信息与临期告警
		VerifyPeerCertificate: c.gmPeerCertLogger(),
	}

	// TODO 目前 没找到支持http2的方法 暂时注释掉
	// 根据HTTP/2配置决定是否启用HTTP/2协议协商
	if c.config.Http2Enabled {
		// gmTLSConfig.NextProtos = []string{"h2", "http/1.1"}
		gmTLSConfig.NextProtos = []string{"http/1.1"}
		c.logger.Infof("[http/control] HTTP/2 enabled for TLS connections")
	} else {
		gmTLSConfig.NextProtos = []string{"http/1.1"}
		c.logger.Infof("[http/control] HTTP/2 disabled, using HTTP/1.1 only")
	}

	// 根据配置决定使用单证书还是双证书模式
	if c.config.TlsGMSingleCert {
		// 单证书模式
		c.logger.Info("[http/control] GM TLS using single certificate mode")
		return c.loadGMSingleCert(gmTLSConfig)
	}

	// 双证书模式（默认）
	c.logger.Info("[http/control] GM TLS using dual certificate mode")
	return c.loadGMDualCert(gmTLSConfig)
}

// loadGMSingleCert 加载GM单证书模式
// @description 加载国密TLS单证书模式，从配置数组中取第一对证书和密钥文件
// @param gmTLSConfig *gmtls.Config GM TLS配置对象
// @return error 加载过程中可能产生的错误
func (c *Control) loadGMSingleCert(gmTLSConfig *gmtls.Config) error {
	// 验证证书和密钥文件数量
	if len(c.config.TlsCertFile) < 1 || len(c.config.TlsKeyFile) < 1 {
		c.logger.Errorf("[http/control] GM TLS单证书模式至少需要一对证书和密钥文件，当前证书数量: %d, 密钥数量: %d", len(c.config.TlsCertFile), len(c.config.TlsKeyFile))
		return errors.New("GM TLS单证书模式至少需要一对证书和密钥文件")
	}

	// 取第一对证书
	certFile := strings.TrimSpace(c.config.TlsCertFile[0])
	keyFile := strings.TrimSpace(c.config.TlsKeyFile[0])

	// 验证文件路径不为空
	if stringer.IsBlank(certFile) {
		c.logger.Errorf("[http/control] GM TLS certificate file path is empty")
		return errors.New("GM TLS certificate file path is empty")
	}
	if stringer.IsBlank(keyFile) {
		c.logger.Errorf("[http/control] GM TLS key file path is empty")
		return errors.New("GM TLS key file path is empty")
	}

	// 加载单证书
	cert, err := gmtls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		c.logger.Errorf("[http/control] failed to load GM TLS certificate: %v", err)
		return fmt.Errorf("加载GM TLS证书失败: %v", err)
	}
	// 预解析 leaf 证书并回写 Leaf 字段：跳过握手时每连接一次的重复 ASN.1 解析，
	// 同时在启动阶段打印证书主体/SAN/到期时间，临期告警便于运维巡检
	if leaf, parseErr := gmx509.ParseCertificate(cert.Certificate[0]); parseErr != nil {
		c.logger.Warnf("[http/control] 解析GM TLS leaf证书失败: %v", parseErr)
	} else {
		cert.Leaf = leaf
		c.logGMServerCert(leaf, certFile)
	}

	gmTLSConfig.Certificates = []gmtls.Certificate{cert}
	c.logger.Infof("[http/control] 成功加载GM模式单证书: %s -> %s", certFile, keyFile)

	// 配置根证书
	if err := c.setupGMRootCA(gmTLSConfig); err != nil {
		return err
	}

	c.gmTlsConfig = gmTLSConfig
	return nil
}

// loadGMDualCert 加载GM双证书模式
// @description 加载国密TLS双证书模式，从配置数组中加载所有证书和密钥文件对（签名和加密）
// @param gmTLSConfig *gmtls.Config GM TLS配置对象
// @return error 加载过程中可能产生的错误
func (c *Control) loadGMDualCert(gmTLSConfig *gmtls.Config) error {
	// GM模式需要至少两套keypair：一个签名，一个加密
	certFiles := c.config.TlsCertFile
	keyFiles := c.config.TlsKeyFile

	// 检查证书和密钥文件数量是否匹配且至少有两对
	if len(certFiles) != len(keyFiles) {
		c.logger.Errorf("[http/control] GM双证书模式证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(certFiles), len(keyFiles))
		return fmt.Errorf("GM双证书模式证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(certFiles), len(keyFiles))
	}
	if len(certFiles) < 2 {
		c.logger.Errorf("[http/control] GM双证书模式至少需要两套keypair（签名和加密），当前只有 %d 套", len(certFiles))
		return fmt.Errorf("GM双证书模式至少需要两套keypair（签名和加密），当前只有 %d 套", len(certFiles))
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
			c.logger.Errorf("[http/control] 第%d个证书文件路径为空", i+1)
			return fmt.Errorf("第%d个证书文件路径为空", i+1)
		}
	}
	for i, file := range keyFiles {
		if stringer.IsBlank(file) {
			c.logger.Errorf("[http/control] 第%d个密钥文件路径为空", i+1)
			return fmt.Errorf("第%d个密钥文件路径为空", i+1)
		}
	}

	// 加载所有keypair
	var certificates []gmtls.Certificate
	for i := 0; i < len(certFiles); i++ {
		cert, err := gmtls.LoadX509KeyPair(certFiles[i], keyFiles[i])
		if err != nil {
			c.logger.Errorf("[http/control] 加载第%d套GM TLS证书失败: %v", i+1, err)
			return fmt.Errorf("加载第%d套GM TLS证书失败: %v", i+1, err)
		}
		// 预解析 leaf 证书并回写 Leaf 字段：跳过握手时每连接一次的重复 ASN.1 解析，
		// 同时在启动阶段打印证书主体/SAN/到期时间，临期告警便于运维巡检
		if leaf, parseErr := gmx509.ParseCertificate(cert.Certificate[0]); parseErr != nil {
			c.logger.Warnf("[http/control] 解析第%d套GM TLS leaf证书失败: %v", i+1, parseErr)
		} else {
			cert.Leaf = leaf
			c.logGMServerCert(leaf, certFiles[i])
		}
		certificates = append(certificates, cert)
		c.logger.Debugf("[http/control] 成功加载第%d套GM TLS证书: %s -> %s", i+1, certFiles[i], keyFiles[i])
	}

	// 设置证书到GM TLS配置
	gmTLSConfig.Certificates = certificates
	c.logger.Infof("[http/control] 成功加载GM模式 %d 套keypair", len(certificates))
	// 配置根证书
	if err := c.setupGMRootCA(gmTLSConfig); err != nil {
		return err
	}

	c.gmTlsConfig = gmTLSConfig
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
		c.logger.Errorf("[http/control] failed to read GM CA cert file: %v", err)
		return fmt.Errorf("读取GM CA证书失败: %v", err)
	}

	caCertPool := gmx509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		c.logger.Errorf("[http/control] failed to parse GM CA cert file: %s", rootCaCertFile)
		return fmt.Errorf("解析GM CA证书失败: %s", rootCaCertFile)
	}

	gmTLSConfig.ClientCAs = caCertPool
	gmTLSConfig.ClientAuth = gmtls.VerifyClientCertIfGiven // 根据需要验证客户端证书
	c.logger.Debugf("[http/control] client certificate verification enabled with GM CA cert: %s", rootCaCertFile)
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
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,

			// TLS 1.2 - ECDSA 证书
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,

			// TLS 1.3 套件（Go 运行时自动选择，此处仅作参考）
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
		// InsecureSkipVerify: false,
		CurvePreferences: []tls.CurveID{
			//tls.X25519MLKEM768, 后量子 go1.22 不支持
			tls.X25519, // 优先使用X25519椭圆曲线
			tls.CurveP256,
			tls.CurveP384,
			tls.CurveP521,
		},
		// PreferServerCipherSuites: true,  // 优先使用服务端加密套件 Deprecated
		SessionTicketsDisabled: false, // 启用会话票据
		// 对端（客户端）证书链校验通过后的审计回调：记录 mTLS 对端证书信息与临期告警
		VerifyPeerCertificate: c.stdlibPeerCertLogger(),
		ClientSessionCache:    tls.NewLRUClientSessionCache(256),
	}

	// 根据HTTP/2配置决定是否启用HTTP/2协议协商
	if c.config.Http2Enabled {
		tlsConfig.NextProtos = []string{"h2", "http/1.1"}
		c.logger.Infof("[http/control] HTTP/2 enabled for TLS connections")
	} else {
		tlsConfig.NextProtos = []string{"http/1.1"}
		c.logger.Infof("[http/control] HTTP/2 disabled, using HTTP/1.1 only")
	}

	// 验证证书和密钥文件数量匹配
	if len(c.config.TlsCertFile) != len(c.config.TlsKeyFile) {
		c.logger.Errorf("[http/control] TLS证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(c.config.TlsCertFile), len(c.config.TlsKeyFile))
		return fmt.Errorf("TLS证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(c.config.TlsCertFile), len(c.config.TlsKeyFile))
	}
	if len(c.config.TlsCertFile) == 0 {
		c.logger.Error("[http/control] TLS至少需要一对证书和密钥文件")
		return errors.New("TLS至少需要一对证书和密钥文件")
	}

	// 加载所有证书
	var certificates []tls.Certificate
	for i := 0; i < len(c.config.TlsCertFile); i++ {
		certFile := strings.TrimSpace(c.config.TlsCertFile[i])
		keyFile := strings.TrimSpace(c.config.TlsKeyFile[i])

		if stringer.IsBlank(certFile) {
			c.logger.Errorf("[http/control] 第%d个TLS证书文件路径为空", i+1)
			return fmt.Errorf("第%d个TLS证书文件路径为空", i+1)
		}
		if stringer.IsBlank(keyFile) {
			c.logger.Errorf("[http/control] 第%d个TLS密钥文件路径为空", i+1)
			return fmt.Errorf("第%d个TLS密钥文件路径为空", i+1)
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			c.logger.Errorf("[http/control] 加载第%d套TLS证书失败: %v", i+1, err)
			return fmt.Errorf("加载第%d套TLS证书失败: %v", i+1, err)
		}
		// 预解析 leaf 证书并回写 Leaf 字段：跳过握手时每连接一次的重复 ASN.1 解析（性能优化），
		// 同时在启动阶段打印证书主体/SAN/到期时间，临期告警便于运维巡检
		if leaf, parseErr := x509.ParseCertificate(cert.Certificate[0]); parseErr != nil {
			c.logger.Warnf("[http/control] 解析第%d套TLS leaf证书失败: %v", i+1, parseErr)
		} else {
			cert.Leaf = leaf
			c.logStdlibServerCert(leaf, certFile)
		}
		certificates = append(certificates, cert)
	}
	tlsConfig.Certificates = certificates
	c.logger.Infof("[http/control] 成功加载 %d 套TLS证书", len(certificates))

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
		c.logger.Errorf("[http/control] failed to read CA cert file: %v", err)
		return fmt.Errorf("读取CA证书失败: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		c.logger.Errorf("[http/control] failed to parse CA cert file: %s", rootCaCertFile)
		return fmt.Errorf("解析CA证书失败: %s", rootCaCertFile)
	}

	tlsConfig.ClientCAs = caCertPool
	tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven // 根据需要验证客户端证书
	c.logger.Debugf("[http/control] client certificate verification enabled with CA cert: %s", rootCaCertFile)
	return nil
}

func (c *Control) serverGMTls(failedFunc func(err error)) {
	// 先建立带 KeepAlive 探测的裸 TCP 监听器，再包装为 GM TLS 监听器，
	// 不用 gmtls.Listen 是为了显式控制 TCP KeepAlive，及时回收半开连接
	tcpListener, err := c.newTCPListener()
	if err != nil {
		c.logger.Errorf("[http/control] failed to create TCP listener: %v", err)
		if failedFunc != nil {
			failedFunc(err)
		}
		return
	}
	listener := gmtls.NewListener(tcpListener, c.gmTlsConfig)

	// 注意：GM TLS不支持HTTP/2，因为HTTP/2需要的ALPN协议协商和GM TLS不兼容
	if c.config.Http2Enabled {
		c.logger.Warnf("[http/control] HTTP/2 is not supported with GM TLS, falling back to HTTP/1.1")
		if err := http2.ConfigureServer(c.server, http2Settings()); err != nil {
			c.logger.Errorf("[http/control] failed to configure HTTP/2 server: %v", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
		c.logger.Info("[http/control] HTTP/2 server configured successfully")
	}

	defer func(listener net.Listener) {
		if err := listener.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			c.logger.Errorf("[http/control] failed to close gm TLS listener: %v", err)
		}
	}(listener)
	// 使用srv.Serve启动服务器
	if err := c.server.Serve(listener); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
		c.logger.Errorf("[http/control] HTTPS server error: %v", err)
		if failedFunc != nil {
			failedFunc(err)
		}
		return
	}
}

func (c *Control) serverTls(failedFunc func(err error)) {
	// 先建立带 KeepAlive 探测的裸 TCP 监听器，再包装为 TLS 监听器，
	// 不用 tls.Listen 是为了显式控制 TCP KeepAlive，及时回收半开连接
	tcpListener, err := c.newTCPListener()
	if err != nil {
		c.logger.Errorf("[http/control] failed to create TCP listener: %v", err)
		if failedFunc != nil {
			failedFunc(err)
		}
		return
	}
	listener := tls.NewListener(tcpListener, c.tlsConfig)

	// 为HTTPS服务器启用HTTP/2支持：ConfigureServer 注册 TLSNextProto["h2"] 处理函数，
	// ALPN 协商由 tlsConfig.NextProtos(h2,http/1.1) 完成
	if c.config.Http2Enabled {
		if err := http2.ConfigureServer(c.server, http2Settings()); err != nil {
			c.logger.Errorf("[http/control] failed to configure HTTP/2 server: %v", err)
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}
		c.logger.Info("[http/control] HTTP/2 server configured successfully")
	}

	defer func(listener net.Listener) {
		if err := listener.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			c.logger.Errorf("[http/control] failed to close TLS listener: %v", err)
		}
	}(listener)

	// 使用srv.Serve启动服务器
	if err := c.server.Serve(listener); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
		c.logger.Errorf("[http/control] HTTPS server error: %v", err)
		if failedFunc != nil {
			failedFunc(err)
		}
		return
	}
}

func (c *Control) serverNoTls(failedFunc func(err error)) {
	// 使用带 KeepAlive 探测的 TCP 监听器，及时回收半开连接
	listener, err := c.newTCPListener()
	if err != nil {
		c.logger.Errorf("[http/control] failed to create TCP listener: %v", err)
		if failedFunc != nil {
			failedFunc(err)
		}
		return
	}
	defer func(listener net.Listener) {
		if err := listener.Close(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			c.logger.Errorf("[http/control] failed to close TCP listener: %v", err)
		}
	}(listener)

	// 使用srv.Serve启动服务器
	if err := c.server.Serve(listener); err != nil && !commonhttp.IsHttpErrServerClosed(err) {
		c.logger.Errorf("[http/control] HTTP server error: %v", err)
		if failedFunc != nil {
			failedFunc(err)
		}
		return
	}
}
