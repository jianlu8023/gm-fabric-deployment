package http

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"time"
)

func getDefaultConfig() *config.HttpServerConfig {
	return &config.HttpServerConfig{
		Enabled:      false,
		Http2Enabled: true, // 默认启用HTTP/2支持
	}
}

// HTTP 服务器网络与超时相关默认参数
//
// @description 统一收敛在常量中，避免散落在启动流程各处；取值参考生产环境加固实践
const (
	// tcpKeepAlive TCP 保活探测间隔，用于及时回收异常断电/断网对端留下的半开连接，避免句柄与 goroutine 堆积
	tcpKeepAlive = 3 * time.Minute
	// readHeaderTimeout 读取请求头的最长时间，只约束头部阶段，防御 Slowloris 慢连接攻击，不影响大包体上传
	readHeaderTimeout = 15 * time.Second
	// maxHeaderBytes 单个请求头集合的最大字节数，防止头部放大攻击
	maxHeaderBytes = 1 << 20 // 1MB
	// shutdownTimeout 优雅关闭等待进行中请求完成的最长时间，超时后强制关闭，避免被卡死的长连接无限阻塞
	shutdownTimeout = 15 * time.Second
	// h2MaxConcurrentStreams HTTP/2 单连接最大并发流数，HTTP/2 规范建议值
	h2MaxConcurrentStreams = 250
	// h2MaxBufferSize HTTP/2 可接受的单帧最大字节数与流控窗口大小（合法范围 16KB~16MB）
	h2MaxBufferSize = 1 << 20 // 1MB
	// http2IdleTimeout HTTP/2 空闲连接回收时间，与 http.Server.IdleTimeout 保持一致
	http2IdleTimeout = 120 * time.Second

	// peerCertExpiringSoonWarn 对端证书剩余有效期小于该阈值时输出告警，便于运维提前续期。
	peerCertExpiringSoonWarn = 30 * 24 * time.Hour
)
