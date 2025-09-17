package main

import (
	"fmt"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/grpc/pb"
	"github.com/jianlu8023/golang-example/pkg/control/job"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/server"
	"github.com/jianlu8023/golang-example/pkg/json"
	"github.com/jianlu8023/golang-example/pkg/system/pidfile"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var (
	version string
)

func main() {

	serverControl, err := server.NewServerControlFromFile()
	if err != nil {
		fmt.Printf("generate server control failed: %v\n", err)
		return
	}

	mainLogger := serverControl.GetLoggerControl().GenLogger("main")

	mainLogger.Infof("start server version %v", version)

	// pidfile
	{
		runOS := runtime.GOOS
		switch runOS {
		case "windows":
			mainLogger.Infof("running on Windows")
		case "linux":
			mainLogger.Infof("running on Linux")
			if err := pidfile.CreateOrUpdatePIDFile("client.pid"); err != nil {
				mainLogger.Errorf("generate pid file failed: %v", err)
				return
			}
			defer func() {
				pidfile.ReleasePID()
			}()
		case "darwin": // macOS
			mainLogger.Infof("running on MacOS")
			if err := pidfile.CreateOrUpdatePIDFile("client.pid"); err != nil {
				mainLogger.Errorf("generate pid file failed: %v", err)
				return
			}
			defer func() {
				pidfile.ReleasePID()
			}()
		default:
			mainLogger.Warnf("running on an unknown operating system: %s", runOS)
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	serverControl.StartUp(func(err error) {
		mainLogger.Errorf("start server failed: %v", err)
		quit <- os.Interrupt
	})
	defer func(serverControl *server.Control) {
		if err := serverControl.Shutdown(); err != nil {
			mainLogger.Errorf("shutdown server failed: %v", err)
		}
	}(serverControl)

	{

		if serverControl.GetJobControl() != nil {
			serverControl.GetJobControl().RegisterJob(&job.Job{
				Name: "grpc-ping-message",
				Task: func() {
					response, err := serverControl.GetGrpcControl().Call(&pb.BaseRequest{
						MessageType: libp2p.MsgBasePing,
					})
					if err != nil {
						mainLogger.Errorf("send ping message err: %v", err)
					} else {
						mainLogger.Debugf("send ping message success %v", response)
					}
				},
				Interval: time.Second * time.Duration(rand.IntN(5-3)+3),
			})

			serverControl.GetJobControl().RegisterJob(&job.Job{
				Name:     "libp2p-ping-message",
				Interval: time.Duration(rand.IntN(10-5)+5) * time.Second,
				Task: func() {
					pingMsg := &libp2p.Message{
						Type:    libp2p.MsgBasePing,
						Content: []byte("ping"),
						From:    serverControl.GetLibp2pControl().GetLocalhostPeerID(),
					}
					if err := serverControl.GetLibp2pControl().BroadcastMessage(pingMsg); err != nil {
						mainLogger.Errorf("broadcast ping message failed: %v", err)
					}
				},
			})
		}
	}

	// libp2p
	{
		serverControl.GetLibp2pControl().RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %v messageType from %s content %v", protocolID, msg.Type, msg.From, string(msg.Content))
		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgCollectionNode, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Infof("received %v protocol %v message from %s content %v", protocolID, msg.Type, msg.From, string(msg.Content))
			nodeInfoMsg := &libp2p.Message{
				From:    msg.To,
				To:      msg.From,
				Content: []byte("success"),
				Type:    libp2p.MsgLibp2pNode,
			}
			if err := serverControl.GetLibp2pControl().SendMessageToPeer(nodeInfoMsg.To, nodeInfoMsg); err != nil {
				mainLogger.Errorf("send node info to peer %v failed: %v", nodeInfoMsg.To, err)
			}
		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgCollectionDockerNetworks, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %v message from %v", protocolID, msg.Type, msg.From)
			if serverControl.GetDockerControl() == nil {
				return
			}
			listNetworks, err := serverControl.GetDockerControl().ListNetworks()
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
				Type:    libp2p.MsgDockerNetworks,
			}
			if err = serverControl.GetLibp2pControl().SendMessageToPeer(dockerNetworkMsg.To, dockerNetworkMsg); err != nil {
				mainLogger.Errorf("send docker networks to peer %v failed: %v", dockerNetworkMsg.To, err)
			}
		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgCollectionDockerImages, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %s message from %v", protocolID, msg.Type, msg.From)

			if serverControl.GetDockerControl() == nil {
				return
			}

			listImages, err := serverControl.GetDockerControl().ListImages()
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
				Type:    libp2p.MsgDockerImages,
				Content: content,
			}
			if err = serverControl.GetLibp2pControl().SendMessageToPeer(dockerImageMsg.To, dockerImageMsg); err != nil {
				mainLogger.Errorf("send docker images to peer %v failed: %v", dockerImageMsg.To, err)
			}

		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.MsgDockerImagePull, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol %s message from %v", protocolID, msg.Type, msg.From)

			if serverControl.GetDockerControl() == nil {
				return
			}

			var dockerPullContent libp2p.DockerImagePullContent
			if err := json.Unmarshal(msg.Content, &dockerPullContent); err != nil {
				mainLogger.Errorf("unmarshal docker image pull content failed: %v", err)
				return
			}

			if err := serverControl.GetDockerControl().PullImage(dockerPullContent.ImageName,
				docker.WithImagePullPlatform(dockerPullContent.Platform),
				docker.WithImagePullRegistryAuth(dockerPullContent.RegistryAuth),
			); err != nil {
				mainLogger.Errorf("pull docker image failed: %v", err)
				return
			}

			listImages, err := serverControl.GetDockerControl().ListImages()
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
				Type:    libp2p.MsgDockerImages,
				Content: content,
			}
			if err = serverControl.GetLibp2pControl().SendMessageToPeer(dockerImageMsg.To, dockerImageMsg); err != nil {
				mainLogger.Errorf("send docker images to peer %v failed: %v", dockerImageMsg.To, err)
			}

		})

	}

	<-quit
	mainLogger.Infof("received shutdown signal...")

	time.Sleep(5 * time.Second)
}
