package grpc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.uber.org/zap"
)

// Control gRPC控制器
//
// @description 管理gRPC服务端和多客户端连接池，提供消息发送、健康检查和断线重连能力
// @struct
type Control struct {
	server        *ServerControl
	client        *ClientControl
	config        *config.GrpcConfig
	logger        *zap.SugaredLogger
	once          sync.Once
	tracerControl *tracer.Control
}

// StartUp 启动gRPC控制器
//
// @description 启动gRPC服务端和健康检查协程
// @param failedFunc func(err error) 启动失败回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Debugf("[grpc/control] starting grpc server...")
			c.server.StartUp(failedFunc)
			// 客户端无需显式启动，已在 NewClientControl 中完成初始化（建立连接）
			c.logger.Debugf("[grpc/control] grpc client already initialized")

			// 启动健康检查协程
			if c.client != nil {
				checkInterval := 60 * time.Second
				if c.config.HealthCheckInterval > 0 {
					checkInterval = time.Duration(c.config.HealthCheckInterval) * time.Second
				}
				go c.client.StartHealthCheckWithInterval(checkInterval)
			}
		}
	})
}

// Shutdown 关闭gRPC控制器
//
// @description 先关闭所有客户端连接，再优雅停止服务端
// @return error 错误信息
func (c *Control) Shutdown() error {
	var errs []error

	// c.client 可能为 nil（如初始化失败），需做 nil 检查避免 panic
	if c.client != nil {
		if err := c.client.Stop(); err != nil {
			c.logger.Errorf("[grpc/control] grpc client stop err: %v", err)
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

// NewGrpcControl 创建gRPC控制器
//
// @description 根据配置创建gRPC控制器，包括服务端和多客户端连接池
// @param grpcConfig *config.GrpcConfig gRPC配置
// @param loggerControl *logger.Control 日志控制器
// @param opts ...Option 可选参数
// @return *Control gRPC控制器
// @return error 错误信息
func NewGrpcControl(grpcConfig *config.GrpcConfig, loggerControl *logger.Control, opts ...Option) (*Control, error) {
	grpcConfig = check.IF[*config.GrpcConfig](grpcConfig == nil,
		getDefaultConfig(),
		grpcConfig,
	)
	if !grpcConfig.Enabled {
		return nil, errors.New("grpc is not enabled")
	}
	// 校验 Server 子配置非 nil，避免后续 NewServerControl 访问字段时空指针 panic
	if grpcConfig.Server == nil {
		return nil, errors.New("grpc server config is nil")
	}
	// Clients 允许为空（仅作为服务端运行，不主动连接其他节点）
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	// grpcLogger := loggerControl.GenLogger(logger.ModuleGrpc)
	grpcLogger := loggerControl.GenLogger("")

	grpcLogger.Infof("[grpc/control] starting new grpc control...")
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
		grpcLogger.Errorf("[grpc/control] new grpc server control err: %v", err)
		return nil, err
	}

	if err := NewClientControl(control); err != nil {
		grpcLogger.Errorf("[grpc/control] new grpc client control err: %v", err)
		return nil, err
	}

	// if err = serverControl.StartUp(failedFunc); err != nil {
	// 	grpcLogger.Errorf("grpc server start err: %v", err)
	// 	return nil, err
	// }

	grpcLogger.Infof("[grpc/control] grpc control started...")
	control.RegisterHandler(BaseShutdown, func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error) {
		control.logger.Debugf("[grpc/control] %v finish...", BaseShutdown)
		return &pb.BaseResponse{
			Success:         true,
			ResponseCode:    200,
			ResponseMessage: "success",
			Message:         []byte("success"),
		}, nil
	})
	control.RegisterHandler(BasePing, func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error) {
		control.logger.Debugf("[grpc/control] %v finish...", BasePing)
		return &pb.BaseResponse{
			Success:         true,
			ResponseCode:    200,
			ResponseMessage: "pong",
			Message:         []byte("pong"),
		}, nil
	})

	control.printHandlers()

	return control, nil
}

// printHandlers 打印已注册的处理器
func (c *Control) printHandlers() {
	c.logger.Debugf("[grpc/control] print handler...")
	c.server.mServer.handler.PrintHandler()
}

// ClientState 检查所有客户端的连接状态
//
// @description 遍历所有客户端发送ping消息，返回整体状态
// @return bool 是否所有客户端都正常
// @return string 状态描述
func (c *Control) ClientState() (bool, string) {
	c.logger.Debugf("[grpc/control] check client state...")
	if c.client == nil {
		return false, "grpc client is not initialized"
	}

	clientIDs := c.client.ListClients()
	if len(clientIDs) == 0 {
		return true, "no clients configured"
	}

	var failedMsgs []string
	for _, clientID := range clientIDs {
		resp, err := c.SendToClient(clientID, &pb.BaseRequest{MessageType: BasePing})
		if err != nil {
			failedMsgs = append(failedMsgs, fmt.Sprintf("%s: failed -> %s; ", clientID, err.Error()))
			continue
		}
		if !resp.Success {
			failedMsgs = append(failedMsgs, fmt.Sprintf("%s: failed -> code:%d, message:%s; ", clientID, resp.ResponseCode, resp.ResponseMessage))
		}
	}

	if len(failedMsgs) > 0 {
		return false, fmt.Sprintf("%v", failedMsgs)
	}

	return true, ""
}

