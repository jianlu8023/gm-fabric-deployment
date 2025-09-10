package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/jianlu8023/gm-fabric-deployment/internal/datasource/docker/image"
	"github.com/jianlu8023/gm-fabric-deployment/internal/datasource/docker/network"
	"github.com/jianlu8023/gm-fabric-deployment/internal/datasource/node"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
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
	// pidfile
	{
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
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	// config
	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	// logger
	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	mainLogger := loggerControl.GenLogger("main")

	// datasource
	var (
		imageMapper   *image.Mapper
		nodeMapper    *node.Mapper
		networkMapper *network.Mapper
	)
	{
		mainLogger.Infof("starting datasource server...")
		dataSourceControl, err := datasource.NewDataSourceControl(configControl.GetDataSourceConfig(), loggerControl)
		if err != nil {
			mainLogger.Fatalf("create datasource control failed: %v", err)
			return
		}
		defer func() {
			if err := dataSourceControl.Close(); err != nil {
				mainLogger.Errorf("close datasource control failed: %v", err)
			}
		}()
		if err = dataSourceControl.AutoMigrateTable(&image.Info{}, &node.Info{}, &network.Info{}); err != nil {
			mainLogger.Errorf("auto migrate table failed: %v", err)
		}
		imageMapper = image.NewImageMapper(dataSourceControl.GetConn())
		nodeMapper = node.NewNodeMapper(dataSourceControl.GetConn())
		networkMapper = network.NewNetworkMapper(dataSourceControl.GetConn())
	}

	// grpc
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

	// libp2p
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

		myself := node.NewNodeInfo()
		myself.NodeId = libp2pControl.GetLocalID().String()
		myself.IsAlive = sql.NullBool{Bool: true, Valid: true}
		myself.IsMySelf = sql.NullBool{Bool: true, Valid: true}
		myself.LastAliveMessageTime = time.Now()
		if err = nodeMapper.InsertOrUpdate(myself); err != nil {
			mainLogger.Errorf("insert myself info failed: %v", err)
		}

		libp2pControl.RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol chat message from %s content %v", protocolID, msg.From, string(msg.Content))
		})
		libp2pControl.RegisterMessageHandler(libp2p.Libp2pNode, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol node info message from %s content %v", protocolID, msg.From, string(msg.Content))
			mainLogger.Infof("starting insert or update node info...")
			// 返回节点信息
			info := node.NewNodeInfo()
			info.NodeId = msg.From.String()
			info.IsAlive = sql.NullBool{Bool: true, Valid: true}
			info.LastAliveMessageTime = time.Now()
			if err := nodeMapper.InsertOrUpdate(info); err != nil {
				mainLogger.Errorf("insert or update node info failed: %v", err)
			}
		})

		libp2pControl.RegisterMessageHandler(libp2p.BaseShutdown, func(protocolId protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("[control] received %v protocol shutdown message from %v", protocolId, msg.From)
			libp2pControl.DisconnectFromPeer(msg.From)
			mainLogger.Infof("from connect peer list remove peer %v", msg.From)
			info := node.NewNodeInfo()
			info.NodeId = msg.From.String()
			info.IsAlive = sql.NullBool{Bool: false, Valid: true}
			info.LastAliveMessageTime = time.Now()
			if err := nodeMapper.InsertOrUpdate(info); err != nil {
				mainLogger.Errorf("update node info failed: %v", err)
			}
		})

		libp2pControl.RegisterMessageHandler(libp2p.DockerNetworks, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol message from %v", protocolID, msg.From)
			// 处理消息
			var networks []types.NetworkResource

			if err := json.Unmarshal(msg.Content, &networks); err != nil {
				mainLogger.Errorf("unmarshal docker networks failed: %v", err)
				return
			}

			for _, net := range networks {
				info := network.NewNetworkInfo()
				info.NetworkName = net.Name
				info.NetworkID = net.ID
				info.NetworkCreateTime = net.Created
				info.NetworkScope = net.Scope
				info.NetworkDriver = net.Driver
				info.NetworkEnableIPv6 = sql.NullBool{Bool: net.EnableIPv6, Valid: true}
				ipamBytes, err := json.Marshal(net.IPAM)
				if err != nil {
					mainLogger.Errorf("marshal network ipam failed: %v", err)
					continue
				}
				info.NetworkIpam = string(ipamBytes)
				info.NetworkInternal = sql.NullBool{Bool: net.Internal, Valid: true}
				info.NetworkAttachable = sql.NullBool{Bool: net.Attachable, Valid: true}
				info.NetworkIngress = sql.NullBool{Bool: net.Ingress, Valid: true}
				info.NetworkLocationPeerId = msg.From.String()
				info.IsDelete = sql.NullBool{Bool: false, Valid: true}
				if err := networkMapper.InsertOrUpdateOne(info); err != nil {
					mainLogger.Errorf("insert or update network info failed: %v", err)
				}
			}
		})

		// 注册处理docker镜像的消息
		libp2pControl.RegisterMessageHandler(libp2p.DockerImages, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("receive %v protocol %v message from %v", protocolID, msg.Type, msg.From)

			var imageList []types.ImageSummary
			if err := json.Unmarshal(msg.Content, &imageList); err != nil {
				mainLogger.Errorf("unmarshal docker images failed: %v", err)
				return
			}

			for _, img := range imageList {
				info := image.NewImageInfo()
				info.ImageName = img.RepoTags[0]
				info.IsDelete = sql.NullBool{Bool: false, Valid: true}
				info.ImageLocationPeerId = msg.From.String()
				if err := imageMapper.InsertOrUpdateOne(info); err != nil {
					mainLogger.Errorf("insert or update image info failed: %v", err)
					continue
				}
			}

		})

		// 模拟发送一条消息
		go func() {
			// 等待一段时间，让节点有机会发现其他节点
			time.Sleep(5 * time.Second)
			chatMsgTicker := time.NewTicker(time.Second * 5)
			defer chatMsgTicker.Stop()
			collectDockerNetworkTicker := time.NewTicker(time.Second * 10)
			defer collectDockerNetworkTicker.Stop()
			collectNodeTicker := time.NewTicker(time.Second * 7)
			defer collectNodeTicker.Stop()
			collectDockerImageTicker := time.NewTicker(time.Second * 15)
			defer collectDockerImageTicker.Stop()

			for {
				select {
				case <-ctx.Done():
					mainLogger.Infof("received cancel signal, stopping server")
					return
				case <-chatMsgTicker.C:
					chatMsg := &libp2p.Message{
						Type:    "chat_message",
						Content: []byte("hello libp2p"),
					}
					if err := libp2pControl.BroadcastMessage(chatMsg); err != nil {
						mainLogger.Errorf("broadcast chat message failed: %v", err)
					}
				case <-collectDockerNetworkTicker.C:
					// 收集docker网络信息
					dockerNetworkMsg := &libp2p.Message{
						Content: []byte("collect all node docker network info..."),
						Type:    libp2p.CollectionDockerNetworks,
					}
					if err := libp2pControl.BroadcastMessage(dockerNetworkMsg); err != nil {
						mainLogger.Errorf("broadcast collect docker network info message failed: %v", err)
					}
				case <-collectNodeTicker.C:
					collectInfoMsg := &libp2p.Message{
						Type:    libp2p.CollectionNode,
						Content: []byte("collect all node info"),
					}
					// 广播消息
					if err := libp2pControl.BroadcastMessage(collectInfoMsg); err != nil {
						mainLogger.Errorf("broadcast collect info message failed: %v", err)
					}
				case <-collectDockerImageTicker.C:
					// 收集docker镜像信息
					collectDockerImageMsg := &libp2p.Message{
						Type:    libp2p.CollectionDockerImages,
						Content: []byte("collect all node docker image info..."),
					}
					if err := libp2pControl.BroadcastMessage(collectDockerImageMsg); err != nil {
						mainLogger.Errorf("broadcast collect docker image info message failed: %v", err)
					}
				}
			}

		}()
	}

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
			dockerControl, err := docker.NewDockerControl(configControl.GetDockerConfig(), loggerControl)
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
	cancel()
}
