package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"net/http"
	"net/http/pprof"
	"strings"
)

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

	c.registerDefaultRouter()

	c.deduplicateRouters()

	// 确保路由索引是最新的
	c.buildRouteIndex()

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

			requestPath := ctx.Request.URL.Path
			requestMethod := ctx.Request.Method

			// 使用路由索引进行 O(1) 查找
			if methods, exists := c.routeIndex[requestPath]; exists {
				// 路由存在，检查请求方法是否允许
				if !methods[requestMethod] {
					// 路由存在，但请求方法不匹配 -> 405
					c.logger.Warnf("[control] 405 Method Not Allowed: %s %s", requestMethod, requestPath)
					msg := fmt.Sprintf("Router %s Not Allow %s Method", requestPath, requestMethod)
					ctx.JSON(http.StatusMethodNotAllowed, commonhttp.BaseResponse{
						Code:    http.StatusMethodNotAllowed,
						Data:    msg,
						Success: false,
						Message: "业务处理失败",
					})
					return
				}
			}

			// 路由不存在
			c.logger.Warnf("[control] 404 Not Found: %s %s", requestMethod, requestPath)
			msg := fmt.Sprintf("Router %s Not Found", requestPath)
			ctx.JSON(http.StatusNotFound, commonhttp.BaseResponse{
				Code:    http.StatusNotFound,
				Data:    msg,
				Success: false,
				Message: "业务处理失败",
			})
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

// buildRouteIndex 构建路由索引
// @description 基于当前已注册的所有路由，构建一个快速查找索引
// 索引结构为: path -> method -> bool (表示该路径是否支持该HTTP方法)
// 用于优化 NoRoute 处理器的性能，避免每次404请求都遍历所有路由
func (c *Control) buildRouteIndex() {
	c.logger.Debug("[control] building route index...")

	routeIndex := make(map[string]map[string]bool)

	// 遍历所有路由组中的路由
	iter := c.routerGroups.Iterator()
	defer func() {
		_ = iter.Close()
	}()
	for iter.HasNext() {
		next := iter.Value()
		groupRouter := next.Value
		routers := groupRouter.GetRouterHandler()

		for _, router := range routers {
			if !router.IsEnabled() {
				continue
			}
			uri := router.GetUri()
			method := router.GetMethod()

			// 规范化路径
			path := check.IF[string](strings.HasPrefix(uri, "/"), uri, "/"+uri)

			// 构建 path -> methods 索引
			if methods, exists := routeIndex[path]; exists {
				methods[method] = true
			} else {
				routeIndex[path] = map[string]bool{method: true}
			}
		}
	}

	c.routeIndex = routeIndex
	c.logger.Debugf("[control] route index built: %d unique paths", len(routeIndex))

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
