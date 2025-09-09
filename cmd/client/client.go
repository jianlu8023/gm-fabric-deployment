package main

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc/pb"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/system/pidfile"
	"github.com/libp2p/go-libp2p/core/protocol"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	version string
)

func main() {

	fmt.Printf("start client version %s\n", version)

	{
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
	}

	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	mainLogger := loggerControl.GenLogger("main")

	{
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
		grpcControl, err := grpc.NewGrpcControl(grpcConfig, loggerControl, func(err error) {
			mainLogger.Errorf("start grpc server failed: %v", err)
			quit <- os.Interrupt
		})
		if err != nil {
			mainLogger.Errorf("start grpc server failed: %v", err)
			return
		}
		defer grpcControl.Shutdown()

		go func() {
			ticker := time.NewTicker(time.Second * 10)
			for range ticker.C {
				response, err := grpcControl.Call(&pb.BaseRequest{
					MessageType: "base/ping",
				})
				if err != nil {
					mainLogger.Errorf("send ping pb err: %v", err)
				} else {
					mainLogger.Debugf("send ping pb success %v", response)
				}
			}
		}()
	}

	{
		mainLogger.Infof("starting libp2p server...")
		libp2pConfig := &config.Libp2pConfig{
			ListenAddr: []string{
				"/ip4/0.0.0.0/tcp/2001",
			},
			ProtocolID:    "/gm-fabric/chat/1.0.0",
			ServiceTag:    "gm-fabric-deployment",
			BootstrapList: []string{},
		}

		libp2pControl, err := libp2p.NewLibp2pControl(libp2pConfig, loggerControl)
		if err != nil {
			mainLogger.Errorf("create libp2p grpcControl failed: %v", err)
			return
		}
		libp2pControl.StartUp(func(err error) {
			mainLogger.Errorf("start libp2p failed: %v", err)
			quit <- os.Interrupt
		})
		defer func() {
			if err := libp2pControl.Shutdown(); err != nil {
				mainLogger.Errorf("shutdown libp2p failed: %v", err)
			}
		}()

		libp2pControl.RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Infof("received %v protocol chat pb from %s content %v", protocolID, msg.From, string(msg.Content))
		})
	}

	<-quit
	mainLogger.Infof("received shutdown signal...")
}
