package grpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"net"
	"os"
)

type MessageHandler struct {
	handlerMap map[string]func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error)
}

func (h *MessageHandler) RegisterHandler(path string, handle func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error)) {
	h.handlerMap[path] = handle
}

func (h *MessageHandler) GetHandler(path string) (func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error), error) {
	if handle, exists := h.handlerMap[path]; exists {
		return handle, nil
	}
	return nil, fmt.Errorf("the processor corresponding to protocol '%s' is not registered", path)
}

type server struct {
	pb.UnimplementedMessageServiceServer
	handler      *MessageHandler
	serverConfig *config.GrpcServerConfig
	logger       *zap.SugaredLogger
}

func (s *server) SendMessageBidi(stream pb.MessageService_SendMessageBidiServer) error {
	s.logger.Debugf("[server] starting send message bidi stream...")
	// 收集上行数据到 buffer（注意：若 payload 极大，这里会占用内存）
	var buf bytes.Buffer
	var msgType, clientID string
	var seenFirst bool

	// 接收客户端上行（直到客户端 CloseSend 导致 io.EOF）
	for {
		ch, err := stream.Recv()
		if err == io.EOF {
			// 客户端已上行完毕
			break
		}
		if err != nil {
			// 读取出错，直接返回错误（gRPC 会把错误传回客户端）
			return err
		}
		if ch == nil {
			continue
		}

		// 记录第一次出现的 message_type / client_id
		if !seenFirst {
			msgType = ch.MessageType
			clientID = ch.ClientId
			seenFirst = true
		}

		// oneof payload: 可能是 data 或 meta（如果客户端发送 meta，这里忽略）
		switch payload := ch.Payload.(type) {
		case *pb.Chunk_Data:
			if payload != nil && len(payload.Data) > 0 {
				if _, werr := buf.Write(payload.Data); werr != nil {
					// 写入 buffer 失败，向客户端回报 meta 后返回错误
					_ = stream.Send(&pb.Chunk{
						MessageType: msgType,
						ClientId:    clientID,
						Seq:         -1,
						Payload: &pb.Chunk_Meta{
							Meta: &pb.Meta{
								Success:         false,
								ResponseCode:    500,
								ResponseMessage: "server internal buffer write error: " + werr.Error(),
							},
						},
					})
					return werr
				}
			}
		case *pb.Chunk_Meta:
			// 客户端如果在上行发送 meta，我们可以记录或忽略（当前忽略）
		default:
			// 未知 payload，忽略或记录日志
		}
	}

	// 构造 BaseRequest 并调用 handlerMap（保持你原有 handler 用法）
	req := &pb.BaseRequest{
		MessageType: msgType,
		ClientId:    clientID,
		MessageBody: buf.Bytes(),
	}

	handler, err := s.handler.GetHandler(req.MessageType)
	if err != nil {
		// 返回 meta 表示没有 handler，并结束流
		_ = stream.Send(&pb.Chunk{
			MessageType: req.MessageType,
			ClientId:    req.ClientId,
			Seq:         -1,
			Payload: &pb.Chunk_Meta{
				Meta: &pb.Meta{
					Success:         false,
					ResponseCode:    495,
					ResponseMessage: err.Error(),
				},
			},
		})
		return nil
	}

	// 调用业务处理器
	resp, herr := handler(stream.Context(), req)
	if herr != nil {
		// 处理器报错，返回 meta 并结束
		_ = stream.Send(&pb.Chunk{
			MessageType: req.MessageType,
			ClientId:    req.ClientId,
			Seq:         -1,
			Payload: &pb.Chunk_Meta{
				Meta: &pb.Meta{
					Success:         false,
					ResponseCode:    500,
					ResponseMessage: herr.Error(),
				},
			},
		})
		return nil
	}

	// 发送 meta（将 handler 返回的 success/code/pb 放入 meta）
	_ = stream.Send(&pb.Chunk{
		MessageType: req.MessageType,
		ClientId:    req.ClientId,
		Seq:         -1,
		Payload: &pb.Chunk_Meta{
			Meta: &pb.Meta{
				Success:         resp.Success,
				ResponseCode:    resp.ResponseCode,
				ResponseMessage: resp.ResponseMessage,
			},
		},
	})

	// 如果 handler 没有携带数据（nil 或长度 0），直接结束
	if resp == nil || len(resp.Message) == 0 {
		return nil
	}

	// 分片下发 resp.Message（data payload）
	var outSeq int32 = 0
	total := len(resp.Message)
	sent := 0
	for sent < total {
		end := sent + s.serverConfig.ChunkSize
		if end > total {
			end = total
		}
		piece := resp.Message[sent:end]

		err = stream.Send(&pb.Chunk{
			MessageType: req.MessageType,
			ClientId:    req.ClientId,
			Seq:         outSeq,
			Last:        end == total,
			Payload: &pb.Chunk_Data{
				Data: piece,
			},
		})
		if err != nil {
			// 发送出错时结束（客户端会收到错误）
			return err
		}
		outSeq++
		sent = end
	}

	// 全部发送完毕，返回 nil 让流正常结束
	return nil
}

