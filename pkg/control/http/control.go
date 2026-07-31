package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/tjfoc/gmsm/gmtls"

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
	routeIndex     map[string]map[string]bool                            // 路由索引，用于快速查找 path -> methods
}

// NewWebServerControl 创建Web服务器控制器
// @param config *config.HttpServerConfig HTTP服务器配置
// @param loggerControl *logger.Control 日志控制器
// @param opts ...Option 可选的配置选项
// @return *Control Web服务器控制器实例
func NewWebServerControl(serverConfig *config.HttpServerConfig, loggerControl *logger.Control, opts ...Option) (*Control, error) {
	serverConfig = check.IF[*config.HttpServerConfig](serverConfig == nil,
		getDefaultConfig(),
		serverConfig,
	)
	if !serverConfig.Enabled {
		return nil, errors.New("http is not enabled")
	}
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	webLogger := loggerControl.GenLogger(logger.ModuleWeb)
	webLogger.Info("[control] start new http server control...")
	gin.SetMode(serverConfig.RunMode)
	// 强制彩色输出
	gin.ForceConsoleColor()

	ctx := context.Background()

	webLogger.Debug("[control] generate gin engine...")
	engine := gin.New()

	// 设置可信代理，影响所有 c.ClientIP() 调用（requestid/secure/logger 中间件及 ulule 限流内部均使用）
	// 与 registerMiddlewares 中的 trustedProxiesCIDRList 保持一致，避免同一请求在不同位置获取到不同的客户端 IP
	var trustedProxies []string
	if serverConfig.TrustedProxies != nil && serverConfig.TrustedProxies.Enabled {
		trustedProxies = serverConfig.TrustedProxies.IPs
	} else {
		// 未启用时传入空切片，gin 将不信任任何代理，c.ClientIP() 仅返回 RemoteAddr，防止 XFF 头被伪造
		trustedProxies = []string{}
	}
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		webLogger.Errorf("[control] failed to set trusted proxies: %v", err)
	} else if serverConfig.TrustedProxies != nil && serverConfig.TrustedProxies.Enabled {
		webLogger.Infof("[control] gin trusted proxies enabled with %d entries", len(trustedProxies))
	} else {
		webLogger.Info("[control] gin trusted proxies disabled, c.ClientIP() will use RemoteAddr only")
	}

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
		routeIndex:   make(map[string]map[string]bool),
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

	// 注册中间件
	control.registerMiddlewares(engine)

	// 配置TLS
	if err := control.setupTLSConfig(); err != nil {
		return nil, err
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

			c.initRouters()

			if c.config.TlsEnabled {
				if c.config.TlsGM {
					c.logger.Infof("[control] start gm https server on %v", c.config.Address)
					go c.serverGMTls(failedFunc)
				} else {
					c.logger.Infof("[control] start https server on %v", c.config.Address)
					go c.serverTls(failedFunc)
				}

			} else {
				c.logger.Infof("[control] start http server on %v", c.config.Address)
				go c.serverNoTls(failedFunc)
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

// GetSessionManager 获取会话管理器
// @return jwt.SessionManager 会话管理器
func (c *Control) GetSessionManager() jwt.SessionManager {
	return c.sessionManager
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

// GetUploadDir 获取上传目录路径
// 该方法返回配置中指定的文件上传目录路径
// 返回值:
//
//	清理后的上传目录绝对路径，如果未配置则返回空字符串
//
// 处理逻辑:
// 1. 检查配置中的UploadDir是否为空
// 2. 如果为空，返回空字符串
// 3. 否则，使用filepath.Clean规范化路径格式
func (c *Control) GetUploadDir() string {
	c.logger.Debugf("[control] get upload dir...")
	if stringer.IsBlank(c.config.UploadDir) {
		return ""
	}
	return filepath.Clean(c.config.UploadDir)
}

// GetUploadCacheDir 获取上传缓存目录路径
// 该方法返回文件上传的临时缓存目录路径，通常用于存储上传过程中的临时文件
// 返回值:
//
//	清理后的缓存目录绝对路径，格式为：上传目录/temp
//
// 实现逻辑:
// 1. 调用GetUploadDir获取基础上传目录
// 2. 在基础上传目录下创建名为"temp"的子目录作为缓存目录
// 3. 使用filepath.Clean规范化最终路径格式
func (c *Control) GetUploadCacheDir() string {
	c.logger.Debugf("[control] get upload cache dir...")
	return filepath.Clean(filepath.Join(c.GetUploadDir(), "temp"))
}

// ClearUploadCache 清理上传缓存目录
// 该方法清空上传缓存目录中的所有文件和子目录，释放临时占用的存储空间
// 返回值:
//
//	成功时返回nil，失败时返回错误信息
//
// 实现逻辑:
// 1. 获取上传缓存目录路径
// 2. 调用path.ClearDir清空该目录下的所有内容
// 3. 如果清理过程中发生错误，记录错误日志并返回错误
// 注意:
//
//	该操作会永久性删除缓存目录中的所有内容，请谨慎使用
func (c *Control) ClearUploadCache() error {
	c.logger.Debugf("[control] clear upload cache...")
	if err := path.ClearDir(c.GetUploadCacheDir()); err != nil {
		c.logger.Errorf("[control] clear upload cache failed: %v", err)
		return err
	}
	return nil
}
