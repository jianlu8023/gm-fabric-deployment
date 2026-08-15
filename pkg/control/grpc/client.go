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
	"sync/atomic"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
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
	// ctxKeyID 是用于存储客户端标识（节点ID）的 context key
	ctxKeyID contextKey = "id"
)

// singleClient 单个客户端连接
//
// @description 封装一条 gRPC 客户端连接的所有状态，包括配置、连接、stub、上下文和健康状态
// @struct
type singleClient struct {
	id      string                   // 客户端节点标识（如 "client-win"）
	config  *config.GrpcClientConfig // 连接配置（用于断线重连）
	conn    *grpc.ClientConn         // gRPC 连接
	stub    pb.MessageServiceClient  // 消息服务 stub
	ctx     context.Context          // 带节点标识的上下文
	healthy atomic.Bool              // 健康状态：true=正常，false=断线
}

// SendMessage 向该客户端发送消息
//
// @description 使用配置的 CallTimeout 创建超时上下文，调用双向流发送消息
// @param req *pb.BaseRequest 请求
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (sc *singleClient) SendMessage(req *pb.BaseRequest) (*pb.BaseResponse, error) {
	return sc.SendMessageBidi(req, 0)
}

// SendMessageWithCtx 使用指定的上下文向该客户端发送消息
//
// @description 与 SendMessage 的区别在于使用调用方传入的 context，而非内部按 CallTimeout 创建的超时上下文
// @param ctx context.Context 调用方控制的上下文
// @param req *pb.BaseRequest 请求
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (sc *singleClient) SendMessageWithCtx(ctx context.Context, req *pb.BaseRequest) (*pb.BaseResponse, error) {
	return sc.SendMessageBidiWithCtx(ctx, req, 0)
}

// SendMessageBidi 使用双向流发送消息
//
// @description 校验 CallTimeout 后创建超时上下文，调用 SendMessageBidiWithCtx
// @param req *pb.BaseRequest 请求
// @param chunkSize int 分块大小，<=0 时使用配置的 ChunkSize
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (sc *singleClient) SendMessageBidi(req *pb.BaseRequest, chunkSize int) (*pb.BaseResponse, error) {
	callTimeout := sc.config.CallTimeout
	if callTimeout <= 0 {
		callTimeout = 30
	}
	ctx, cancel := context.WithTimeout(sc.ctx, time.Duration(callTimeout)*time.Second)
	defer cancel()

	return sc.SendMessageBidiWithCtx(ctx, req, chunkSize)
}

