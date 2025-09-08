package grpc

import (
	"context"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/internal/proto/message"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	logger2 "github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/logger"
	"go.uber.org/zap"
	"os"
)

type Control struct {
	server *ServerControl
	client *ClientControl
	config *config.GrpcConfig
	logger *zap.SugaredLogger
}

func NewGrpcControl(grpcConfig *config.GrpcConfig, loggerControl *logger2.Control) (*Control, error) {
	grpcLogger := loggerControl.GenLogger(logger2.ModuleGrpc)
	grpcLogger.Infof("starting new grpc control...")
	serverControl, err := NewServerControl(grpcConfig.Server, grpcLogger)
	if err != nil {
		grpcLogger.Errorf("new grpc server control err: %v", err)
		return nil, err
	}
	if err = serverControl.SetUp(func(err error) {
		grpcLogger.Errorf("failedFunc grpc server start err: %v", err)
		os.Exit(1)
	}); err != nil {
		grpcLogger.Errorf("grpc server start err: %v", err)
		return nil, err
	}

	clientControl, err := NewClientControl(grpcConfig.Client, grpcLogger)
	if err != nil {
		grpcLogger.Errorf("new grpc client control err: %v", err)
		return nil, err
	}

	grpcLogger.Infof("grpc control started...")
	control := &Control{
		config: grpcConfig,
		server: serverControl,
		client: clientControl,
		logger: grpcLogger,
	}

	control.RegisterHandler("base/shutdown", func(ctx context.Context, in *message.BaseRequest) (*message.BaseResponse, error) {
		control.logger.Debugf("base/shutdown finish...")
		return &message.BaseResponse{
			Message:      []byte("success"),
			ResponseCode: 200,
		}, nil
	})
	control.RegisterHandler("base/ping", func(ctx context.Context, in *message.BaseRequest) (*message.BaseResponse, error) {
		control.logger.Debugf("base/ping finish...")
		return &message.BaseResponse{
			Message:      []byte("pong"),
			ResponseCode: 200,
		}, nil

	})

	control.printHandlers()

	return control, nil
}

func (c *Control) printHandlers() {
	for handlerName, _ := range c.server.mServer.handler.handlerMap {
		c.logger.Debugf("register handler %v", handlerName)
	}
}

func (c *Control) ClientState() (bool, string) {
	resp, err := c.Call(&message.BaseRequest{MessageType: "base/ping"})
	if err != nil {
		return false, fmt.Sprintf("%s: failed -> code:[nil], message:%s; ", c.client.Config.Host, err.Error())
	}

	if !resp.Success {
		return false, fmt.Sprintf("%s: failed -> code:%d, message:%s; ", c.client.Config.Host, resp.ResponseCode, resp.ResponseMessage)
	}

	return true, ""
}

func (c *Control) Call(req *message.BaseRequest) (*message.BaseResponse, error) {
	c.logger.Debugf("call messageType %v", req.MessageType)
	return c.client.SendMessage(req)
}

func (c *Control) RegisterHandler(handlerName string, handle func(ctx context.Context, in *message.BaseRequest) (*message.BaseResponse, error)) {
	c.server.mServer.handler.RegisterHandler(handlerName, handle)
}

func (c *Control) Shutdown() {
	_ = c.client.Stop()
	c.server.GracefulStop()
}