// isResponseSuccess 判断 gRPC 响应是否成功
//
// @description 优先使用 Success 字段；若未显式设置，则使用 ResponseCode 2xx 判定为成功；若 Success=false 且 code=2xx 也视为成功（兼容性）
// @param resp *pb.BaseResponse 响应
// @return bool 是否成功
func isResponseSuccess(resp *pb.BaseResponse) bool {
	if resp == nil {
		return false
	}
	if resp.Success {
		return true
	}
	code := resp.ResponseCode
	// ResponseCode 2xx 视为成功，0（未设置）也作为成功的默认值
	return code == 0 || (code >= 200 && code < 300)
}

// Call 发送消息（广播模式，返回第一个成功的响应）
//
// @description 向所有已连接的客户端发送消息，返回第一个成功的响应
// @param req *pb.BaseRequest 请求
// @return *pb.BaseResponse 第一个成功的响应
// @return error 错误信息
func (c *Control) Call(req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("[grpc/control] call(broadcast) messageType %v", req.MessageType)
	if c.client == nil {
		return nil, errors.New("grpc client is not initialized")
	}

	// 自动填充发送方标识
	if req.ClientId == "" {
		req.ClientId = c.config.NodeID
	}

	clientIDs := c.client.ListClients()
	if len(clientIDs) == 0 {
		return nil, errors.New("no clients connected")
	}

	var lastErr error
	for _, clientID := range clientIDs {
		resp, err := c.client.SendToClient(clientID, req)
		if err != nil {
			c.logger.Debugf("[grpc/control] call to %s failed: %v", clientID, err)
			lastErr = err
			continue
		}
		if isResponseSuccess(resp) {
			return resp, nil
		}
		lastErr = fmt.Errorf("%s: code=%d, msg=%s", clientID, resp.ResponseCode, resp.ResponseMessage)
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("all clients failed")
}

// SendToClient 向指定客户端发送消息
//
// @description 自动填充ClientId后发送到指定节点
// @param clientID string 目标客户端节点标识
// @param req *pb.BaseRequest 请求
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (c *Control) SendToClient(clientID string, req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("[grpc/control] send to client %s, messageType %v", clientID, req.MessageType)
	if c.client == nil {
		return nil, errors.New("grpc client is not initialized")
	}

	// 自动填充发送方标识
	if req.ClientId == "" {
		req.ClientId = c.config.NodeID
	}

	return c.client.SendToClient(clientID, req)
}

// Broadcast 向所有已连接的客户端广播消息
//
// @description 并发向所有客户端发送消息，返回每个客户端的响应
// @param req *pb.BaseRequest 请求
// @return map[string]*pb.BaseResponse 各客户端响应，key=clientID
// @return map[string]error 各客户端错误，key=clientID
func (c *Control) Broadcast(req *pb.BaseRequest) (map[string]*pb.BaseResponse, map[string]error) {
	c.logger.Debugf("[grpc/control] broadcast messageType %v", req.MessageType)

	responses := make(map[string]*pb.BaseResponse)
	errs := make(map[string]error)

	if c.client == nil {
		errs["_"] = errors.New("grpc client is not initialized")
		return responses, errs
	}

	// 自动填充发送方标识
	if req.ClientId == "" {
		req.ClientId = c.config.NodeID
	}

	clientIDs := c.client.ListClients()
	if len(clientIDs) == 0 {
		errs["_"] = errors.New("no clients connected")
		return responses, errs
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, clientID := range clientIDs {
		wg.Add(1)
		go func(cid string) {
			defer wg.Done()
			resp, err := c.client.SendToClient(cid, req)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs[cid] = err
			} else {
				responses[cid] = resp
			}
		}(clientID)
	}

	wg.Wait()
	return responses, errs
}

// ListClients 列出所有已连接的客户端ID
//
// @return []string 客户端节点标识列表
func (c *Control) ListClients() []string {
	if c.client == nil {
		return []string{}
	}
	return c.client.ListClients()
}

// GetClient 获取指定客户端的连接信息
//
// @param clientID string 客户端节点标识
// @return *singleClient 客户端实例
// @return error 不存在时返回错误
func (c *Control) GetClient(clientID string) (*singleClient, error) {
	if c.client == nil {
		return nil, errors.New("grpc client is not initialized")
	}
	return c.client.GetClient(clientID)
}

// RegisterHandler 注册消息处理器
//
// @description 在服务端注册指定消息类型的处理函数
// @param handlerName string 处理器名称（消息类型）
// @param handle func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error) 处理函数
func (c *Control) RegisterHandler(handlerName string, handle func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error)) {
	c.logger.Debugf("[grpc/control] register handler %v", handlerName)
	c.server.mServer.handler.RegisterHandler(handlerName, handle)
}
