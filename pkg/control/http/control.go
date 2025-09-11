package http

import (
	"context"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http/middleware/jwt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http/middleware"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	serverConfig   *config.HttpServerConfig
	server         *http.Server
	ginRouter      *gin.Engine
	ctx            context.Context
	logger         *zap.SugaredLogger
	routers        []commonhttp.RouterHandler
	sessionManager jwt.SessionManager
	once           sync.Once
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.serverConfig.TlsEnabled {
			c.logger.Infof("[control] start https server on %v", c.serverConfig.Address)
			go func() {
				if err := c.server.ListenAndServeTLS(c.serverConfig.TlsCertFile, c.serverConfig.TlsKeyFile); err != nil && !webhttp.IsHttpErrServerClosed(err) {

					failedFunc(err)
				}
			}()
		} else {
			c.logger.Infof("[control] start http server on %v", c.serverConfig.Address)
			go func() {
				if err := c.server.ListenAndServe(); err != nil && !webhttp.IsHttpErrServerClosed(err) {
					failedFunc(err)
				}
			}()
		}
	})
}

func (c *Control) Shutdown() error {
	c.logger.Infof("[control] shutdown http server...")
	if err := c.server.Shutdown(c.ctx); err != nil {
		c.logger.Errorf("[control] shutdown http server failed: %v", err)
		return err
	}
	return nil
}

func NewWebServerControl(serverConfig *config.HttpServerConfig, loggerControl *logger.Control) *Control {
	webLogger := loggerControl.GenLogger(logger.ModuleWeb)
	webLogger.Infof("[control] start new http server control...")
	gin.SetMode(serverConfig.RunMode)
	// 强制彩色输出
	gin.ForceConsoleColor()

	ctx := context.Background()

	webLogger.Debugf("[control] generate gin engine...")
	engine := gin.Default()
	webLogger.Debugf("[control] register gin middleware...")
	engine.Use(middleware.EnableRequestID(webLogger))
	engine.Use(middleware.EnableCors())
	engine.Use(middleware.EnableGzip())

	// 创建会话管理器
	webLogger.Debugf("[control] create session manager...")
	sessionManager := jwt.NewMemorySessionManager(webLogger, ctx)

	// 注册JWT中间件
	// webLogger.Debugf("[control] register JWT middleware...")
	// engine.Use(middleware.EnableJWT(webLogger, sessionManager))

	// engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 	// 你的自定义格式
	// 127.0.0.1 - [2025-09-08 19:57:12.078] "GET /example/ping HTTP/2.0 200 50.066µs "curl/7.68.0" "
	// "%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 	return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 		param.ClientIP,
	// 		param.TimeStamp.Format("2006-01-02 15:04:05.000"),
	// 		param.Method,
	// 		param.Path,
	// 		param.Request.Proto,
	// 		param.StatusCode,
	// 		param.Latency,
	// 		param.Request.UserAgent(),
	// 		param.ErrorMessage,
	// 	)
	// }))
	webLogger.Debugf("[control] generate http server...")
	srv := &http.Server{
		Addr:    serverConfig.Address,
		Handler: engine,
	}
	webLogger.Debugf("[control] generate http control...")
	control := &Control{
		serverConfig:   serverConfig,
		server:         srv,
		ctx:            ctx,
		logger:         webLogger,
		ginRouter:      engine,
		sessionManager: sessionManager,
	}

	control.logger.Warnf("[control] current not setting router please call control.RegisterRouter to register router...")
	// 在这里不调用
	// control.initRouters()
	return control
}

func (c *Control) RegisterRouter(routers []commonhttp.RouterHandler) {
	c.logger.Infof("[control] register router...")
	c.routers = routers
	c.logger.Debugf("[control] starting define router...")
	c.initRouters()
	c.logger.Infof("[control] register router success...")
}

func (c *Control) initRouters() {
	c.logger.Infof("[control] start init routers...")

	for _, router := range c.routers {
		if router.IsEnabled() {
			var url string
			if strings.HasPrefix(router.GetUri(), "/") {
				// 避免重复添加前缀
				url = fmt.Sprintf("%s%s", c.serverConfig.ContextPath, router.GetUri())
			} else {
				url = fmt.Sprintf("%s/%s", c.serverConfig.ContextPath, router.GetUri())
			}
			c.logger.Debugf("[control] register router uri %s method %s", url, router.GetMethod())

			// 获取处理器函数
			handlerFunc := router.GetHandlerFunc()

			// 检查是否需要启用JWT验证
			switch router.GetMethod() {
			case http.MethodPatch:
				if router.GetEnableJWtVerify() {
					c.ginRouter.PATCH(url, handlerFunc, jwt.EnableJWT(c.logger, c.sessionManager))
				} else {
					c.ginRouter.PATCH(url, handlerFunc)
				}
			case http.MethodOptions:
				if router.GetEnableJWtVerify() {
					c.ginRouter.OPTIONS(url, handlerFunc, jwt.EnableJWT(c.logger, c.sessionManager))
				} else {
					c.ginRouter.OPTIONS(url, handlerFunc)
				}
			case http.MethodPut:
				if router.GetEnableJWtVerify() {
					c.ginRouter.PUT(url, handlerFunc, jwt.EnableJWT(c.logger, c.sessionManager))
				} else {
					c.ginRouter.PUT(url, handlerFunc)
				}
			case http.MethodHead:
				if router.GetEnableJWtVerify() {
					c.ginRouter.HEAD(url, handlerFunc, jwt.EnableJWT(c.logger, c.sessionManager))
				} else {
					c.ginRouter.HEAD(url, handlerFunc)
				}
			case http.MethodDelete:
				if router.GetEnableJWtVerify() {
					c.ginRouter.DELETE(url, handlerFunc, jwt.EnableJWT(c.logger, c.sessionManager))
				} else {
					c.ginRouter.DELETE(url, handlerFunc)
				}
			case http.MethodPost:
				if router.GetEnableJWtVerify() {
					c.ginRouter.POST(url, handlerFunc, jwt.EnableJWT(c.logger, c.sessionManager))
				} else {
					c.ginRouter.POST(url, handlerFunc)
				}
			case http.MethodGet:
				fallthrough
			default:
				if router.GetEnableJWtVerify() {
					c.ginRouter.GET(url, handlerFunc, jwt.EnableJWT(c.logger, c.sessionManager))
				} else {
					c.ginRouter.GET(url, handlerFunc)
				}
			}
		}
	}
}
