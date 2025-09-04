package http

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
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
	engine := gin.Default()
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
