package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/http/middleware/cors"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/http/middleware/gzip"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/http/middleware/request"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/http/middleware/requestid"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type ServerControl struct {
	serverConfig *config.HttpServerConfig
	server       *http.Server
	ctx          context.Context
	logger       *zap.SugaredLogger
}

func (s *ServerControl) StartUp(failedFunc func(err error)) {
	if s.serverConfig.TlsEnabled {
		s.logger.Infof("start https server on %v", s.serverConfig.Address)
		go func() {
			if err := s.server.ListenAndServeTLS(s.serverConfig.TlsCertFile, s.serverConfig.TlsKeyFile); err != nil {
				failedFunc(err)
			}
		}()
	} else {
		s.logger.Infof("start http server on %v", s.serverConfig.Address)
		go func() {
			if err := s.server.ListenAndServe(); err != nil {
				failedFunc(err)
			}
		}()
	}
}

func (s *ServerControl) Shutdown() {
	s.logger.Infof("shutdown http server...")
	_ = s.server.Shutdown(s.ctx)
}

func NewServerControl(serverConfig *config.HttpServerConfig, loggerControl *mylogger.Control) *ServerControl {
	webLogger := loggerControl.GenLogger(mylogger.ModuleWeb)
	webLogger.Infof("start new http server control...")
	gin.SetMode(serverConfig.RunMode)
	// gin.ForceConsoleColor()

	engine := gin.Default()
	engine.Use(requestid.EnableRequestID())
	engine.Use(request.EnableRequestLog(webLogger))
	engine.Use(cors.EnableCors())
	engine.Use(gzip.EnableGzip())

	// engine.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 	// 你的自定义格式
	// 127.0.0.1 - [2025-09-08 19:57:12.078] "GET /example/ping HTTP/2.0 200 50.066µs "curl/7.68.0" "
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

	initRouters(engine, serverConfig, webLogger)
	srv := &http.Server{
		Addr:    serverConfig.Address,
		Handler: engine,
	}

	return &ServerControl{
		serverConfig: serverConfig,
		server:       srv,
		ctx:          context.Background(),
		logger:       webLogger,
	}
}

func initRouters(ginEngine *gin.Engine, serverConfig *config.HttpServerConfig, logger *zap.SugaredLogger) {
	logger.Infof("start init routers...")

	for _, router := range NewRouter() {
		if router.Enabled {
			logger.Debugf("register router uri %s method %s", router.Uri, router.Method)
			var url string
			if strings.HasPrefix(router.Uri, "/") {
				// 避免重复添加前缀
				url = fmt.Sprintf("%s%s", serverConfig.ContextPath, router.Uri)
			} else {
				url = fmt.Sprintf("%s/%s", serverConfig.ContextPath, router.Uri)
			}

			switch router.Method {
			case http.MethodPatch:
				ginEngine.PATCH(url, router.HandlerFunc)
			case http.MethodOptions:
				ginEngine.OPTIONS(url, router.HandlerFunc)
			case http.MethodPut:
				ginEngine.PUT(url, router.HandlerFunc)
			case http.MethodHead:
				ginEngine.HEAD(url, router.HandlerFunc)
			case http.MethodDelete:
				ginEngine.DELETE(url, router.HandlerFunc)
			case http.MethodPost:
				ginEngine.POST(url, router.HandlerFunc)
			case http.MethodGet:
				fallthrough
			default:
				ginEngine.GET(url, router.HandlerFunc)
			}
		}
	}

}