// SendMessageBidiWithCtx 使用指定的上下文进行双向流式消息收发
//
// @description 完成上行分片发送、meta/data 下行接收与拼接，返回统一的 BaseResponse
// @param ctx context.Context 调用方控制的上下文
// @param req *pb.BaseRequest 请求
// @param chunkSize int 分块大小，<=0 时使用配置的 ChunkSize
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (sc *singleClient) SendMessageBidiWithCtx(ctx context.Context, req *pb.BaseRequest, chunkSize int) (*pb.BaseResponse, error) {
	if chunkSize <= 0 {
		chunkSize = sc.config.ChunkSize
	}

	stream, err := sc.stub.SendMessageBidi(ctx)
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

		switch payload := ch.Payload.(type) {
		case *pb.Chunk_Meta:
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

// Stop 关闭该客户端连接
//
// @description 发送 shutdown 消息后关闭连接，失败不阻塞关闭流程
// @return error 错误信息
func (sc *singleClient) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试发送 shutdown 消息，失败不阻塞关闭流程
	if _, sendErr := sc.SendMessageWithCtx(shutdownCtx, &pb.BaseRequest{
		MessageType: BaseShutdown,
		ClientId:    sc.id,
	}); sendErr != nil {
		// shutdown 消息发送失败不阻塞关闭流程
	}

	if err := sc.conn.Close(); err != nil {
		return err
	}
	return nil
}

// ClientControl 客户端控制器
//
// @description 管理多个 gRPC 客户端连接，支持多远端节点连接、健康检查和断线重连
// @struct
type ClientControl struct {
	nodeID  string                                // 本地节点标识
	clients concurrent.Map[string, *singleClient] // 多连接池，key=节点标识
	ctx     context.Context                       // 根上下文（用于健康检查协程）
	cancel  context.CancelFunc                    // 取消函数
	logger  *zap.SugaredLogger
}

// NewClientControl 创建客户端控制器
//
// @description 遍历配置中的 Clients map，为每个远端节点创建 gRPC 连接
// @param control *Control 控制器
// @return error 错误信息
func NewClientControl(control *Control) error {
	control.logger.Infof("[grpc/client] start new grpc client control...")

	clientsConfig := control.config.Clients
	if clientsConfig == nil || len(clientsConfig) == 0 {
		control.logger.Warnf("[grpc/client] no clients configured, skipping client control initialization")
		// 仍然创建 ClientControl，但 clients 为空 map
		ctx, cancel := context.WithCancel(context.Background())
		control.client = &ClientControl{
			nodeID:  control.config.NodeID,
			clients: concurrentmap.NewRWMap[string, *singleClient](),
			ctx:     ctx,
			cancel:  cancel,
			logger:  control.logger,
		}
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	control.client = &ClientControl{
		nodeID:  control.config.NodeID,
		clients: concurrentmap.NewRWMap[string, *singleClient](),
		ctx:     ctx,
		cancel:  cancel,
		logger:  control.logger,
	}

	// 遍历 clients map，为每个节点创建连接
	for clientID, cliCfg := range clientsConfig {
		sc, err := dialOne(clientID, cliCfg, control)
		if err != nil {
			control.logger.Warnf("[grpc/client] connect %s(%s) failed: %v", clientID, cliCfg.Host, err)
			continue
		}
		control.client.clients.Put(clientID, sc)
		control.logger.Infof("[grpc/client] connected to %s @ %s", clientID, cliCfg.Host)
	}

	control.logger.Infof("[grpc/client] grpc client control started, connected %d clients",
		control.client.clients.Len())
	return nil
}

// dialOne 创建单个客户端连接
//
// @description 根据配置创建 gRPC 连接，支持 TLS/GM-TLS/无TLS 三种模式
// @param clientID string 节点标识
// @param cliCfg *config.GrpcClientConfig 客户端配置
// @param control *Control 控制器
// @return *singleClient 客户端连接实例
// @return error 错误信息
func dialOne(clientID string, cliCfg *config.GrpcClientConfig, control *Control) (*singleClient, error) {
	var gClient *grpc.ClientConn
	// clientConfig := control.config.Client
	var err error

	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cliCfg.MaxCallRecvMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(cliCfg.MaxCallSendMsgSize)),
	}

	if control.tracerControl != nil {
		control.logger.Debugf("[grpc/client] starting client %s with tracer...", clientID)
		opts = append(opts,
			grpc.WithStatsHandler(
				otelgrpc.NewClientHandler(
					otelgrpc.WithTracerProvider(control.tracerControl.TracerProvider()),
					otelgrpc.WithMeterProvider(control.tracerControl.MeterProvider()),
				),
			),
		)
	}

	if cliCfg.TlsEnabled {
		if cliCfg.TlsGM {
			control.logger.Debugf("[grpc/client] generate gm tls grpc client for %s...", clientID)

			gmTlsConfig := &gmtls.Config{
				GMSupport: &gmtls.GMSupport{
					WorkMode: gmtls.ModeGMSSLOnly,
				},
				SessionTicketsDisabled: false, // 启用会话票据
				NextProtos:             []string{"h2", "http/1.1"},
				ServerName:             cliCfg.TlsServerName, // 必须与服务器证书的Common Name匹配
			}

			certFiles := cliCfg.TlsCertFile
			keyFiles := cliCfg.TlsKeyFile

			// 检查证书和密钥文件数量是否匹配且至少有两对
			if len(certFiles) != len(keyFiles) {
				control.logger.Errorf("[grpc/client] GM模式证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(certFiles), len(keyFiles))
				return nil, ErrNoCACert
			}
			if len(certFiles) < 2 {
				control.logger.Errorf("[grpc/client] GM模式至少需要两套keypair（签名和加密），当前只有 %d 套", len(certFiles))
				return nil, ErrNoCACert
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
					return nil, ErrNoCACert
				}
			}
			for i, file := range keyFiles {
				if stringer.IsBlank(file) {
					control.logger.Errorf("[grpc/client] 第%d个密钥文件路径为空", i+1)
					return nil, ErrNoCACert
				}
			}

			// 加载所有keypair
			var certificates []gmtls.Certificate
			for i := 0; i < len(certFiles); i++ {
				cert, err := gmtls.LoadX509KeyPair(certFiles[i], keyFiles[i])
				if err != nil {
					control.logger.Errorf("[grpc/client] 加载第%d套GM TLS证书失败: %v", i+1, err)
					return nil, err
				}
				certificates = append(certificates, cert)
				control.logger.Debugf("[grpc/client] 成功加载第%d套GM TLS证书: %s -> %s", i+1, certFiles[i], keyFiles[i])
			}

			// 设置证书到GM TLS配置
			gmTlsConfig.Certificates = certificates
			control.logger.Infof("[grpc/client] %s 成功加载GM模式 %d 套keypair", clientID, len(certificates))

			rootCaCertFile := cliCfg.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
				caCertPool := gmx509.NewCertPool()
				caCert, err := os.ReadFile(rootCaCertFile)
				if err != nil {
					control.logger.Errorf("[grpc/client] failed to read CA cert file: %v", err)
					return nil, err
				}
				if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
					control.logger.Errorf("[grpc/client] failed to append CA cert to pool")
					return nil, ErrFailAppendCert
				}
				gmTlsConfig.RootCAs = caCertPool
				control.logger.Debugf("[grpc/client] %s mutual TLS enabled, server certificate will be verified", clientID)
			} else {
				control.logger.Warnf("[grpc/client] CA cert file is not configured for mutual TLS")
				return nil, ErrNoCACert
			}

			// 创建凭证
			transportCredentials := gmcredentials.NewTLS(gmTlsConfig)

			opts = append(opts, grpc.WithTransportCredentials(transportCredentials))
			// opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))

			gClient, err = grpc.NewClient(cliCfg.Host, opts...)
			if err != nil {
				control.logger.Errorf("[grpc/client] generate gm tls client err: %v", err)
				return nil, err
			}
		} else {
			control.logger.Debugf("[grpc/client] generate tls client for %s...", clientID)

			// var transportCredentials credentials.TransportCredentials
			// transportCredentials, err = credentials.NewClientTLSFromFile(clientConfig.TlsRCACertFile,
			// 	"grpc")
			// if err != nil {
			// 	logger.Errorf("[client] generate transportCredentials err: %v", err)
			// 	return nil, err
			// }

			// 为双向TLS验证创建正确的客户端配置
			tlsConfig := &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
				},
				CurvePreferences: []tls.CurveID{
					tls.CurveP256, tls.X25519,
				},
				SessionTicketsDisabled: false,
				NextProtos:             []string{"h2", "http/1.1"},
				ServerName:             cliCfg.TlsServerName,
			}

			if len(cliCfg.TlsCertFile) != len(cliCfg.TlsKeyFile) {
				control.logger.Errorf("[grpc/client] TLS证书和密钥文件数量必须匹配，当前证书数量: %d, 密钥数量: %d", len(cliCfg.TlsCertFile), len(cliCfg.TlsKeyFile))
				return nil, fmt.Errorf("TLS证书和密钥文件数量必须匹配")
			}
			if len(cliCfg.TlsCertFile) == 0 {
				control.logger.Error("[grpc/client] TLS至少需要一对证书和密钥文件")
				return nil, errors.New("TLS至少需要一对证书和密钥文件")
			}

			// 加载所有客户端证书
			var certificates []tls.Certificate
			for i := 0; i < len(cliCfg.TlsCertFile); i++ {
				certFile := strings.TrimSpace(cliCfg.TlsCertFile[i])
				keyFile := strings.TrimSpace(cliCfg.TlsKeyFile[i])

				if stringer.IsBlank(certFile) {
					control.logger.Errorf("[grpc/client] 第%d个TLS证书文件路径为空", i+1)
					return nil, fmt.Errorf("第%d个TLS证书文件路径为空", i+1)
				}
				if stringer.IsBlank(keyFile) {
					control.logger.Errorf("[grpc/client] 第%d个TLS密钥文件路径为空", i+1)
					return nil, fmt.Errorf("第%d个TLS密钥文件路径为空", i+1)
				}

				cert, err := tls.LoadX509KeyPair(certFile, keyFile)
				if err != nil {
					control.logger.Errorf("[grpc/client] 加载第%d套TLS证书失败: %v", i+1, err)
					return nil, fmt.Errorf("加载第%d套TLS证书失败: %v", i+1, err)
				}
				certificates = append(certificates, cert)
			}
			tlsConfig.Certificates = certificates
			control.logger.Infof("[grpc/client] %s 成功加载 %d 套TLS客户端证书", clientID, len(certificates))

			rootCaCertFile := cliCfg.TlsRCACertFile
			if !stringer.IsBlank(rootCaCertFile) {
				caCertPool := x509.NewCertPool()
				caCert, err := os.ReadFile(rootCaCertFile)
				if err != nil {
					control.logger.Errorf("[grpc/client] failed to read CA cert file: %v", err)
					return nil, err
				}
				if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
					control.logger.Errorf("[grpc/client] failed to append CA cert to pool")
					return nil, ErrFailAppendCert
				}
				tlsConfig.RootCAs = caCertPool
				control.logger.Debugf("[grpc/client] %s mutual TLS enabled, server certificate will be verified", clientID)
			} else {
				control.logger.Warnf("[grpc/client] CA cert file is not configured for mutual TLS")
				return nil, ErrNoCACert
			}

			// 创建凭证
			transportCredentials := credentials.NewTLS(tlsConfig)

			opts = append(opts, grpc.WithTransportCredentials(transportCredentials))
			// opts = append(opts, grpc.WithPerRPCCredentials(new(customCredential)))

			gClient, err = grpc.NewClient(cliCfg.Host, opts...)
			if err != nil {
				control.logger.Errorf("[grpc/client] generate tls client err: %v", err)
				return nil, err
			}
		}
	} else {
		control.logger.Debugf("[grpc/client] generate no tls client for %s...", clientID)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		gClient, err = grpc.NewClient(cliCfg.Host, opts...)
		if err != nil {
			control.logger.Errorf("[grpc/client] generate no tls client err: %v", err)
			return nil, err
		}
	}

	ctx := context.WithValue(context.Background(), ctxKeyID, clientID)
	mClient := pb.NewMessageServiceClient(gClient)

	sc := &singleClient{
		id:     clientID,
		config: cliCfg,
		conn:   gClient,
		stub:   mClient,
		ctx:    ctx,
	}
	sc.healthy.Store(true)

	return sc, nil
}

