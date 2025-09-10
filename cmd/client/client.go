package main

import (
	"context"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc/pb"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/system/pidfile"
	"github.com/libp2p/go-libp2p/core/protocol"
	"math/rand/v2"
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelFunc := context.WithCancel(context.Background())

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
			mainLogger.Errorf("start grpc server failed: %v", err)
			quit <- os.Interrupt
		})
		if err != nil {
			mainLogger.Errorf("start grpc server failed: %v", err)
			return
		}
		defer grpcControl.Shutdown()

		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(rand.IntN(5-3)+3))
			for range ticker.C {
				select {
				default:
					response, err := grpcControl.Call(&pb.BaseRequest{
						MessageType: libp2p.BasePing,
					})
					if err != nil {
						mainLogger.Errorf("send ping message err: %v", err)
					} else {
						mainLogger.Debugf("send ping message success %v", response)
					}
				case <-ctx.Done():
					mainLogger.Infof("received interrupt signal, exiting...")
					return
				}
			}
		}()
	}

	var dockerControl *docker.Control
	{
		switch runtime.GOOS {
		case "windows":
			mainLogger.Infof("windows os not starting docker server...")
		case "linux":
			fallthrough
		case "darwin":
			fallthrough
		default:
			mainLogger.Infof("starting docker server...")
			dockerControl, err = docker.NewDockerControl(configControl.GetDockerConfig(), loggerControl)
			if err != nil {
				mainLogger.Fatalf("create docker control failed: %v", err)
			}
			dockerControl.StartUp(func(err error) {
				mainLogger.Errorf("check docker daemon failed: %v", err)
				quit <- os.Interrupt
			})
			defer func() {
				if err := dockerControl.Shutdown(); err != nil {
					mainLogger.Errorf("shutdown docker control failed: %v", err)
				}
			}()

		}
	}

	{
		mainLogger.Infof("starting libp2p server...")
		libp2pControl, err := libp2p.NewLibp2pControl(configControl.GetLibp2pConfig(), loggerControl)
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
			mainLogger.Debugf("received %v protocol %v messageType from %s content %v", protocolID, msg.Type, msg.From, string(msg.Content))
		})

		libp2pControl.RegisterMessageHandler(libp2p.CollectionNode, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Infof("received %v protocol %v message from %s content %v", protocolID, msg.Type, msg.From, string(msg.Content))
			nodeInfoMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Content: []byte("success"),
				Type:    libp2p.Libp2pNode,
			}
			if err := libp2pControl.SendMessageToPeer(nodeInfoMsg.To, nodeInfoMsg); err != nil {
				mainLogger.Errorf("send node info to peer %v failed: %v", nodeInfoMsg.To, err)
			}
		})

		libp2pControl.RegisterMessageHandler(libp2p.CollectionDockerNetworks, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %v message from %v", protocolID, msg.Type, msg.From)
			if dockerControl == nil {
				return
			}
			listNetworks, err := dockerControl.ListNetworks()
			if err != nil {
				mainLogger.Errorf("list docker networks failed: %v", err)
				return
			}
			bytes, err := json.Marshal(listNetworks)
			if err != nil {
				mainLogger.Errorf("marshal docker networks failed: %v", err)
				return
			}

			dockerNetworkMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Content: bytes,
				Type:    libp2p.DockerNetworks,
			}
			if err = libp2pControl.SendMessageToPeer(dockerNetworkMsg.To, dockerNetworkMsg); err != nil {
				mainLogger.Errorf("send docker networks to peer %v failed: %v", dockerNetworkMsg.To, err)
			}
		})

		libp2pControl.RegisterMessageHandler(libp2p.CollectionDockerImages, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %s message from %v", protocolID, msg.Type, msg.From)

			if dockerControl == nil {
				return
			}

			listImages, err := dockerControl.ListImages()
			if err != nil {
				mainLogger.Errorf("list docker images failed: %v", err)
				return
			}
			content, err := json.Marshal(listImages)
			if err != nil {
				mainLogger.Errorf("marshal docker images failed: %v", err)
				return
			}

			dockerImageMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Type:    libp2p.DockerImages,
				Content: content,
			}
			if err = libp2pControl.SendMessageToPeer(dockerImageMsg.To, dockerImageMsg); err != nil {
				mainLogger.Errorf("send docker images to peer %v failed: %v", dockerImageMsg.To, err)
			}

		})

		go func() {
			ticker := time.NewTicker(time.Duration(rand.IntN(10-5)+5) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				select {
				case <-ctx.Done():
					mainLogger.Infof("received ctx done signal, exit")
					return
				default:
					pingMsg := &libp2p.Message{
						Type:    libp2p.BasePing,
						Content: []byte("ping"),
						From:    libp2pControl.GetLocalID(),
					}
					if err := libp2pControl.BroadcastMessage(pingMsg); err != nil {
						mainLogger.Errorf("broadcast ping message failed: %v", err)
					}
				}

			}
		}()
	}

	<-quit
	mainLogger.Infof("received shutdown signal...")
	cancelFunc()
	time.Sleep(3 * time.Second)
}
