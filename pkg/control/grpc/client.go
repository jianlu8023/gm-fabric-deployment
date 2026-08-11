package grpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	"google.golang.org/grpc/credentials/insecure"
	// "github.com/hxx258456/ccgo/grpc"
	// "github.com/hxx258456/ccgo/grpc/credentials"
	// "github.com/hxx258456/ccgo/grpc/credentials/insecure"
	// "github.com/hxx258456/ccgo/grpc/peer"
	// "gitee.com/zhaochuninhefei/gmgo/grpc"
	// "gitee.com/zhaochuninhefei/gmgo/grpc/credentials"
	// "gitee.com/zhaochuninhefei/gmgo/grpc/credentials/insecure"
	// "gitee.com/zhaochuninhefei/gmgo/grpc/peer"
)

// contextKey 是自定义的 context key 类型，避免与其他包使用字符串作为 key 产生冲突（Go 官方推荐做法）
type contextKey string

const (
	// ctxKeyID 是用于存储客户端标识（Host）的 context key
	ctxKeyID contextKey = "id"
)

type ClientControl struct {
	config  *config.GrpcClientConfig
	gClient *grpc.ClientConn
	mClient pb.MessageServiceClient
	ctx     context.Context
	logger  *zap.SugaredLogger
}

func NewClientControl(control *Control) error {
	control.logger.Infof("[grpc/client] start new grpc client control...")
	var gClient *grpc.ClientConn
	clientConfig := control.config.Client
	var err error

	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(clientConfig.MaxCallRecvMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(clientConfig.MaxCallSendMsgSize)),
	}

	if control.tracerControl != nil {
		control.logger.Debugf("[grpc/client] starting client with tracer...")
		opts = append(opts,
			grpc.WithStatsHandler(
				otelgrpc.NewClientHandler(
					otelgrpc.WithTracerProvider(control.tracerControl.TracerProvider()),
					otelgrpc.WithMeterProvider(control.tracerControl.MeterProvider()),
				),
			),
		)
	}

	if clientConfig.TlsEnabled {
		if clientConfig.TlsGM {
			control.logger.Debugf("[grpc/client] generate gm tls grpc client...")

			gmTlsConfig := &gmtls.Config{
				GMSupport: &gmtls.GMSupport{
					WorkMode: gmtls.ModeGMSSLOnly,
				},
				SessionTicketsDisabled: false, // 启用会话票据
				NextProtos:             []string{"h2", "http/1.1"},
				ServerName:             clientConfig.TlsServerName, // 必须与服务器证书的Common Name匹配
			}

			// GM模式需要至少两套keypair：一个签名，一个加密
			certFiles := clientConfig.TlsCertFile
			keyFiles := clientConfig.TlsKeyFile

			// 检查证书和密钥文件数量是否匹配且至少有两对
			if len(certFiles) != len(keyFiles) {
				control.logger.Errorf("[grpc/client] GM模式证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(certFiles), len(keyFiles))
				return ErrNoCACert
			}
			if len(certFiles) < 2 {
				control.logger.Errorf("[grpc/client] GM模式至少需要两套keypair（签名和加密），当前只有 %d 套", len(certFiles))
				return ErrNoCACert
			}

			// 去除文件路径的空格
			for i := range certFiles {
				certFiles[i] = strings.TrimSpace(certFiles[i])
			}
			for i := range keyFiles {
				keyFiles[i] = strings.TrimSpace(keyFiles[i])
			}

			// 验证文件路径不为空
			for i, file := range certFiles {
				if stringer.IsBlank(file) {
					control.logger.Errorf("[grpc/client] 第%d个证书文件路径为空", i+1)
					return ErrNoCACert
				}
			}
			for i, file := range keyFiles {
				if stringer.IsBlank(file) {
					control.logger.Errorf("[grpc/client] 第%d个密钥文件路径为空", i+1)
					return ErrNoCACert
				}
			}

			// 加载所有keypair
			var certificates []gmtls.Certificate
			for i := 0; i < len(certFiles); i++ {
				cert, err := gmtls.LoadX509KeyPair(certFiles[i], keyFiles[i])
				if err != nil {
					control.logger.Errorf("[grpc/client] 加载第%d套GM TLS证书失败: %v", i+1, err)
					return err
				}
				certificates = append(certificates, cert)
				control.logger.Debugf("[grpc/client] 成功加载第%d套GM TLS证书: %s -> %s", i+1, certFiles[i], keyFiles[i])
			}

			// 设置证书到GM TLS配置
			gmTlsConfig.Certificates = certificates
			control.logger.Infof("[grpc/client] 成功加载GM模式 %d 套keypair", len(certificates))

			// 加载CA证书用于验证服务器证书
			rootCaCertFile := clientConfig.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
				caCertPool := gmx509.NewCertPool()
				caCert, err := os.ReadFile(rootCaCertFile)
				if err != nil {
					control.logger.Errorf("[grpc/client] failed to read CA cert file: %v", err)
					return err
				}
				if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
					control.logger.Errorf("[grpc/client] failed to append CA cert to pool")
					return ErrFailAppendCert
				}
				gmTlsConfig.RootCAs = caCertPool
				control.logger.Debugf("[grpc/client] mutual TLS enabled, server certificate will be verified")
			} else {
				control.logger.Warnf("[grpc/client] CA cert file is not configured for mutual TLS")
				return ErrNoCACert
			}

			// 创建凭证
			transportCredentials := gmcredentials.NewTLS(gmTlsConfig)

			opts = append(opts, grpc.WithTransportCredentials(transportCredentials))
			// opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))

			gClient, err = grpc.NewClient(clientConfig.Host, opts...)
			// gClient, err = grpc.Dial(clientConfig.Host, opts...)
			if err != nil {
				control.logger.Errorf("[grpc/client] generate tls client err: %v", err)
				return err
			}
		} else {
			control.logger.Debugf("[grpc/client] generate tls client client...")

			// var transportCredentials credentials.TransportCredentials
			// transportCredentials, err = credentials.NewClientTLSFromFile(clientConfig.TlsRCACertFile,
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
				ServerName:             clientConfig.TlsServerName, // 必须与服务器证书的Common Name匹配
			}

			// 验证证书和密钥文件数量匹配
			if len(clientConfig.TlsCertFile) != len(clientConfig.TlsKeyFile) {
				control.logger.Errorf("[grpc/client] TLS证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(control.config.Client.TlsCertFile), len(control.config.Client.TlsKeyFile))
				return fmt.Errorf("TLS证书和密钥文件数量必须匹配")
			}
			if len(clientConfig.TlsCertFile) == 0 {
				control.logger.Error("[grpc/client] TLS至少需要一对证书和密钥文件")
				return errors.New("TLS至少需要一对证书和密钥文件")
			}

			// 加载所有客户端证书
			var certificates []tls.Certificate
			for i := 0; i < len(clientConfig.TlsCertFile); i++ {
				certFile := strings.TrimSpace(clientConfig.TlsCertFile[i])
				keyFile := strings.TrimSpace(clientConfig.TlsKeyFile[i])

				if stringer.IsBlank(certFile) {
					control.logger.Errorf("[grpc/client] 第%d个TLS证书文件路径为空", i+1)
					return fmt.Errorf("第%d个TLS证书文件路径为空", i+1)
				}
				if stringer.IsBlank(keyFile) {
					control.logger.Errorf("[grpc/client] 第%d个TLS密钥文件路径为空", i+1)
					return fmt.Errorf("第%d个TLS密钥文件路径为空", i+1)
				}

				cert, err := tls.LoadX509KeyPair(certFile, keyFile)
				if err != nil {
					control.logger.Errorf("[grpc/client] 加载第%d套TLS证书失败: %v", i+1, err)
					return fmt.Errorf("加载第%d套TLS证书失败: %v", i+1, err)
				}
				certificates = append(certificates, cert)
			}
			tlsConfig.Certificates = certificates
			control.logger.Infof("[grpc/client] 成功加载 %d 套TLS客户端证书", len(certificates))

			// 加载CA证书用于验证服务器证书
			rootCaCertFile := clientConfig.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
				caCertPool := x509.NewCertPool()
				caCert, err := os.ReadFile(rootCaCertFile)
				if err != nil {
					control.logger.Errorf("[grpc/client] failed to read CA cert file: %v", err)
					return err
				}
				if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
					control.logger.Errorf("[grpc/client] failed to append CA cert to pool")
					return ErrFailAppendCert
				}
				tlsConfig.RootCAs = caCertPool
				control.logger.Debugf("[grpc/client] mutual TLS enabled, server certificate will be verified")
			} else {
				control.logger.Warnf("[grpc/client] CA cert file is not configured for mutual TLS")
				return ErrNoCACert
			}

			// 创建凭证
			transportCredentials := credentials.NewTLS(tlsConfig)

			opts = append(opts, grpc.WithTransportCredentials(transportCredentials))
			// opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))

			gClient, err = grpc.NewClient(clientConfig.Host, opts...)
			// gClient, err = grpc.Dial(clientConfig.Host, opts...)
			if err != nil {
				control.logger.Errorf("[grpc/client] generate tls client err: %v", err)
				return err
			}
		}
	} else {
		control.logger.Debugf("[grpc/client] generate no tls client server...")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		gClient, err = grpc.NewClient(clientConfig.Host, opts...)
		// gClient, err = grpc.Dial(clientConfig.Host, opts...)
		if err != nil {
			control.logger.Errorf("[grpc/client] generate no tls client server err: %v", err)
			return err
		}
	}
	ctx := context.WithValue(context.Background(), ctxKeyID, clientConfig.Host)
	mClient := pb.NewMessageServiceClient(gClient)
	control.client = &ClientControl{
		config:  clientConfig,
		gClient: gClient,
		mClient: mClient,
		ctx:     ctx,
		logger:  control.logger,
	}
	return nil
}

