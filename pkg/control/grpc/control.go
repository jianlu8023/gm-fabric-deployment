package grpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc/pb"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	server *ServerControl
	client *ClientControl
	config *config.GrpcConfig
	logger *zap.SugaredLogger
	once   sync.Once
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting server grpc server...")
		c.server.StartUp(failedFunc)
		c.logger.Debugf("[control] starting client grpc server...")
	})
}

func (c *Control) Shutdown() error {
	_ = c.client.Stop()
	c.server.GracefulStop()
	return nil
}

func NewGrpcControl(grpcConfig *config.GrpcConfig, loggerControl *mylogger.Control) (*Control, error) {
	grpcLogger := loggerControl.GenLogger(mylogger.ModuleGrpc)
	grpcLogger.Infof("[control] starting new grpc control...")
	serverControl, err := NewServerControl(grpcConfig.Server, grpcLogger)
	if err != nil {
		grpcLogger.Errorf("[control] new grpc server control err: %v", err)
		return nil, err
	}

	// if err = serverControl.StartUp(failedFunc); err != nil {
	// 	grpcLogger.Errorf("grpc server start err: %v", err)
	// 	return nil, err
	// }

	clientControl, err := NewClientControl(grpcConfig.Client, grpcLogger)
	if err != nil {
		grpcLogger.Errorf("[control] new grpc client control err: %v", err)
		return nil, err
	}

	grpcLogger.Infof("[control] grpc control started...")
	control := &Control{
		config: grpcConfig,
		server: serverControl,
		client: clientControl,
		logger: grpcLogger,
	}

	control.RegisterHandler(BaseShutdown, func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error) {
		control.logger.Debugf("[control] base/shutdown finish...")
		return &pb.BaseResponse{
			Message:      []byte("success"),
			ResponseCode: 200,
		}, nil
	})
	control.RegisterHandler(BasePing, func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error) {
		control.logger.Debugf("[control] base/ping finish...")
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
	for handlerName := range c.server.mServer.handler.handlerMap {
		c.logger.Debugf("[control] register handler %v", handlerName)
	}
}

func (c *Control) ClientState() (bool, string) {
	c.logger.Debugf("[control] check client state...")
	resp, err := c.Call(&pb.BaseRequest{MessageType: BasePing})
	if err != nil {
		return false, fmt.Sprintf("%s: failed -> code:[nil], message:%s; ", c.client.Config.Host, err.Error())
	}

	if !resp.Success {
		return false, fmt.Sprintf("%s: failed -> code:%d, message:%s; ", c.client.Config.Host, resp.ResponseCode, resp.ResponseMessage)
	}

	return true, ""
}

func (c *Control) Call(req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("[control] call messageType %v", req.MessageType)
	return c.client.SendMessage(req)
}

func (c *Control) RegisterHandler(handlerName string, handle func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error)) {
	c.logger.Debugf("[control] register handler %v", handlerName)
	c.server.mServer.handler.RegisterHandler(handlerName, handle)
}
