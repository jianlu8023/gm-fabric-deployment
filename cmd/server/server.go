package main

import (
	"database/sql"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/internal/datasource/docker/image"
	"github.com/jianlu8023/gm-fabric-deployment/internal/datasource/node"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
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
		imageMapper *image.Mapper
		nodeMapper  *node.Mapper
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
		if err = dataSourceControl.AutoMigrateTable(&image.Info{}, &node.Info{}); err != nil {
			mainLogger.Errorf("auto migrate table failed: %v", err)
		}
		imageMapper = image.NewImageMapper(dataSourceControl.GetConn())
		nodeMapper = node.NewNodeMapper(dataSourceControl.GetConn())
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

		// libp2pControl.RegisterNotifyPeerFound(func(id peer.ID, info peer.AddrInfo) {
		// 	fmt.Printf("New peer discovered and connected: %s\n", id)
		// })

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
		libp2pControl.RegisterMessageHandler("node_info", func(protocolID protocol.ID, msg *libp2p.Message) {
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

		libp2pControl.RegisterMessageHandler("base/shutdown", func(protocolId protocol.ID, msg *libp2p.Message) {
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

		// 模拟发送一条消息
		go func() {
			// 等待一段时间，让节点有机会发现其他节点
			time.Sleep(5 * time.Second)
			ticker := time.NewTicker(time.Second * 5)

			for range ticker.C {
				// 创建一条聊天消息
				collectInfoMsg := &libp2p.Message{
					From:    libp2pControl.GetLocalID(),
					Type:    "collect_info",
					Content: []byte("collect all node info"),
				}

				// 广播消息
				if err := libp2pControl.BroadcastMessage(collectInfoMsg); err != nil {
					mainLogger.Errorf("broadcast collect info message failed: %v", err)
				}
				chatMsg := &libp2p.Message{
					From:    libp2pControl.GetLocalID(),
					Type:    "chat_message",
					Content: []byte("hello libp2p"),
				}
				if err := libp2pControl.BroadcastMessage(chatMsg); err != nil {
					mainLogger.Errorf("broadcast chat message failed: %v", err)
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
			imageList, err := dockerControl.ImageList()
			if err != nil {
				mainLogger.Errorf("get docker image list failed: %v", err)

			} else {
				mainLogger.Infof("get docker img success...")
				for _, img := range imageList {
					info := image.NewImageInfo()
					info.ImageName = img.RepoTags[0]
					info.ImageLocationPeerId = configControl.GetLibp2pConfig().Identity.PeerID
					info.IsDelete = sql.NullBool{Bool: false, Valid: true}
					if err := imageMapper.InsertOrUpdateOne(info); err != nil {
						mainLogger.Errorf("insert image info failed: %v", err)
					}
				}
			}
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
}
