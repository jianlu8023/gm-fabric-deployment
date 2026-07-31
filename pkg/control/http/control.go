package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"strings"
	"sync"

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
	// routerGroups   map[string]commonhttp.GroupRouterHandler
	// routerMutex    sync.RWMutex
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

	srv := &http.Server{
		Addr:    serverConfig.Address,
		Handler: engine.Handler(),
		// ReadTimeout:  30 * time.Second,  // 设置读取超时
		// WriteTimeout: 60 * time.Second,  // 设置写入超时
		// IdleTimeout:  120 * time.Second, // 设置空闲超时
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
			c.registerDefaultRouter()
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
		allRouterUri := fmt.Sprintf("%s", "endpoints")
		c.RegisterRouter([]commonhttp.RouterHandler{
			&commonhttp.MyRouter{
				Name:   "endpoints",
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

// deduplicateRouters 去重路由组中的重复路由
// 该方法遍历所有注册的路由组，检查并移除每个组内的重复路由（相同URI和HTTP方法的路由）
// 重复路由定义为：具有相同URI路径和HTTP方法的路由
// 实现逻辑：
// 1. 遍历所有路由组
// 2. 对于每个组，首先检查是否存在重复路由
// 3. 如果存在重复，则创建一个只包含唯一路由的新列表
// 4. 使用新的唯一路由列表替换原有的路由组
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
// 该方法将通用路由器接口转换为Gin框架的具体路由注册
// 参数:
//
//	ginGroup: Gin路由组，用于注册路由的目标组
//	router: 通用路由器接口，包含路由的URI、HTTP方法和处理函数等信息
//
// 处理逻辑:
// 1. 检查路由是否启用，如未启用则跳过注册
// 2. 规范化URL路径，确保以'/'开头
// 3. 根据HTTP方法类型，将路由注册到相应的Gin方法处理器
// 4. 支持GET、POST、PUT、DELETE、PATCH、OPTIONS和HEAD方法
// 5. 未知HTTP方法默认使用GET方法注册
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
// 该方法是路由系统的核心初始化函数，负责将所有注册的路由转换为Gin框架可识别的路由
// 主要功能:
// 1. 遍历所有注册的路由组
// 2. 根据配置创建相应的Gin路由组，并应用组级别中间件
// 3. 处理上下文路径(ContextPath)的拼接和规范化
// 4. 为每个路由组创建认证(auth)和非认证(no-auth)子组
// 5. 根据路由的JWT验证设置，将路由注册到相应的子组
// 路由组路径生成规则:
// - 默认为'/'，非默认组名为'/groupName'
// - 如果配置了ContextPath，则会作为前缀添加
// - 自动处理路径分隔符，避免重复的'/'
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
