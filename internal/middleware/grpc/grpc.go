package grpc

import (
	"context"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/internal/proto/message"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"os"
)

type Control struct {
	server *ServerControl
	client *ClientControl
	config *config.GrpcConfig
}

func NewGrpcControl(grpcConfig *config.GrpcConfig) (*Control, error) {
	serverControl, err := NewServerControl(grpcConfig.Server)
	if err != nil {
		fmt.Printf("new server control err: %v\n", err)
		return nil, err
	}
	if err = serverControl.SetUp(func(err error) {
		fmt.Printf("failedHandle grpc server start err: %v\n", err)
		os.Exit(1)
	}); err != nil {
		fmt.Printf("grpc server start err: %v\n", err)
		return nil, err
	}

	clientControl, err := NewClientControl(grpcConfig.Client)
	if err != nil {
		fmt.Printf("new client control err: %v\n", err)
		return nil, err
	}

	control := &Control{
		config: grpcConfig,
		server: serverControl,
		client: clientControl,
	}

	control.RegisterHandler("base/shutdown", func(ctx context.Context, in *message.BaseRequest) (*message.BaseResponse, error) {
		fmt.Printf("base/shutdown bye \n")
		return &message.BaseResponse{
			Message:      []byte("success"),
			ResponseCode: 200,
		}, nil
	})
	control.RegisterHandler("base/ping", func(ctx context.Context, in *message.BaseRequest) (*message.BaseResponse, error) {
		fmt.Printf("base/ping success\n")

		return &message.BaseResponse{
			Message:      []byte("pong"),
			ResponseCode: 200,
		}, nil

	})

	for handlerName, _ := range control.server.mServer.handler.handlerMap {
		fmt.Printf("register handler %v\n", handlerName)
	}

	return control, nil

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
	fmt.Printf("call messageType %v\n", req.MessageType)
	return c.client.SendMessage(req)
}

func (c *Control) RegisterHandler(handlerName string, handle func(ctx context.Context, in *message.BaseRequest) (*message.BaseResponse, error)) {
	c.server.mServer.handler.RegisterHandler(handlerName, handle)
}

func (c *Control) Shutdown() {
	_ = c.client.Stop()
	c.server.GracefulStop()
}
