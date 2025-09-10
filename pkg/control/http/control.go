package http

import (
	"context"
	"fmt"
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
	serverConfig *config.HttpServerConfig
	server       *http.Server
	ginRouter    *gin.Engine
	ctx          context.Context
	logger       *zap.SugaredLogger
	routers      []commonhttp.RouterHandler
	once         sync.Once
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
	// gin.ForceConsoleColor()

	webLogger.Debugf("[control] generate gin engine...")
	engine := gin.Default()
	webLogger.Debugf("[control] register gin middleware...")
	engine.Use(middleware.EnableRequestID(webLogger))
	engine.Use(middleware.EnableCors())
	engine.Use(middleware.EnableGzip())

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
		serverConfig: serverConfig,
		server:       srv,
		ctx:          context.Background(),
		logger:       webLogger,
		ginRouter:    engine,
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
			c.logger.Debugf("[control] register router uri %s method %s", router.GetUri(), router.GetMethod())
			var url string
			if strings.HasPrefix(router.GetUri(), "/") {
				// 避免重复添加前缀
				url = fmt.Sprintf("%s%s", c.serverConfig.ContextPath, router.GetUri())
			} else {
				url = fmt.Sprintf("%s/%s", c.serverConfig.ContextPath, router.GetUri())
			}

			switch router.GetMethod() {
			case http.MethodPatch:
				c.ginRouter.PATCH(url, router.GetHandlerFunc())
			case http.MethodOptions:
				c.ginRouter.OPTIONS(url, router.GetHandlerFunc())
			case http.MethodPut:
				c.ginRouter.PUT(url, router.GetHandlerFunc())
			case http.MethodHead:
				c.ginRouter.HEAD(url, router.GetHandlerFunc())
			case http.MethodDelete:
				c.ginRouter.DELETE(url, router.GetHandlerFunc())
			case http.MethodPost:
				c.ginRouter.POST(url, router.GetHandlerFunc())
			case http.MethodGet:
				fallthrough
			default:
				c.ginRouter.GET(url, router.GetHandlerFunc())
			}
		}
	}
}
