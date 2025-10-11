package grpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/tjfoc/gmsm/gmtls"
	"github.com/tjfoc/gmsm/gmtls/gmcredentials"
	gmx509 "github.com/tjfoc/gmsm/x509"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	// "github.com/hxx258456/ccgo/grpc"
	// "github.com/hxx258456/ccgo/grpc/credentials"
	// "gitee.com/zhaochuninhefei/gmgo/grpc"
	// "gitee.com/zhaochuninhefei/gmgo/grpc/credentials"
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

func NewServerControl(control *Control) error {
	control.logger.Infof("[server] start new server control...")
	var gServer *grpc.Server
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(control.config.Server.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(control.config.Server.MaxSendMsgSize),
	}
	if control.tracerControl != nil {
		control.logger.Debugf("[server] starting server with tracer...")
		opts = append(opts, grpc.StatsHandler(
			otelgrpc.NewServerHandler(
				otelgrpc.WithTracerProvider(control.tracerControl.TracerProvider()),
				otelgrpc.WithMeterProvider(control.tracerControl.MeterProvider()),
			),
		))
	}
	if control.config.Server.TlsEnabled {
		if control.config.Server.TlsGM {
			control.logger.Debugf("[server] generate gm tls grpc server...")

			gmTlsConfig := &gmtls.Config{
				GMSupport: &gmtls.GMSupport{
					WorkMode: gmtls.ModeGMSSLOnly,
				},
				SessionTicketsDisabled: false,                      // 启用会话票据
				NextProtos:             []string{"h2", "http/1.1"}, // 设置支持的协议
			}
			// TODO 差 加载两套keypair 一套签名 一套加密

			// 加载并配置CA证书用于验证客户端证书
			rootCaCertFile := control.config.Server.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
				caCertPool := gmx509.NewCertPool()
				caCert, err := os.ReadFile(rootCaCertFile)
				if err != nil {
					control.logger.Errorf("[server] failed to read CA cert file: %v", err)
					return err
				}
				if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
					control.logger.Errorf("[server] failed to append CA cert to pool")
					return ErrFailAppendCert
				}
				// 要求并验证客户端证书 - 双向TLS的关键设置
				gmTlsConfig.ClientCAs = caCertPool
				gmTlsConfig.ClientAuth = gmtls.RequireAndVerifyClientCert
				control.logger.Debugf("[server] mutual TLS enabled with client certificate verification")
			} else {
				control.logger.Warnf("[server] CA cert file is not configured for mutual TLS")
				return ErrNoCACert
			}

			// 创建凭证
			transportCredentials := gmcredentials.NewTLS(gmTlsConfig)

			// transportCredentials, err := credentials.NewServerTLSFromFile(control.config.Server.TlsCertFile, control.config.Server.TlsKeyFile)
			// if err != nil {
			// 	control.logger.Errorf("[server] generate transportCredentials err: %v", err)
			// 	return nil, err
			// }
			opts = append(opts, grpc.Creds(transportCredentials))

			gServer = grpc.NewServer(opts...)
		} else {
			control.logger.Debugf("[server] generate tls grpc server...")

			// 为双向TLS验证创建正确的配置
			tlsConfig := &tls.Config{
				MinVersion: tls.VersionTLS12, // 设置最低TLS版本
				MaxVersion: tls.VersionTLS13, // 设置最高TLS版本
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_AES_128_GCM_SHA256,       // secure 1.3
					tls.TLS_AES_256_GCM_SHA384,       // secure 1.3
					tls.TLS_CHACHA20_POLY1305_SHA256, // secure 1.3
				},
				CurvePreferences: []tls.CurveID{
					tls.CurveP256, tls.X25519,
				},
				SessionTicketsDisabled: false, // 启用会话票据
				NextProtos:             []string{"h2", "http/1.1"},
			}

			// 加载服务端证书
			certificates, err := tls.LoadX509KeyPair(control.config.Server.TlsCertFile, control.config.Server.TlsKeyFile)
			if err != nil {
				control.logger.Errorf("[server] failed to load TLS certificate: %v", err)
				return err
			}
			tlsConfig.Certificates = []tls.Certificate{certificates}

			// 加载并配置CA证书用于验证客户端证书
			rootCaCertFile := control.config.Server.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
				caCertPool := x509.NewCertPool()
				caCert, err := os.ReadFile(rootCaCertFile)
				if err != nil {
					control.logger.Errorf("[server] failed to read CA cert file: %v", err)
					return err
				}
				if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
					control.logger.Errorf("[server] failed to append CA cert to pool")
					return ErrFailAppendCert
				}
				// 要求并验证客户端证书 - 双向TLS的关键设置
				tlsConfig.ClientCAs = caCertPool
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
				control.logger.Debugf("[server] mutual TLS enabled with client certificate verification")
			} else {
				control.logger.Warnf("[server] CA cert file is not configured for mutual TLS")
				return ErrNoCACert
			}

			// 创建凭证
			transportCredentials := credentials.NewTLS(tlsConfig)

			// transportCredentials, err := credentials.NewServerTLSFromFile(control.config.Server.TlsCertFile, control.config.Server.TlsKeyFile)
			// if err != nil {
			// 	control.logger.Errorf("[server] generate transportCredentials err: %v", err)
			// 	return nil, err
			// }
			opts = append(opts, grpc.Creds(transportCredentials))

			gServer = grpc.NewServer(opts...)
		}
	} else {
		control.logger.Debugf("[server] generate no tls grpc server...")
		gServer = grpc.NewServer(opts...)
	}

	messageServer := &server{
		handler: &MessageHandler{
			handlerMap: make(map[string]func(ctx context.Context, in *pb.BaseRequest) (*pb.BaseResponse, error)),
		},
		serverConfig: control.config.Server,
		logger:       control.logger,
	}

	pb.RegisterMessageServiceServer(gServer, messageServer)
	control.server = &ServerControl{
		Config:  control.config.Server,
		gServer: gServer,
		mServer: messageServer,
		logger:  control.logger,
	}
	return nil
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
