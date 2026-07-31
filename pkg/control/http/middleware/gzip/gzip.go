package gzip

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// headerAcceptEncoding HTTP Accept-Encoding 头名称
	headerAcceptEncoding = "Accept-Encoding"
	// headerContentEncoding HTTP Content-Encoding 头名称
	headerContentEncoding = "Content-Encoding"
	// headerVary HTTP Vary 头名称
	headerVary = "Vary"
	// headerContentLength HTTP Content-Length 头名称
	headerContentLength = "Content-Length"
	// headerETag HTTP ETag 头名称
	headerETag = "ETag"
	// headerConnection HTTP Connection 头名称
	headerConnection = "Connection"
	// defaultPoolSize gzip.Writer 池默认大小，需根据并发量调整
	// 每个 writer 内部占用约 640KB，32 个约 20MB 常驻内存
	defaultPoolSize = 32
)

// defaultExcludedExtensions 默认不压缩的文件扩展名集合
// 这些扩展名多为已压缩的二进制格式，再次压缩收益小且浪费 CPU
var defaultExcludedExtensions = map[string]struct{}{
	".png":  {},
	".gif":  {},
	".jpeg": {},
	".jpg":  {},
	".gz":   {},
	".zip":  {},
	".br":   {},
	".webp": {},
}

// gzipWriter 包装 gin.ResponseWriter，将写入数据通过 gzip.Writer 压缩后再写入底层 ResponseWriter
//
// @description 转发 Write/WriteString 等写入操作到 gzip.Writer，由 gzip.Writer 完成压缩后写入底层 ResponseWriter
// @struct
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write 写入字节数据并进行 gzip 压缩
//
// @description 写入数据时移除 Content-Length 头避免长度不匹配，并通过 gzip.Writer 进行压缩
// @param data []byte 待写入的字节数据
// @return int 写入的字节数
// @return error 写入过程中的错误
func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del(headerContentLength)
	return g.writer.Write(data)
}

// WriteString 写入字符串数据并进行 gzip 压缩
//
// @description 写入字符串时移除 Content-Length 头避免长度不匹配，并通过 gzip.Writer 进行压缩
// @param s string 待写入的字符串
// @return int 写入的字节数
// @return error 写入过程中的错误
func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del(headerContentLength)
	return g.writer.Write([]byte(s))
}

// Flush 刷新 gzip.Writer 内部缓冲并刷新底层 ResponseWriter
//
// @description 用于流式响应场景，先刷新 gzip.Writer 的内部缓冲到底层 ResponseWriter，再刷新底层 ResponseWriter
func (g *gzipWriter) Flush() {
	_ = g.writer.Flush()
	g.ResponseWriter.Flush()
}

// WriteHeader 写入响应状态码
//
// @description 写入状态码时移除 Content-Length 头避免长度不匹配
// @param code int HTTP 状态码
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del(headerContentLength)
	g.ResponseWriter.WriteHeader(code)
}

// Hijack 实现 http.Hijacker 接口，支持 WebSocket 等协议的连接接管
//
// @description 当客户端升级协议（如 WebSocket）时，接管底层 TCP 连接，绕过 gzip 压缩
// @return net.Conn 底层网络连接
// @return *bufio.ReadWriter 缓冲读写器
// @return error 接管失败时的错误
func (g *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := g.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

// 确保 gzipWriter 实现了 http.Hijacker 接口
var _ http.Hijacker = (*gzipWriter)(nil)

// gzipHandler gzip 压缩处理器
//
// @description 使用基于 channel 的固定大小池复用 gzip.Writer，避免 sync.Pool 在 GC 时被清空导致的反复分配
// @struct
type gzipHandler struct {
	level int
	pool  chan *gzip.Writer
}

// newGzipHandler 创建 gzip 处理器
//
// @description 创建 gzip 处理器，初始化固定大小的 gzip.Writer 池。压缩级别非法时使用 DefaultCompression
// @param level int 压缩级别，参考 compress/gzip 包中的常量
// @return *gzipHandler gzip 处理器实例
func newGzipHandler(level int) *gzipHandler {
	if !isCompressionLevelValid(level) {
		level = gzip.DefaultCompression
	}
	return &gzipHandler{
		level: level,
		pool:  make(chan *gzip.Writer, defaultPoolSize),
	}
}

// isCompressionLevelValid 校验压缩级别是否合法
//
// @description 校验压缩级别是否在 compress/gzip 包定义的合法范围内
// @param level int 压缩级别
// @return bool true 表示合法
func isCompressionLevelValid(level int) bool {
	return level == gzip.DefaultCompression ||
		level == gzip.NoCompression ||
		level == gzip.HuffmanOnly ||
		(level >= gzip.BestSpeed && level <= gzip.BestCompression)
}

// acquire 从池中获取一个 gzip.Writer
//
// @description 优先从 channel 池中获取复用的 writer 并 Reset 到目标写入器；池为空时新建一个 writer。
// 该方法为非阻塞，池空时直接新建，避免高并发下请求被串行化
// @param w io.Writer 底层写入器
// @return *gzip.Writer gzip 写入器
func (h *gzipHandler) acquire(w io.Writer) *gzip.Writer {
	select {
	case gz := <-h.pool:
		gz.Reset(w)
		return gz
	default:
		gz, _ := gzip.NewWriterLevel(w, h.level)
		return gz
	}
}

// release 归还 writer 到池中
//
// @description 将 writer 重置到 io.Discard 后归还到 channel 池；池满时丢弃由 GC 回收。
// 重置到 io.Discard 是为了避免归还的 writer 持有对前一个 ResponseWriter 的引用导致内存无法回收
// @param gz *gzip.Writer 待归还的 gzip 写入器
func (h *gzipHandler) release(gz *gzip.Writer) {
	gz.Reset(io.Discard)
	select {
	case h.pool <- gz:
	default:
		// 池已满，丢弃 writer 由 GC 回收
	}
}

// shouldCompress 判断请求是否需要压缩
//
// @description 根据请求头和路径判断是否需要 gzip 压缩：
// 1. 客户端必须通过 Accept-Encoding 头声明支持 gzip
// 2. 排除协议升级请求（如 WebSocket）
// 3. 排除默认不压缩的文件扩展名（已压缩的二进制格式）
// @param req *http.Request HTTP 请求
// @return bool true 表示需要压缩
func (h *gzipHandler) shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get(headerAcceptEncoding), "gzip") {
		return false
	}
	if strings.Contains(req.Header.Get(headerConnection), "Upgrade") {
		return false
	}
	extension := filepath.Ext(req.URL.Path)
	if _, ok := defaultExcludedExtensions[extension]; ok {
		return false
	}
	return true
}

