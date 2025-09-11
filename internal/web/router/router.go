package router

import (
	"net/http"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/handler"
	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
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
				Name:   "sse",
				Uri:    "/sse",
				Method: http.MethodGet,
				HandlerFunc: func(ctx *gin.Context) {
					sse.Encode(ctx.Writer, sse.Event{
						Event: "message",
						Data:  "some data\nmore data",
					})

					// also a complex type, like a map, a struct or a slice
					sse.Encode(ctx.Writer, sse.Event{
						Id:    "124",
						Event: "message",
						Data: map[string]interface{}{
							"user":    "manu",
							"date":    time.Now().Unix(),
							"content": "hi!",
						},
					})
				},
				Enabled: true,
				Desc:    "sse",
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
	}
	result := make([]commonhttp.RouterHandler, 0, 64)

	for _, v := range routers {
		result = append(result, v...)
	}
	return result
}