// SendToClient 向指定客户端发送消息
//
// @description 根据客户端节点标识查找连接并发送消息
// @param clientID string 目标客户端节点标识
// @param req *pb.BaseRequest 请求
// @return *pb.BaseResponse 响应
// @return error 错误信息
func (c *ClientControl) SendToClient(clientID string, req *pb.BaseRequest) (*pb.BaseResponse, error) {
	sc, ok := c.clients.Get(clientID)
	if !ok {
		return nil, fmt.Errorf("client %s not found", clientID)
	}
	if !sc.healthy.Load() {
		return nil, fmt.Errorf("client %s is not healthy", clientID)
	}
	return sc.SendMessage(req)
}

// ListClients 列出所有已连接的客户端ID
//
// @return []string 客户端节点标识列表
func (c *ClientControl) ListClients() []string {
	return c.clients.Keys()
}

// GetClient 获取指定客户端
//
// @param clientID string 客户端节点标识
// @return *singleClient 客户端实例
// @return error 不存在时返回错误
func (c *ClientControl) GetClient(clientID string) (*singleClient, error) {
	sc, ok := c.clients.Get(clientID)
	if !ok {
		return nil, fmt.Errorf("client %s not found", clientID)
	}
	return sc, nil
}

// StopClient 关闭指定客户端连接
//
// @description 发送 shutdown 消息后关闭指定客户端的连接，并从连接池中移除
// @param clientID string 客户端节点标识
// @return error 错误信息
func (c *ClientControl) StopClient(clientID string) error {
	sc, ok := c.clients.Get(clientID)
	if !ok {
		return fmt.Errorf("client %s not found", clientID)
	}
	c.clients.Del(clientID)
	return sc.Stop()
}

