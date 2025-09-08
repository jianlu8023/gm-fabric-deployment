package main

import (
	"fmt"
	mygrpc "github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/grpc"
	mylibp2p "github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/libp2p"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/pkg/middleware/logger"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/jianlu8023/gm-fabric-deployment/internal/proto/message"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/pidfile"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var (
	version string
)

func main() {

	fmt.Printf("start client version %s\n", version)
	runOS := runtime.GOOS
	switch runOS {
	case "windows":
		fmt.Println("Running on Windows...")
	case "linux":
		fmt.Println("Running on Linux...")
		if err := pidfile.CreateOrUpdatePIDFile("client.pid"); err != nil {
			fmt.Printf("generate pid file failed: %v\n", err)
			return
		}
		defer func() {
			pidfile.ReleasePID()
		}()
	case "darwin": // macOS
		fmt.Println("Running on macOS...")
		if err := pidfile.CreateOrUpdatePIDFile("client.pid"); err != nil {
			fmt.Printf("generate pid file failed: %v\n", err)
			return
		}
		defer func() {
			pidfile.ReleasePID()
		}()
	default:
		fmt.Printf("Running on an unknown operating system: %s\n", runOS)
	}

	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	loggerControl := mylogger.NewLoggerControl(configControl.GetLoggerConfig())
	mainLogger := loggerControl.GenLogger("main")

	mainLogger.Infof("starting grpc server...")
	grpcConfig := &config.GrpcConfig{
		Server: &config.GrpcServerConfig{
			Host:           "127.0.0.1:65533",
			MaxRecvMsgSize: 5242880,
			MaxSendMsgSize: 5242880,
			ChunkSize:      4194304,
			TlsEnabled:     true,
			TlsCertFile:    "./certs/gserver.crt",
			TlsKeyFile:     "./certs/gserver.key",
			TlsRCACertFile: "./certs/root-ca.crt",
		},
		Client: &config.GrpcClientConfig{
			Host:               "127.0.0.1:65534",
			MaxCallRecvMsgSize: 5242880,
			MaxCallSendMsgSize: 5242880,
			ChunkSize:          4194304,
			CallTimeout:        99999,
			TlsEnabled:         true,
			TlsCertFile:        "./certs/gclient.crt",
			TlsKeyFile:         "./certs/gclient.key",
			TlsRCACertFile:     "./certs/root-ca.crt",
		},
	}
	control, err := mygrpc.NewGrpcControl(grpcConfig, loggerControl)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer control.Shutdown()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for range ticker.C {
			response, err := control.Call(&message.BaseRequest{
				MessageType: "base/ping",
			})
			if err != nil {
				mainLogger.Errorf("send ping message err: %v", err)
			} else {
				mainLogger.Debugf("send ping message success %v", response)
			}
		}
	}()

	libp2pConfig := &config.Libp2pConfig{
		ListenAddr: []string{
			"/ip4/0.0.0.0/tcp/2001",
		},
		ProtocolID:    "/gm-fabric/chat/1.0.0",
		ServiceTag:    "gm-fabric-deployment",
		BootstrapList: []string{},
	}

	mainLogger.Infof("starting libp2p server...")
	libp2pControl, err := mylibp2p.NewLibp2pControl(libp2pConfig, loggerControl)
	if err != nil {
		mainLogger.Errorf("create libp2p control failed: %v", err)
		return
	}
	libp2pControl.Start(func(err error) {
		mainLogger.Errorf("start libp2p failed: %v", err)
		quit <- os.Interrupt
	})
	defer func() {
		if err := libp2pControl.Shutdown(); err != nil {
			mainLogger.Errorf("shutdown libp2p failed: %v", err)
		}
	}()

	libp2pControl.RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *mylibp2p.Message) {
		mainLogger.Infof("received %v protocol chat message from %s content %v", protocolID, msg.From, string(msg.Content))
	})

	<-quit
	mainLogger.Infof("received shutdown signal...")

}
