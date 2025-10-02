package http

import (
	"net/http"
	"path"
	"strings"

	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/static"
)

// StaticFileConfig 静态文件配置
// @description 用于配置静态文件服务的参数
// @param Prefix 静态文件URL前缀
// @param Enabled 是否启用静态文件服务
type StaticFileConfig struct {
	Prefix  string
	Enabled bool
}

// WithStaticFiles 启用静态文件服务
// @description 配置并启用嵌入的静态文件服务
// @param config StaticFileConfig 静态文件配置
// @return Option 配置函数
func WithStaticFiles(config StaticFileConfig) Option {
	return func(control *Control) {
		if !config.Enabled || config.Prefix == "" {
			return
		}

		// 确保前缀格式正确（以/开头）
		prefix := config.Prefix
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}

		// 创建文件服务器
		fileServer := http.FileServer(http.FS(static.EmbeddedStaticFiles))

		// 注册静态文件路由
		control.RegisterRouter([]commonhttp.RouterHandler{
			&commonhttp.MyRouter{
				Name:            "static",
				Uri:             prefix + "/*filepath",
				Method:          http.MethodGet,
				Enabled:         true,
				Desc:            "静态文件服务",
				EnableJWtVerify: false, // 明确指定不需要JWT验证
				HandlerFunc: func(ctx *gin.Context) {
					_, span := tracer.StartSpan(ctx.Request.Context(), "ginRouter", "static")
					defer span.End()
					// 获取完整的请求路径
					reqPath := ctx.Request.URL.Path
					span.SetAttributes(
						attribute.String("reqPath", reqPath),
					)
					// 安全检查：防止目录遍历攻击
					cleanPath := path.Clean(reqPath)
					if strings.HasPrefix(cleanPath, "../") {
						// ctx.Status(http.StatusForbidden)
						commonhttp.FailedResponseWithMessage(ctx, commonhttp.Forbidden, commonhttp.ErrMsgForbidden)
						span.SetStatus(codes.Error, "forbidden")
						return
					}

					// 获取URL中的filepath参数
					filepath := ctx.Param("filepath")
					if filepath == "" {
						filepath = "index.html"
					}

					// 清理文件路径，确保安全，去除前导斜杠
					// filepath = strings.TrimPrefix(filepath, "/")
					// if filepath == "" {
					// 	filepath = "index.html"
					// }

					// 构建新的请求URL路径
					newURL := *ctx.Request.URL
					newURL.Path = filepath

					// 使用文件服务器处理请求
					fileServer.ServeHTTP(ctx.Writer, &http.Request{
						Method: ctx.Request.Method,
						URL:    &newURL,
						Header: ctx.Request.Header,
						Body:   ctx.Request.Body,
						Host:   ctx.Request.Host,
					})
				},
			},
		})

		control.logger.Infof("[control] static file service enabled at %s", prefix)
	}
}

// WithDefaultStaticFiles 使用默认配置启用静态文件服务
// @description 使用默认配置(前缀为/static)启用嵌入的静态文件服务
// @return Option 配置函数
func WithDefaultStaticFiles() Option {
	return WithStaticFiles(StaticFileConfig{
		Prefix:  "/static",
		Enabled: true,
	})
}
