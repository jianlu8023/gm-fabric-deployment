package grpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	// "github.com/hxx258456/ccgo/grpc"
	// "github.com/hxx258456/ccgo/grpc/credentials"
	// "github.com/hxx258456/ccgo/grpc/credentials/insecure"
	// "github.com/hxx258456/ccgo/grpc/peer"
	// "gitee.com/zhaochuninhefei/gmgo/grpc"
	// "gitee.com/zhaochuninhefei/gmgo/grpc/credentials"
	// "gitee.com/zhaochuninhefei/gmgo/grpc/credentials/insecure"
	// "gitee.com/zhaochuninhefei/gmgo/grpc/peer"
)

type ClientControl struct {
	Config  *config.GrpcClientConfig
	gClient *grpc.ClientConn
	mClient pb.MessageServiceClient
	ctx     context.Context
	logger  *zap.SugaredLogger
}

type customCredential struct{}

func (c *customCredential) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to get peer from context")
	}
	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, fmt.Errorf("failed to get tls info from peer")
	}
	if len(tlsInfo.State.PeerCertificates) > 0 {
		cert := tlsInfo.State.PeerCertificates[0]
		subject := cert.Subject.String()
		fmt.Printf("客户端证书 Subject: %s\n", subject)
		// 可以在这里将客户端证书信息添加到 metadata 中
		return map[string]string{
			"client-subject": subject,
		}, nil
	}
	return nil, fmt.Errorf("没有客户端证书")
}

func (c *customCredential) RequireTransportSecurity() bool {
	return false
}

func NewClientControl(control *Control) (*ClientControl, error) {
	control.logger.Infof("[client] start new grpc client control...")
	var gClient *grpc.ClientConn
	var err error

	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(control.config.Client.MaxCallRecvMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(control.config.Client.MaxCallSendMsgSize)),
	}

	if control.tracerControl != nil {
		control.logger.Debugf("[client] starting client with tracer...")
		opts = append(opts,
			grpc.WithStatsHandler(
				otelgrpc.NewClientHandler(
					otelgrpc.WithTracerProvider(control.tracerControl.GetProvider()),
				),
			),
		)
	}

	if control.config.Client.TlsEnabled {
		logger.Debugf("[client] generate tls client server...")

		// var transportCredentials credentials.TransportCredentials
		// transportCredentials, err = credentials.NewClientTLSFromFile(control.config.Client.TlsRCACertFile,
		// 	"grpc")
		// if err != nil {
		// 	logger.Errorf("[client] generate transportCredentials err: %v", err)
		// 	return nil, err
		// }

		// 为双向TLS验证创建正确的客户端配置
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
			ServerName:             "grpc", // 必须与服务器证书的Common Name匹配
		}

		// 加载客户端证书（注意：应该使用客户端自己的证书，而不是服务端证书）
		certificates, err := tls.LoadX509KeyPair(control.config.Client.TlsCertFile, control.config.Client.TlsKeyFile)
		if err != nil {
			control.logger.Errorf("[client] failed to load client TLS certificate: %v", err)
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{certificates}

		// 加载CA证书用于验证服务器证书
		rootCaCertFile := control.config.Client.TlsRCACertFile
		if !stringer.IsBlank(rootCaCertFile) {
			caCertPool := x509.NewCertPool()
			caCert, err := os.ReadFile(rootCaCertFile)
			if err != nil {
				control.logger.Errorf("[client] failed to read CA cert file: %v", err)
				return nil, err
			}
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				control.logger.Errorf("[client] failed to append CA cert to pool")
				return nil, ErrFailAppendCert
			}
			tlsConfig.RootCAs = caCertPool
			control.logger.Debugf("[client] mutual TLS enabled, server certificate will be verified")
		} else {
			control.logger.Warnf("[client] CA cert file is not configured for mutual TLS")
			return nil, ErrNoCACert
		}

		// 创建凭证
		transportCredentials := credentials.NewTLS(tlsConfig)

		opts = append(opts, grpc.WithTransportCredentials(transportCredentials))
		// opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))

		gClient, err = grpc.NewClient(control.config.Client.Host, opts...)
		// gClient, err = grpc.Dial(clientConfig.Host, opts...)
		if err != nil {
			logger.Errorf("[client] generate tls client err: %v", err)
			return nil, err
		}
	} else {
		logger.Debugf("[client] generate no tls client server...")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		gClient, err = grpc.NewClient(control.config.Client.Host, opts...)
		// gClient, err = grpc.Dial(clientConfig.Host, opts...)
		if err != nil {
			logger.Errorf("[client] generate no tls client server err: %v", err)
			return nil, err
		}
	}
	ctx := context.WithValue(context.Background(), "id", control.config.Client.Host)
	mClient := pb.NewMessageServiceClient(gClient)
	return &ClientControl{
		Config:  control.config.Client,
		gClient: gClient,
		mClient: mClient,
		ctx:     ctx,
		logger:  control.logger,
	}, nil
}