type ServerControl struct {
	Config  *config.GrpcServerConfig
	gServer *grpc.Server
	mServer *server
	logger  *zap.SugaredLogger
}

func genServerTlsConfig(serverConfig *config.GrpcServerConfig) (*tls.Config, error) {
	certificates, err := tls.LoadX509KeyPair(serverConfig.TlsCertFile, serverConfig.TlsKeyFile)
	if err != nil {
		fmt.Printf("gen server tls config error: %v", err)
		return nil, err
	}
	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(serverConfig.TlsRCACertFile)
	if err != nil {
		fmt.Printf("read root-ca err: %v", err)
		return nil, err
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, errors.New("append root-ca err")
	}
	return &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,  // 要求客户端证书
		Certificates: []tls.Certificate{certificates}, // 服务端证书
		ClientCAs:    ca,                              // 根证书
	}, nil
}

func NewServerControl(serverConfig *config.GrpcServerConfig, logger *zap.SugaredLogger) (*ServerControl, error) {
	logger.Infof("[server] start new server control...")
	var gServer *grpc.Server
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(serverConfig.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(serverConfig.MaxSendMsgSize),
	}
	if serverConfig.TlsEnabled {
		logger.Debugf("[server] generate tls grpc server...")

		transportCredentials, err := credentials.NewServerTLSFromFile(serverConfig.TlsCertFile, serverConfig.TlsKeyFile)
		if err != nil {
			logger.Errorf("[server] generate transportCredentials err: %v", err)
			return nil, err
		}
		opts = append(opts, grpc.Creds(transportCredentials))

		// serverTlsConfig, err := genServerTlsConfig(serverConfig)
		// if err != nil {
		// 	fmt.Printf("gen serverTlsConfig err: %v\n", err)
		// 	return nil, err
		// }
		// transportCredentials := credentials.NewTLS(serverTlsConfig)
		// opts = append(opts, grpc.Creds(transportCredentials))

		gServer = grpc.NewServer(opts...)
	} else {
		logger.Debugf("[server] generate no tls grpc server...")
		gServer = grpc.NewServer(opts...)
	}

	messageServer := &server{
		handler: &MessageHandler{
			handlerMap: make(map[string]func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error)),
		},
		serverConfig: serverConfig,
		logger:       logger,
	}

	pb.RegisterMessageServiceServer(gServer, messageServer)
	return &ServerControl{
		Config:  serverConfig,
		gServer: gServer,
		mServer: messageServer,
		logger:  logger,
	}, nil
}

func (s *ServerControl) StartUp(failedFunc func(err error)) {
	s.logger.Infof("[server] start grpc server on %v", s.Config.Host)
	listen, err := net.Listen("tcp", s.Config.Host)
	if err != nil {
		s.logger.Errorf("[server] grpc generate listener failed: %v", err)
		failedFunc(err)
	}
	go func() {
		if err := s.gServer.Serve(listen); err != nil {
			s.logger.Errorf("[server] grpc server start failed: %v", err)
			failedFunc(err)
		}
	}()
}

func (s *ServerControl) Stop() {
	s.logger.Infof("[server] grpc server stop...")
	s.gServer.Stop()
}

func (s *ServerControl) GracefulStop() {
	s.logger.Infof("[server] grpc server graceful stop...")
	s.gServer.GracefulStop()
}
