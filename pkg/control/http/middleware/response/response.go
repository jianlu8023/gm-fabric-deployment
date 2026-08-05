package response

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// 默认配置常量
const (
	// defaultMaxBodyBytes 默认响应体最大捕获字节数
	// 超过此长度的响应体仅记录长度，不再缓存到 buffer，避免大响应撑爆内存/日志
	defaultMaxBodyBytes = 1024
	// truncationSuffix 截断时追加的省略号标记，提示日志查看者响应体已被截断
	truncationSuffix = "...[truncated]"
)

// EnableResponseLog 启用响应日志中间件
//
// @description 包装 gin.ResponseWriter 捕获响应体用于日志记录；
// 配置驱动：通过 cfg.MaxBodyBytes 限制单次响应体缓存上限（默认 1024 字节），
// 超过部分仅记录长度不缓存，避免大响应（如文件下载、大 JSON）撑爆内存和日志；
// bodyLogWriter 已实现 Hijack/Flush/WriteString/WriteHeader，与 gzipWriter 风格一致，
// 兼容 WebSocket(Hijack) 与流式响应(Flush) 场景
// @param webLogger *zap.SugaredLogger 日志记录器
// @param cfg *config.ResponseLogConfig 响应日志配置，为 nil 时使用默认值
// @return gin.HandlerFunc gin 中间件函数
func EnableResponseLog(webLogger *zap.SugaredLogger, cfg *config.ResponseLogConfig) gin.HandlerFunc {
	maxBytes := defaultMaxBodyBytes
	if cfg != nil && cfg.MaxBodyBytes > 0 {
		maxBytes = cfg.MaxBodyBytes
	}

	return func(ctx *gin.Context) {
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "response")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)

		// 包装 ResponseWriter，捕获响应体（受 maxBytes 限制）
		blw := &bodyLogWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
			maxBytes:       maxBytes,
		}
		ctx.Writer = blw

		startTime := time.Now()
		ctx.Next()
		duration := time.Since(startTime)

		// 日志字段：耗时 + 状态码 + 实际写入字节数 + 截断后的响应体预览
		webLogger.Debugf("Response DurationTime: %v Status: %d BodyBytes: %d Body: %s",
			duration.String(),
			blw.Status(),
			blw.totalBytes,
			blw.bodyPreview(),
		)

		span.SetStatus(codes.Ok, "success")
	}
}

// bodyLogWriter 包装 gin.ResponseWriter，捕获写入的响应体用于日志记录
//
// @description 转发 Write/WriteString/Flush/WriteHeader/Hijack 到底层 ResponseWriter，
// 同时将写入的数据捕获到 body buffer（受 maxBytes 限制，超过后仅累加 totalBytes 不再缓存）；
// 与 gzip.gzipWriter 保持接口一致，确保 WebSocket(Hijack) 与流式响应(Flush) 场景下行为正确
// @struct
type bodyLogWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer // 响应体缓存，受 maxBytes 限制
	maxBytes   int           // body buffer 最大容量，超过后不再累加
	totalBytes int           // 实际写入的总字节数（不受 maxBytes 限制）
}

// Write 写入字节数据并捕获到 body buffer
//
// @description 转发写入操作到底层 ResponseWriter，同时将数据捕获到 body buffer（受 maxBytes 限制）
// @param b []byte 待写入的字节数据
// @return int 写入的字节数
// @return error 写入过程中的错误
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.captureBytes(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 写入字符串数据并捕获到 body buffer
//
// @description 转发写入操作到底层 ResponseWriter，同时将字符串捕获到 body buffer（受 maxBytes 限制）
// @param s string 待写入的字符串
// @return int 写入的字节数
// @return error 写入过程中的错误
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.captureBytes([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

// captureBytes 捕获写入的数据到 body buffer，超过 maxBytes 后停止累加
//
// @description 将写入数据累加到 totalBytes（用于日志显示实际响应大小），
// 同时在 body buffer 未满时缓存数据（用于日志预览响应体内容）
// @param b []byte 待捕获的字节数据
func (w *bodyLogWriter) captureBytes(b []byte) {
	w.totalBytes += len(b)
	if w.body.Len() >= w.maxBytes {
		return
	}
	remaining := w.maxBytes - w.body.Len()
	if len(b) <= remaining {
		w.body.Write(b)
	} else {
		w.body.Write(b[:remaining])
	}
}

// bodyPreview 返回响应体预览字符串，截断时追加省略号标记
//
// @description 当实际写入字节数超过已缓存字节数时，追加 truncationSuffix 提示日志查看者响应体已被截断
// @return string 响应体预览字符串
func (w *bodyLogWriter) bodyPreview() string {
	preview := w.body.String()
	if w.totalBytes > w.body.Len() {
		preview += truncationSuffix
	}
	return preview
}

// Flush 刷新底层 ResponseWriter，用于流式响应（SSE/chunked）场景
//
// @description 直接转发到底层 ResponseWriter 的 Flush，不做额外缓存处理；
// 流式响应通常无需记录每次 flush 的内容，仅记录最终汇总信息
func (w *bodyLogWriter) Flush() {
	w.ResponseWriter.Flush()
}

// WriteHeader 写入响应状态码
//
// @description 直接转发到底层 ResponseWriter 的 WriteHeader
// @param code int HTTP 状态码
func (w *bodyLogWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

// Hijack 实现 http.Hijacker 接口，支持 WebSocket 等协议的连接接管
//
// @description 当客户端升级协议（如 WebSocket）时，接管底层 TCP 连接，
// 绕过响应体捕获（Hijack 后的数据不再经过 ResponseWriter）
// @return net.Conn 底层网络连接
// @return *bufio.ReadWriter 缓冲读写器
// @return error 接管失败时的错误
func (w *bodyLogWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

// 确保 bodyLogWriter 实现了 http.Hijacker 接口
var _ http.Hijacker = (*bodyLogWriter)(nil)
