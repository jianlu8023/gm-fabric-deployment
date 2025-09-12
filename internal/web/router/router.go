package router

import (
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/handler"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	httpcontrol "github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/websocket"
)

type MyRouter struct {
	Name            string                 `yaml:"name"`
	Uri             string                 `yaml:"uri"`
	Method          string                 `yaml:"method"`
	HandlerFunc     func(ctx *gin.Context) `yaml:"handlerFunc"`
	Enabled         bool                   `yaml:"enabled"`
	Desc            string                 `yaml:"desc"`
	EnableJWtVerify bool                   `yaml:"enableJWtVerify"`
}

// GetName 获取路由名称
func (r *MyRouter) GetName() string {
	return r.Name
}

// GetUri 获取路由URI
func (r *MyRouter) GetUri() string {
	return r.Uri
}

// GetMethod 获取HTTP方法
func (r *MyRouter) GetMethod() string {
	return r.Method
}

// GetHandlerFunc 获取处理函数
func (r *MyRouter) GetHandlerFunc() gin.HandlerFunc {
	return r.HandlerFunc
}

// GetEnableJWtVerify 获取是否启用JWT验证
func (r *MyRouter) GetEnableJWtVerify() bool {
	return r.EnableJWtVerify
}

// IsEnabled 检查路由是否启用
func (r *MyRouter) IsEnabled() bool {
	return r.Enabled
}

// GetDesc 获取路由描述
func (r *MyRouter) GetDesc() string {
	return r.Desc
}

// NewRouter 创建新的路由列表
// @func NewRouter
// @since 1.0.0
// @description 创建并返回路由列表
// @param loggerControl *logger.Control 日志控制器
// @param libp2pControl *libp2p.Control libp2p控制器
// @param grpcControl *grpc.Control gRPC控制器
// @param dockerControl *docker.Control Docker控制器
// @param datasourceControl *datasource.Control 数据源控制器
// @returns []http.RouterHandler 路由处理器列表
// @example
// routers := NewRouter(loggerControl, libp2pControl, grpcControl, dockerControl, datasourceControl)
func NewRouter(loggerControl *logger.Control,
	libp2pControl *libp2p.Control,
	grpcControl *grpc.Control,
	dockerControl *docker.Control,
	datasourceControl *datasource.Control,
	websocketControl *websocket.Control,
	httpControl *httpcontrol.Control,
) []commonhttp.RouterHandler {
	webLogger := loggerControl.GenLogger(logger.ModuleWeb)
	baseHandler := handler.NewHandler(webLogger)
	baseService := service.NewService(webLogger)
	baseMapper := mapper.NewMapper(webLogger, datasourceControl.GetConn())

	nodeHandler := handler.NewNodeHandler(
		baseHandler,
		service.NeeNodeService(baseService,
			mapper.NewNodeMapper(baseMapper),
		),
	)
	userHandler := handler.NewUserHandler(
		baseHandler,
		service.NewUserService(baseService,
			mapper.NewUserMapper(baseMapper),
			httpControl.GetSessionManager(),
		),
	)

	websocketHandler := handler.NewWebSocketHandler(
		baseHandler,
		service.NewWebSocketService(
			baseService,
			mapper.NewWebSocketMapper(baseMapper),
			websocketControl,
		),
	)

	sseHandler := handler.NewSSEHandler(
		baseHandler,
		service.NewSSEService(
			baseService,
			mapper.NewSSEMapper(baseMapper),
		),
	)

	systemHandler := handler.NewSystemHandler(
		baseHandler,
		service.NewSystemService(
			baseService,
			mapper.NewSystemMapper(baseMapper),
		),
	)

	routers := map[string][]commonhttp.RouterHandler{
		"base": {
			&MyRouter{
				Name:   "ping",
				Uri:    "/ping",
				Method: http.MethodGet,
				HandlerFunc: func(ctx *gin.Context) {
					id := requestid.Get(ctx)
					webhttp.SuccessResponse(ctx, gin.H{
						"request_id": id,
						"message":    "pong",
					})
				},
				Enabled: true,
				Desc:    "ping",
			},
			&MyRouter{
				Name:   "health",
				Uri:    "/health",
				Method: http.MethodGet,
				HandlerFunc: func(ctx *gin.Context) {
					webhttp.SuccessResponse(ctx, "ok")
				},
				Enabled: true,
				Desc:    "health check",
			},
			&MyRouter{
				Name:        "sse",
				Uri:         "/sse",
				Method:      http.MethodGet,
				HandlerFunc: sseHandler.SSE,
				Enabled:     true,
				Desc:        "sse",
			},
		},
		"auth": {},
		"user": {
			&MyRouter{
				Name:            "registerUser",
				Uri:             "user/register",
				Method:          http.MethodPost,
				HandlerFunc:     userHandler.RegisterUserHandler,
				Enabled:         true,
				Desc:            "register a user",
				EnableJWtVerify: false,
			},
			&MyRouter{
				Name:            "loginUser",
				Uri:             "user/login",
				Method:          http.MethodPost,
				HandlerFunc:     userHandler.LoginUserHandler,
				Enabled:         true,
				Desc:            "user login",
				EnableJWtVerify: false,
			},
		},
		"node": {
			&MyRouter{
				Name:            "nodeList",
				Uri:             "node/list",
				Method:          http.MethodGet,
				HandlerFunc:     nodeHandler.NodeListHandler,
				Enabled:         true,
				Desc:            "node list",
				EnableJWtVerify: true,
			},
		},
		"websocket": {
			// WebSocket连接升级路由
			&MyRouter{
				Name:            "websocketUpgrade",
				Uri:             "ws",
				HandlerFunc:     websocketHandler.UpgradeHandler,
				Enabled:         true,
				EnableJWtVerify: true,
				Desc:            "websocket connection upgrade",
			},
			// WebSocket连接请求路由
			&MyRouter{
				Name:            "websocketConnect",
				Uri:             "ws/connect",
				Method:          http.MethodPost,
				HandlerFunc:     websocketHandler.ConnectHandler,
				Enabled:         true,
				EnableJWtVerify: true,
				Desc:            "websocket connect request",
			},
			// WebSocket断开连接路由
			&MyRouter{
				Name:            "websocketDisconnect",
				Uri:             "ws/disconnect",
				Method:          http.MethodPost,
				HandlerFunc:     websocketHandler.DisconnectHandler,
				Enabled:         true,
				EnableJWtVerify: true,
				Desc:            "websocket disconnect request",
			},
			// WebSocket发送消息路由
			&MyRouter{
				Name:            "websocketSendMessage",
				Uri:             "ws/message",
				Method:          http.MethodPost,
				HandlerFunc:     websocketHandler.SendMessageHandler,
				Enabled:         true,
				EnableJWtVerify: true,
				Desc:            "send websocket message",
			},
			// 获取WebSocket连接列表路由
			&MyRouter{
				Name:            "websocketConnectionList",
				Uri:             "ws/connections",
				Method:          http.MethodGet,
				HandlerFunc:     websocketHandler.GetConnectionListHandler,
				Enabled:         true,
				EnableJWtVerify: true,
				Desc:            "get websocket connection list",
			},
		},
		"system": {
			&MyRouter{
				Name:            "systemOverview",
				Uri:             "system/overview",
				Enabled:         true,
				EnableJWtVerify: true,
				Method:          http.MethodGet,
				Desc:            "get system overview",
				HandlerFunc:     systemHandler.GetSystemOverview,
			},
		},
	}
	result := make([]commonhttp.RouterHandler, 0, 64)

	for _, v := range routers {
		result = append(result, v...)
	}
	return result
}
