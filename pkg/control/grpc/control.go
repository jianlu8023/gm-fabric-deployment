package grpc

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.uber.org/zap"
)

type Control struct {
	server        *ServerControl
	client        *ClientControl
	config        *config.GrpcConfig
	logger        *zap.SugaredLogger
	once          sync.Once
	tracerControl *tracer.Control
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Debugf("[control] starting grpc server...")
			c.server.StartUp(failedFunc)
			// 客户端无需显式启动，已在 NewClientControl 中完成初始化（建立连接）
			c.logger.Debugf("[control] grpc client already initialized")
		}
	})
}

func (c *Control) Shutdown() error {
	var errs []error

	// c.client 可能为 nil（如初始化失败），需做 nil 检查避免 panic
	if c.client != nil {
		if err := c.client.Stop(); err != nil {
			c.logger.Errorf("[control] grpc client stop err: %v", err)
			errs = append(errs, err)
		}
	}

	// c.server 可能为 nil，需做 nil 检查避免 panic
	if c.server != nil {
		c.server.GracefulStop()
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func NewGrpcControl(grpcConfig *config.GrpcConfig, loggerControl *logger.Control, opts ...Option) (*Control, error) {
	grpcConfig = check.IF[*config.GrpcConfig](grpcConfig == nil,
		getDefaultConfig(),
		grpcConfig,
	)
	if !grpcConfig.Enabled {
		return nil, errors.New("grpc is not enabled")
	}
	// 校验 Server/Client 子配置非 nil，避免后续 NewServerControl/NewClientControl 访问字段时空指针 panic
	if grpcConfig.Server == nil {
		return nil, errors.New("grpc server config is nil")
	}
	if grpcConfig.Client == nil {
		return nil, errors.New("grpc client config is nil")
	}
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	grpcLogger := loggerControl.GenLogger(logger.ModuleGrpc)

	grpcLogger.Infof("[control] starting new grpc control...")
	control := &Control{
		config: grpcConfig,
		// server: serverControl,
		// client: clientControl,
		logger: grpcLogger,
	}

	for _, opt := range opts {
		opt(control)
	}

	if err := NewServerControl(control); err != nil {
		grpcLogger.Errorf("[control] new grpc server control err: %v", err)
		return nil, err
	}

	if err := NewClientControl(control); err != nil {
		grpcLogger.Errorf("[control] new grpc client control err: %v", err)
		return nil, err
	}

	// if err = serverControl.StartUp(failedFunc); err != nil {
	// 	grpcLogger.Errorf("grpc server start err: %v", err)
	// 	return nil, err
	// }

	grpcLogger.Infof("[control] grpc control started...")
	control.RegisterHandler(BaseShutdown, func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error) {
		control.logger.Debugf("[control] %v finish...", BaseShutdown)
		return &pb.BaseResponse{
			Message:      []byte("success"),
			ResponseCode: 200,
		}, nil
	})
	control.RegisterHandler(BasePing, func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error) {
		control.logger.Debugf("[control] %v finish...", BasePing)
		return &pb.BaseResponse{
			Message:      []byte("pong"),
			ResponseCode: 200,
		}, nil
	})

	control.printHandlers()

	return control, nil
}

func (c *Control) printHandlers() {
	c.logger.Debugf("[control] print handler...")
	c.server.mServer.handler.PrintHandler()
}

func (c *Control) ClientState() (bool, string) {
	c.logger.Debugf("[control] check client state...")
	// c.client 可能为 nil（如初始化失败），需做 nil 检查避免 panic
	if c.client == nil {
		return false, "grpc client is not initialized"
	}
	resp, err := c.Call(&pb.BaseRequest{MessageType: BasePing})
	if err != nil {
		return false, fmt.Sprintf("%s: failed -> code:[nil], message:%s; ", c.client.config.Host, err.Error())
	}

	if !resp.Success {
		return false, fmt.Sprintf("%s: failed -> code:%d, message:%s; ", c.client.config.Host, resp.ResponseCode, resp.ResponseMessage)
	}

	return true, ""
}

func (c *Control) Call(req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("[control] call messageType %v", req.MessageType)
	// c.client 可能为 nil（如初始化失败），需做 nil 检查避免 panic
	if c.client == nil {
		return nil, errors.New("grpc client is not initialized")
	}
	return c.client.SendMessage(req)
}

func (c *Control) RegisterHandler(handlerName string, handle func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error)) {
	c.logger.Debugf("[control] register handler %v", handlerName)
	c.server.mServer.handler.RegisterHandler(handlerName, handle)
}
