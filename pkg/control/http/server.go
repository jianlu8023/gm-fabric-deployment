package http

import (
	"crypto/tls"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/tjfoc/gmsm/gmtls"
	"golang.org/x/net/http2"
	"net"
	"strings"
)

func (c *Control) serverGMTls(failedFunc func(err error)) {
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
}

func (c *Control) serverTls(failedFunc func(err error)) {
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
}

func (c *Control) serverNoTls(failedFunc func(err error)) {
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
}
