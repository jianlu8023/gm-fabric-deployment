package router

import (
	"github.com/jianlu8023/golang-example/internal/web/handler"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/ants"
	"github.com/jianlu8023/golang-example/pkg/control/captcha"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/grpc"
	"github.com/jianlu8023/golang-example/pkg/control/http"
	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/mfa"
	"github.com/jianlu8023/golang-example/pkg/control/webrtc"
	"github.com/jianlu8023/golang-example/pkg/control/websocket"
)

// NewRouter 创建新的路由列表
//
// @description 创建并返回路由处理器列表，包含所有业务模块的路由
// @param loggerControl *logger.Control 日志控制器
// @param libp2pControl *libp2p.Control libp2p控制器
// @param grpcControl *grpc.Control gRPC控制器
// @param dockerControl *docker.Control Docker控制器
// @param datasourceControl *datasource.Control 数据源控制器
// @param websocketControl *websocket.Control WebSocket控制器
// @param httpControl *http.Control HTTP控制器
// @param captchaControl *captcha.Control 验证码控制器
// @param antsPoolControl *ants.Control 线程池控制器
// @param fabriccaControl *fabricca.Control Fabric CA控制器
// @param mfaControl *mfa.Control MFA控制器
// @return []commonhttp.RouterHandler 路由处理器列表
func NewRouter(loggerControl *logger.Control,
	libp2pControl *libp2p.Control,
	grpcControl *grpc.Control,
	dockerControl *docker.Control,
	datasourceControl *datasource.Control,
	websocketControl *websocket.Control,
	httpControl *http.Control,
	captchaControl *captcha.Control,
	antsPoolControl *ants.Control,
	mfaControl *mfa.Control,
	webrtcControl *webrtc.Control,
) []commonhttp.RouterHandler {
	webLogger := loggerControl.GenLogger(logger.ModuleWeb)

	result := make([]commonhttp.RouterHandler, 0, 64)
	baseHandler := handler.NewHandler(webLogger)
	result = append(result, baseHandler.Routers()...)

	baseService := service.NewService(webLogger)
	baseMapper := mapper.NewMapper(webLogger, datasourceControl.GetConn())

	libp2pNodeHandler := handler.NewLibp2pNodeHandler(
		baseHandler,
		service.NewLibp2pNodeService(baseService,
			mapper.NewLibp2pNodeMapper(baseMapper),
			libp2pControl,
		),
	)
	result = append(result, libp2pNodeHandler.Routers()...)

	userHandler := handler.NewUserHandler(
		baseHandler,
		service.NewUserService(baseService,
			mapper.NewUserMapper(baseMapper),
			httpControl.GetSessionManager(),
		),
	)
	result = append(result, userHandler.Routers()...)

	websocketHandler := handler.NewWebSocketHandler(
		baseHandler,
		service.NewWebSocketService(
			baseService,
			mapper.NewWebSocketMapper(baseMapper),
			websocketControl,
		),
	)
	result = append(result, websocketHandler.Routers()...)

	sseHandler := handler.NewSSEHandler(
		baseHandler,
		service.NewSSEService(
			baseService,
			mapper.NewSSEMapper(baseMapper),
		),
	)
	result = append(result, sseHandler.Routers()...)

	systemHandler := handler.NewSystemHandler(
		baseHandler,
		service.NewSystemService(
			baseService,
			mapper.NewSystemMapper(baseMapper),
		),
	)
	result = append(result, systemHandler.Routers()...)

	// 创建验证码handler
	captchaHandler := handler.NewCaptchaHandler(
		baseHandler,
		service.NewCaptchaService(baseService,
			captchaControl),
	)
	result = append(result, captchaHandler.Routers()...)

	dockerImageHandler := handler.NewDockerImageHandler(
		baseHandler,
		service.NewDockerImageService(
			baseService,
			mapper.NewDockerImageMapper(baseMapper),
			dockerControl,
			websocketControl,
			libp2pControl,
			antsPoolControl,
		),
	)
	result = append(result, dockerImageHandler.Routers()...)

	dockerNetworkHandler := handler.NewDockerNetworkHandler(
		baseHandler,
		service.NewDockerNetworkService(
			baseService,
			mapper.NewDockerNetworkMapper(baseMapper),
			dockerControl,
		),
	)
	result = append(result, dockerNetworkHandler.Routers()...)

	grpcHandler := handler.NewGrpcHandler(
		baseHandler,
		service.NewGrpcService(
			baseService,
			mapper.NewGrpcMapper(baseMapper),
			grpcControl,
		),
	)
	result = append(result, grpcHandler.Routers()...)

	fileHandler := handler.NewFileHandler(baseHandler,
		service.NewFileService(
			baseService,
			mapper.NewFileMapper(baseMapper),
			httpControl.GetUploadDir(),
			httpControl.GetUploadCacheDir(),
		),
	)
	result = append(result, fileHandler.Routers()...)

	mfaHandler := handler.NewMFAHandler(baseHandler,
		service.NewMFAService(baseService,
			mapper.NewUserMapper(baseMapper),
			mfaControl,
		),
	)
	result = append(result, mfaHandler.Routers()...)

	webRTCHandler := handler.NewWebRTCHandler(baseHandler,
		service.NewWebRTCService(baseService,
			webrtcControl,
		),
	)
	result = append(result, webRTCHandler.Routers()...)

	return result
}
