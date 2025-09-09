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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"io"
	"os"
	"time"
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

func genClientTlsConfig(clientConfig *config.GrpcClientConfig) (*tls.Config, error) {
	certificates, err := tls.LoadX509KeyPair(clientConfig.TlsCertFile, clientConfig.TlsKeyFile)
	if err != nil {
		fmt.Printf("gen server tls config error: %v", err)
		return nil, err
	}
	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(clientConfig.TlsRCACertFile)
	if err != nil {
		fmt.Printf("read root-ca err: %v", err)
		return nil, err
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, errors.New("append root-ca err")
	}
	return &tls.Config{
		ServerName:   "grpc",
		Certificates: []tls.Certificate{certificates},
		RootCAs:      ca,
	}, nil
}

func NewClientControl(clientConfig *config.GrpcClientConfig, logger *zap.SugaredLogger) (*ClientControl, error) {
	logger.Infof("start new grpc client control...")
	var gClient *grpc.ClientConn
	var err error

	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(clientConfig.MaxCallRecvMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(clientConfig.MaxCallSendMsgSize)),
	}

	if clientConfig.TlsEnabled {
		logger.Debugf("gen tls client server...")

		var transportCredentials credentials.TransportCredentials
		transportCredentials, err = credentials.NewClientTLSFromFile(clientConfig.TlsRCACertFile,
			"grpc")
		if err != nil {
			logger.Errorf("gen transportCredentials err: %v", err)
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(transportCredentials))
		// opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))

		// clientTlsConfig, err := genClientTlsConfig(clientConfig)
		// if err != nil {
		// 	fmt.Printf("gen client tls config err: %v\n", err)
		// 	return nil, err
		// }
		// transportCredentials := credentials.NewTLS(clientTlsConfig)
		// opts = append(opts, grpc.WithTransportCredentials(transportCredentials))

		gClient, err = grpc.NewClient(clientConfig.Host, opts...)
		if err != nil {
			logger.Errorf("gen tls client err: %v", err)
			return nil, err
		}
	} else {
		logger.Debugf("gen no tls client server...")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		gClient, err = grpc.NewClient(clientConfig.Host, opts...)
		if err != nil {
			logger.Errorf("gen no tls client server err: %v", err)
			return nil, err
		}
	}
	ctx := context.WithValue(context.Background(), "id", clientConfig.Host)
	mClient := pb.NewMessageServiceClient(gClient)
	return &ClientControl{
		Config:  clientConfig,
		gClient: gClient,
		mClient: mClient,
		ctx:     ctx,
		logger:  logger,
	}, nil

}

func (c *ClientControl) Stop() error {
	c.logger.Infof("grpc client stop...")
	_, _ = c.SendMessage(&pb.BaseRequest{
		MessageType: "base/shutdown",
		ClientId:    c.Config.Host,
	})

	defer func() {
		if err := c.gClient.Close(); err != nil {
			c.logger.Errorf("grpc client close err: %v", err)
		}
	}()

	return nil
}

func (c *ClientControl) SendMessage(req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("grpc client send message messageType %v", req.MessageType)
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