func (c *ClientControl) Stop() error {
	c.logger.Infof("[grpc/client] grpc client stop...")

	// 使用独立的短超时上下文发送 shutdown 消息，避免服务端不可达时阻塞过久（CallTimeout 默认分钟级）
	// 此处使用 context.Background() 作为父上下文，防止 c.ctx 已被取消导致 shutdown 消息无法发出
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试发送 shutdown 消息，失败不阻塞关闭流程
	if _, sendErr := c.SendMessageWithCtx(shutdownCtx, &pb.BaseRequest{
		MessageType: BaseShutdown,
		ClientId:    c.config.Host,
	}); sendErr != nil {
		c.logger.Warnf("[grpc/client] send shutdown message err: %v", sendErr)
	}

	if err := c.gClient.Close(); err != nil {
		c.logger.Errorf("[grpc/client] grpc client close err: %v", err)
		return err
	}
	return nil
}

func (c *ClientControl) SendMessage(req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("[grpc/client] grpc client send message messageType %v", req.MessageType)
	return c.SendMessageBidi(req, 0)
}

// SendMessageWithCtx 使用指定的上下文发送消息，适用于需要独立超时控制的场景（如 Stop 时发送 shutdown 消息）
//
// @description 与 SendMessage 的区别在于使用调用方传入的 context，而非内部按 CallTimeout 创建的超时上下文
// @param ctx context.Context 调用方控制的上下文
// @param req *pb.BaseRequest 请求参数
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (c *ClientControl) SendMessageWithCtx(ctx context.Context, req *pb.BaseRequest) (*pb.BaseResponse, error) {
	c.logger.Debugf("[grpc/client] grpc client send message with ctx messageType %v", req.MessageType)
	return c.SendMessageBidiWithCtx(ctx, req, 0)
}