// Stop 关闭所有客户端连接
//
// @description 遍历所有客户端，逐个发送 shutdown 消息并关闭连接
// @return error 错误信息
func (c *ClientControl) Stop() error {
	c.logger.Infof("[grpc/client] stopping all grpc clients...")

	// 取消健康检查协程
	c.cancel()

	var errs []error

	keys := c.clients.Keys()
	for _, clientID := range keys {
		sc, ok := c.clients.Get(clientID)
		if !ok {
			continue
		}
		c.clients.Del(clientID)
		if err := sc.Stop(); err != nil {
			c.logger.Errorf("[grpc/client] grpc client %s stop err: %v", clientID, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// StartHealthCheck 启动客户端健康检查
//
// @description 按配置的间隔时间定期检查所有客户端连接状态，间隔可通过 GrpcConfig.HealthCheckInterval 配置（秒），默认60秒
func (c *ClientControl) StartHealthCheck() {
	c.logger.Infof("[grpc/client] starting client health check...")

	checkInterval := 60 * time.Second
	// HealthCheckInterval 从 control.config 读取，这里通过 clients 的配置间接获取
	// 实际间隔在 Control 层调用时传入

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Infof("[grpc/client] stopping client health check...")
			return
		case <-ticker.C:
			c.CheckClientsHealth()
		}
	}
}

// StartHealthCheckWithInterval 启动客户端健康检查（带自定义间隔）
//
// @description 按指定的间隔时间定期检查所有客户端连接状态
// @param interval time.Duration 检查间隔
func (c *ClientControl) StartHealthCheckWithInterval(interval time.Duration) {
	c.logger.Infof("[grpc/client] starting client health check, interval: %v", interval)

	if interval <= 0 {
		interval = 60 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Infof("[grpc/client] stopping client health check...")
			return
		case <-ticker.C:
			c.CheckClientsHealth()
		}
	}
}

// CheckClientsHealth 检查所有客户端的健康状态
//
// @description 遍历所有客户端，异步发送 ping 消息检查连通性
func (c *ClientControl) CheckClientsHealth() {
	c.logger.Debugf("[grpc/client] checking clients health...")
	keys := c.clients.Keys()
	for _, clientID := range keys {
		go c.checkClientHealth(clientID)
	}
}

// checkClientHealth 检查单个客户端的健康状态
//
// @description 发送 ping 消息，失败则标记为不健康并尝试重连
// @param clientID string 客户端节点标识
func (c *ClientControl) checkClientHealth(clientID string) {
	c.logger.Debugf("[grpc/client] checking health for client %s", clientID)

	sc, ok := c.clients.Get(clientID)
	if !ok {
		return
	}

	// 使用短超时发送 ping 消息
	pingCtx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	resp, err := sc.SendMessageWithCtx(pingCtx, &pb.BaseRequest{
		MessageType: BasePing,
		ClientId:    c.nodeID,
	})
	if err != nil {
		c.logger.Warnf("[grpc/client] client %s health check failed: %v, marking as unhealthy", clientID, err)
		sc.healthy.Store(false)
		// 异步尝试重连
		go c.tryReconnect(clientID)
		return
	}

	if !resp.Success {
		c.logger.Warnf("[grpc/client] client %s health check returned failure: code=%d, msg=%s",
			clientID, resp.ResponseCode, resp.ResponseMessage)
		sc.healthy.Store(false)
		go c.tryReconnect(clientID)
		return
	}

	sc.healthy.Store(true)
	c.logger.Debugf("[grpc/client] client %s is healthy", clientID)
}

// tryReconnect 尝试重连指定客户端
//
// @description 使用指数退避策略重连，最多重试3次，重连成功后更新连接池
// @param clientID string 客户端节点标识
func (c *ClientControl) tryReconnect(clientID string) {
	c.logger.Infof("[grpc/client] trying to reconnect client %s...", clientID)

	// 获取现有客户端配置（用于重连）
	sc, ok := c.clients.Get(clientID)
	if !ok {
		c.logger.Warnf("[grpc/client] client %s not found in pool, cannot reconnect", clientID)
		return
	}

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		select {
		case <-c.ctx.Done():
			c.logger.Infof("[grpc/client] stopping reconnect for %s due to context cancel", clientID)
			return
		case <-time.After(time.Duration(1<<uint(i)) * time.Second):
		}

		// 关闭旧连接
		if sc.conn != nil {
			_ = sc.conn.Close()
		}

		// 创建新连接（复用配置）
		newSC, err := dialOne(clientID, sc.config, &Control{
			logger:        c.logger,
			tracerControl: nil, // tracer 在重连时不重新设置，避免复杂性
		})
		if err != nil {
			c.logger.Warnf("[grpc/client] reconnect attempt %d/%d failed for client %s: %v", i+1, maxRetries, clientID, err)
			continue
		}

		// 重连成功，更新连接池
		c.clients.Put(clientID, newSC)
		c.logger.Infof("[grpc/client] successfully reconnected to client %s", clientID)
		return
	}

	// 重连失败，从连接池中移除
	c.logger.Errorf("[grpc/client] failed to reconnect to client %s after %d attempts, removing from pool", clientID, maxRetries)
	c.clients.Del(clientID)
}
