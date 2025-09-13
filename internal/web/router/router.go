package router

import (
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/handler"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/service"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/captcha"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	httpcontrol "github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/websocket"
)

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
	captchaControl *captcha.Control,
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

	// 创建验证码handler
	captchaHandler := handler.NewCaptchaHandler(
		baseHandler,
		service.NewCaptchaService(baseService,
			captchaControl),
	)

	result := make([]commonhttp.RouterHandler, 0, 64)
	result = append(result, baseHandler.Routers()...)
	result = append(result, userHandler.Routers()...)
	result = append(result, captchaHandler.Routers()...)
	result = append(result, nodeHandler.Routers()...)
	result = append(result, sseHandler.Routers()...)
	result = append(result, systemHandler.Routers()...)
	result = append(result, websocketHandler.Routers()...)

	return result
}
