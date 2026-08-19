package http

// 编码器				平台支持					主要优势
// encoding/json（默认）	所有					最大兼容性，无额外依赖
// go-json				所有					良好的加速效果，纯 Go，广泛兼容
// jsoniter				所有					良好的加速效果，灵活的配置
// sonic				仅支持 AVX 的 x86_64		通过 JIT 和 SIMD 实现最高吞吐量

// 标签			效果
// go_json		将 encoding/json 替换为 go-json
// jsoniter		将 encoding/json 替换为 jsoniter
// sonic avx	将 encoding/json 替换为 sonic（需要 AVX CPU 指令集）
// nomsgpack	禁用 MsgPack 渲染支持

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHtmlTemplate 注册HTML模板引擎
// 该方法用于设置Gin框架的HTML模板引擎，以便在处理请求时能够渲染HTML页面
// @param templ *template.Template HTML模板实例，通常使用template.ParseFiles或template.ParseGlob创建
// 注意：此方法应在启动服务器前调用，确保所有路由处理函数可以正确使用模板渲染功能
func (c *Control) RegisterHtmlTemplate(templ *template.Template) {
	c.ginRouter.SetHTMLTemplate(templ)
}

// SetFuncMap 设置模板函数映射
// 该方法用于注册自定义模板函数，这些函数可以在HTML模板中使用
// @param funcMap template.FuncMap 模板函数映射，键为函数名，值为对应的函数
// 注意：此方法应在注册模板前调用，确保函数可以在模板中使用
func (c *Control) SetFuncMap(funcMap template.FuncMap) {
	c.ginRouter.SetFuncMap(funcMap)
}

// Static 注册静态文件目录
// 该方法用于将指定目录的静态文件（如CSS、JavaScript、图片等）提供给客户端访问
// @param relativePath string 相对URL路径，如 "/static"
// @param root string 静态文件的本地目录路径
func (c *Control) Static(relativePath, root string) {
	c.ginRouter.Static(relativePath, root)
}

// StaticFile 注册单个静态文件
// 该方法用于将单个文件映射到指定的URL路径
// @param relativePath string 相对URL路径，如 "/favicon.ico"
// @param filepath string 本地文件路径
func (c *Control) StaticFile(relativePath, filepath string) {
	c.ginRouter.StaticFile(relativePath, filepath)
}

// StaticFS 注册静态文件系统
// 该方法使用自定义的http.FileSystem提供静态文件服务
// @param relativePath string 相对URL路径
// @param fs http.FileSystem 文件系统实现
func (c *Control) StaticFS(relativePath string, fs http.FileSystem) {
	c.ginRouter.StaticFS(relativePath, fs)
}

// Use 添加中间件
// 该方法用于向Gin引擎添加中间件，中间件会对所有请求生效
// @param middleware ...gin.HandlerFunc 一个或多个中间件处理函数
func (c *Control) Use(middleware ...gin.HandlerFunc) {
	c.ginRouter.Use(middleware...)
}