func (c *ClientControl) Stop() error {
	c.logger.Infof("[client] grpc client stop...")
	_, _ = c.SendMessage(&pb.BaseRequest{
		MessageType: BaseShutdown,
		ClientId:    c.Config.Host,
	})

	defer func() {
		if err := c.gClient.Close(); err != nil {
			c.logger.Errorf("[client] grpc client close err: %v", err)
		}
	}()

	return nil
}

func (c *ClientControl) SendMessage(req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("[client] grpc client send message messageType %v", req.MessageType)
	return c.SendMessageBidi(req, 0)
}

func (c *ClientControl) SendMessageBidi(req *pb.BaseRequest, chunkSize int) (*pb.BaseResponse, error) {
	if chunkSize <= 0 {
		chunkSize = c.Config.ChunkSize
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Duration(c.Config.CallTimeout)*time.Minute)
	defer cancel()

	stream, err := c.mClient.SendMessageBidi(ctx)
	if err != nil {
		return nil, err
	}

	// 上行：把 request.MessageBody 按块发送（包装在 oneof 的 Data）
	bodyReader := bytes.NewReader(req.MessageBody)
	buf := make([]byte, chunkSize)
	var seq int32 = 0
	for {
		n, rerr := bodyReader.Read(buf)
		if n > 0 {
			ch := &pb.Chunk{
				MessageType: req.MessageType,
				ClientId:    req.ClientId,
				Seq:         seq,
				Last:        false,
				Payload: &pb.Chunk_Data{
					Data: buf[:n],
				},
			}
			if err := stream.Send(ch); err != nil {
				return nil, err
			}
			seq++
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return nil, rerr
		}
	}

	// 发送一个 last 标识块（payload 可为空）
	_ = stream.Send(&pb.Chunk{
		MessageType: req.MessageType,
		ClientId:    req.ClientId,
		Seq:         seq,
		Last:        true,
	})

	// 关闭发送端
	if err := stream.CloseSend(); err != nil {
		return nil, err
	}

	// 下行：解析 meta（如果有）并拼接 data
	var meta *pb.Meta
	var gotMeta bool
	var respBuf bytes.Buffer

	for {
		ch, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return nil, recvErr
		}
		if ch == nil {
			continue
		}

		// ch.Payload 是 interface{}，用类型断言处理 oneof
		switch payload := ch.Payload.(type) {
		case *pb.Chunk_Meta:
			// 服务端约定先返回 meta
			if payload != nil && payload.Meta != nil {
				meta = payload.Meta
				gotMeta = true
			}
		case *pb.Chunk_Data:
			if payload != nil && payload.Data != nil {
				if _, werr := respBuf.Write(payload.Data); werr != nil {
					return nil, werr
				}
			}
		default:
			// 未知 payload，忽略
		}
	}

	// 如果没有 meta，默认成功
	final := &pb.BaseResponse{
		Success:         true,
		ResponseCode:    200,
		ResponseMessage: "",
		Message:         respBuf.Bytes(),
	}
	if gotMeta && meta != nil {
		final.Success = meta.Success
		final.ResponseCode = meta.ResponseCode
		final.ResponseMessage = meta.ResponseMessage
	}

	return final, nil
}
