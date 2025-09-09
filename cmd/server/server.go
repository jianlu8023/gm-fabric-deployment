package main

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
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
	fmt.Printf("start server version %s\n", version)

	runOS := runtime.GOOS
	switch runOS {
	case "windows":
		fmt.Println("Running on Windows")
	case "linux":
		fmt.Println("Running on Linux")
		if err := pidfile.CreateOrUpdatePIDFile("server.pid"); err != nil {
			fmt.Printf("generate pid file failed: %v\n", err)
			return
		}
		defer func() {
			pidfile.ReleasePID()
		}()
	case "darwin": // macOS
		fmt.Println("Running on macOS")
		if err := pidfile.CreateOrUpdatePIDFile("server.pid"); err != nil {
			fmt.Printf("generate pid file failed: %v\n", err)
			return
		}
		defer func() {
			pidfile.ReleasePID()
		}()
	default:
		fmt.Printf("Running on an unknown operating system: %s\n", runOS)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	mainLogger := loggerControl.GenLogger("main")

	{
		mainLogger.Infof("starting grpc server...")
		grpcControl, err := grpc.NewGrpcControl(configControl.GetGrpcConfig(), loggerControl, func(err error) {
			mainLogger.Errorf("grpc server startUp failed: %v", err)
			quit <- os.Interrupt
		})
		if err != nil {
			mainLogger.Errorf("create grpc grpcControl failed: %v", err)
			return
		}
		defer grpcControl.Shutdown()
	}

	{
		mainLogger.Infof("starting libp2p server...")
		libp2pControl, err := libp2p.NewLibp2pControl(configControl.GetLibp2pConfig(), loggerControl)
		if err != nil {
			mainLogger.Errorf("create libp2p control failed: %v", err)
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

		// libp2pControl.RegisterNotifyPeerFound(func(id peer.ID, info peer.AddrInfo) {
		// 	fmt.Printf("New peer discovered and connected: %s\n", id)
		// })

		libp2pControl.RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Infof("received %v protocol chat message from %s content %v", protocolID, msg.From, string(msg.Content))
		})

		// 模拟发送一条消息
		go func() {
			// 等待一段时间，让节点有机会发现其他节点
			time.Sleep(5 * time.Second)
			ticker := time.NewTicker(time.Second * 5)

			for range ticker.C {
				// 创建一条聊天消息
				msg := &libp2p.Message{
					From:    libp2pControl.GetLocalID(),
					Type:    "chat_message",
					Content: []byte("Hello from libp2p example!"),
				}

				// 广播消息
				if err := libp2pControl.BroadcastMessage(msg); err != nil {
					mainLogger.Errorf("broadcast message failed: %v", err)
				} else {
					mainLogger.Debugf("broadcast message successfully")
				}
			}
		}()
	}

	{
		mainLogger.Infof("starting datasource server...")
		dataSourceControl, err := datasource.NewDataSourceControl(configControl.GetDataSourceConfig(), loggerControl)
		if err != nil {
			mainLogger.Fatalf("create datasource control failed: %v", err)
			return
		}
		defer dataSourceControl.Close()

	}

	{
		mainLogger.Infof("starting http server...")
		httpControl := http.NewServerControl(configControl.GetWebConfig(), loggerControl)

		httpControl.StartUp(func(err error) {
			if !http.IsHttpErrServerClosed(err) {
				mainLogger.Errorf("start http server err: %v", err)
			}
			quit <- os.Interrupt
		})
		defer httpControl.Shutdown()
	}

	<-quit
	mainLogger.Infof("received shutdown signal...")
}