// Handle gin 中间件处理函数
//
// @description 处理 HTTP 请求的 gzip 压缩流程：
// 1. 判断请求是否需要压缩，不需要则直接放行
// 2. 从池中获取 gzip.Writer 并 Reset 到当前 ResponseWriter
// 3. 设置 Content-Encoding: gzip 等响应头
// 4. 用 gzipWriter 包装 ResponseWriter，后续写入走压缩通道
// 5. 请求处理完成后关闭 gzip.Writer（写出尾部数据），并归还到池中
// @param c *gin.Context gin 上下文
func (h *gzipHandler) Handle(c *gin.Context) {
	if !h.shouldCompress(c.Request) {
		c.Next()
		return
	}

	gz := h.acquire(c.Writer)
	c.Header(headerContentEncoding, "gzip")
	c.Writer.Header().Add(headerVary, headerAcceptEncoding)

	// 处理 ETag 头：压缩会改变响应体，需将强 ETag 转为弱 ETag（W/ 前缀）避免缓存失效
	if originalEtag := c.GetHeader(headerETag); originalEtag != "" && !strings.HasPrefix(originalEtag, "W/") {
		c.Header(headerETag, "W/"+originalEtag)
	}

	c.Writer = &gzipWriter{ResponseWriter: c.Writer, writer: gz}
	defer func() {
		if c.Writer.Size() < 0 {
			// 没有写入任何响应体，重置到 io.Discard 避免 Close 时写出多余的 gzip footer
			gz.Reset(io.Discard)
		}
		_ = gz.Close()
		if c.Writer.Size() > -1 {
			c.Header(headerContentLength, strconv.Itoa(c.Writer.Size()))
		}
		h.release(gz)
	}()
	c.Next()
}

// EnableGzip 启用 gzip 压缩中间件
//
// @description 返回使用默认压缩级别（DefaultCompression）的 gzip 中间件。
// 使用基于 channel 的固定大小池复用 gzip.Writer，避免 sync.Pool 在 GC 时被清空导致的反复分配
// （每个 gzip.Writer 内部约 640KB，sync.Pool 每次 GC 都会清空，导致下次请求重新分配）
// @return gin.HandlerFunc gin 中间件处理函数
func EnableGzip() gin.HandlerFunc {
	handler := newGzipHandler(gzip.DefaultCompression)
	return handler.Handle
}

// EnableGzipWithLevel 启用指定压缩级别的 gzip 压缩中间件
//
// @description 返回指定压缩级别的 gzip 中间件。注意：压缩级别主要影响 CPU 占用，对单次内存分配大小影响较小
// （compress/flate 的 hashHead/hashPrev 内联数组固定约 768KB，与级别无关）
// @param level int 压缩级别，参考 compress/gzip 包中的常量（BestSpeed=1, BestCompression=9, DefaultCompression=-1）
// @return gin.HandlerFunc gin 中间件处理函数
func EnableGzipWithLevel(level int) gin.HandlerFunc {
	handler := newGzipHandler(level)
	return handler.Handle
}