func (c *ClientControl) SendMessageBidi(req *pb.BaseRequest, chunkSize int) (*pb.BaseResponse, error) {
	// 校验 CallTimeout，若配置未设置（默认 0）或为非正值，则使用默认值 5 分钟，避免上下文立即过期
	// 注意：CallTimeout 的单位为分钟
	callTimeout := c.config.CallTimeout
	if callTimeout <= 0 {
		callTimeout = 5 // 默认 5 分钟
	}
	ctx, cancel := context.WithTimeout(c.ctx, time.Duration(callTimeout)*time.Minute)
	defer cancel()

	return c.SendMessageBidiWithCtx(ctx, req, chunkSize)
}

// SendMessageBidiWithCtx 使用指定的上下文进行双向流式消息收发，是 SendMessageBidi 的核心实现
//
// @description 完成上行分片发送、meta/data 下行接收与拼接，返回统一的 BaseResponse
// @param ctx context.Context 调用方控制的上下文
// @param req *pb.BaseRequest 请求参数
// @param chunkSize int 分块大小，<=0 时使用配置的 ChunkSize
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (c *ClientControl) SendMessageBidiWithCtx(ctx context.Context, req *pb.BaseRequest, chunkSize int) (*pb.BaseResponse, error) {
	if chunkSize <= 0 {
		chunkSize = c.config.ChunkSize
	}

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
