package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	mylogger "github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/internal/middleware/grpc"
	myhttp "github.com/jianlu8023/gm-fabric-deployment/internal/middleware/http"
	mylibp2p "github.com/jianlu8023/gm-fabric-deployment/internal/middleware/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/pidfile"
	"github.com/libp2p/go-libp2p/core/protocol"
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	control, err := grpc.NewGrpcControl(configControl.GetGrpcConfig(), loggerControl)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer control.Shutdown()

	mainLogger.Infof("starting http server...")
	httpControl := myhttp.NewServerControl(configControl.GetWebConfig(), loggerControl)

	mainLogger.Infof("starting libp2p server...")
	libp2pControl, err := mylibp2p.NewLibp2pControl(configControl.GetLibp2pConfig(), loggerControl)
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

	// libp2pControl.RegisterNotifyPeerFound(func(id peer.ID, info peer.AddrInfo) {
	// 	fmt.Printf("New peer discovered and connected: %s\n", id)
	// })

	libp2pControl.RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *mylibp2p.Message) {
		mainLogger.Infof("received %v protocol chat message from %s content %v", protocolID, msg.From, string(msg.Content))
	})

	// 模拟发送一条消息
	go func() {
		// 等待一段时间，让节点有机会发现其他节点
		time.Sleep(5 * time.Second)
		ticker := time.NewTicker(time.Second * 5)

		for range ticker.C {
			// 创建一条聊天消息
			msg := &mylibp2p.Message{
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

	httpControl.StartUp(func(err error) {
		if !errors.Is(err, http.ErrServerClosed) {
			mainLogger.Errorf("start http server err: %v", err)
		}
		quit <- os.Interrupt
	})
	defer httpControl.Shutdown()

	<-quit
	mainLogger.Infof("received shutdown signal...")
}
