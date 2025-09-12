package main

import (
	"database/sql"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/router"
	commonhttp "github.com/jianlu8023/gm-fabric-deployment/pkg/common/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model/docker/image"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model/docker/network"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/model/node"

	"github.com/jianlu8023/gm-fabric-deployment/internal/web/http"
	"github.com/jianlu8023/gm-fabric-deployment/internal/web/mapper"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/job"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/server"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/str"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/system/pidfile"
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
			if err := pidfile.CreateOrUpdatePIDFile("server.pid"); err != nil {
				mainLogger.Errorf("generate pid file failed: %v", err)
				return
			}
			defer func() {
				pidfile.ReleasePID()
			}()
		case "darwin": // macOS
			mainLogger.Infof("running on MacOS")
			if err := pidfile.CreateOrUpdatePIDFile("server.pid"); err != nil {
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

	if serverControl.GetDatasourceControl() != nil {
		serverControl.GetDatasourceControl().RegisterAutoMigrateTable(
			&image.Info{},
			&node.Info{},
			&network.Info{},
			&model.UserInfo{},
		)
	}

	if serverControl.GetHttpControl() != nil {
		serverControl.GetHttpControl().RegisterRouter(func() []commonhttp.RouterHandler {
			return []commonhttp.RouterHandler{}
		}())

		serverControl.GetHttpControl().RegisterRouter(router.NewRouter(
			serverControl.GetLoggerControl(),
			serverControl.GetLibp2pControl(),
			serverControl.GetGrpcControl(),
			serverControl.GetDockerControl(),
			serverControl.GetDatasourceControl(),
			serverControl.GetWebsocketControl(),
			serverControl.GetHttpControl(),
		))
	}

	serverControl.StartUp(func(err error) {
		if err != nil && !http.IsHttpErrServerClosed(err) {
			mainLogger.Errorf("http server start err: %v", err)
			quit <- os.Interrupt
		} else {
			mainLogger.Infof("http server closed normally") // 可选：记录正常关闭日志
		}
	})
	defer func(serverControl *server.Control) {
		if err := serverControl.Shutdown(); err != nil {
			mainLogger.Errorf("shutdown server err: %v", err)
		}
	}(serverControl)

	// datasource
	var (
		imageMapper   *image.Mapper
		nodeMapper    *mapper.NodeMapper
		networkMapper *network.Mapper
	)
	{
		imageMapper = image.NewImageMapper(serverControl.GetDatasourceControl().GetConn())
		nodeMapper = mapper.NewNodeMapper(mapper.NewMapper(
			serverControl.GetLoggerControl().GenLogger(logger.ModuleDataSource),
			serverControl.GetDatasourceControl().GetConn()),
		)
		networkMapper = network.NewNetworkMapper(serverControl.GetDatasourceControl().GetConn())
	}

	// libp2p
	{
		myself := node.NewNodeInfo()
		myself.NodeId = serverControl.GetLibp2pControl().GetLocalhostPeerID().String()
		myself.IsAlive = sql.NullBool{Bool: true, Valid: true}
		myself.IsMySelf = sql.NullBool{Bool: true, Valid: true}
		myself.LastAliveMessageTime = time.Now()
		if err = nodeMapper.InsertOrUpdate(myself); err != nil {
			mainLogger.Errorf("insert myself info failed: %v", err)
		}

		serverControl.GetLibp2pControl().RegisterMessageHandler("chat_message", func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("received %v protocol chat message from %s content %v", protocolID, msg.From, string(msg.Content))
		})
		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.Libp2pNode, func(protocolID protocol.ID, msg *libp2p.Message) {
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

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.BaseShutdown, func(protocolId protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("[control] received %v protocol shutdown message from %v", protocolId, msg.From)
			serverControl.GetLibp2pControl().DisconnectFromPeer(msg.From)
			mainLogger.Infof("from connect peer list remove peer %v", msg.From)
			info := node.NewNodeInfo()
			info.NodeId = msg.From.String()
			info.IsAlive = sql.NullBool{Bool: false, Valid: true}
			info.LastAliveMessageTime = time.Now()
			if err := nodeMapper.InsertOrUpdate(info); err != nil {
				mainLogger.Errorf("update node info failed: %v", err)
			}
		})

		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.DockerNetworks, func(protocolID protocol.ID, msg *libp2p.Message) {
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
		serverControl.GetLibp2pControl().RegisterMessageHandler(libp2p.DockerImages, func(protocolID protocol.ID, msg *libp2p.Message) {
			mainLogger.Debugf("receive %v protocol %v message from %v", protocolID, msg.Type, msg.From)

			var imageList []types.ImageSummary
			if err := json.Unmarshal(msg.Content, &imageList); err != nil {
				mainLogger.Errorf("unmarshal docker images failed: %v", err)
				return
			}

			for _, img := range imageList {
				info := image.NewImageInfo()
				info.ImageName = img.RepoTags[0]
				info.ImageId = img.ID
				info.ImageCreated = img.Created
				labels, err := json.Marshal(img.Labels)
				if err != nil {
					mainLogger.Errorf("marshal image labels failed: %v", err)
					continue
				}
				info.ImageLabels = string(labels)
				info.IsDelete = sql.NullBool{Bool: false, Valid: true}
				info.ImageLocationPeerId = msg.From.String()
				if err := imageMapper.InsertOrUpdateOne(info); err != nil {
					mainLogger.Errorf("insert or update image info failed: %v", err)
					continue
				}
			}

		})

	}

	{
		serverControl.GetJobControl().RegisterJob(&job.Job{
			Name:     "collect-docker-network",
			Interval: time.Second * 10,
			Task: func() {
				// 收集docker网络信息
				dockerNetworkMsg := &libp2p.Message{
					Content: []byte("collect all node docker network info..."),
					Type:    libp2p.CollectionDockerNetworks,
				}
				if err := serverControl.GetLibp2pControl().BroadcastMessage(dockerNetworkMsg); err != nil {
					mainLogger.Errorf("broadcast collect docker network info message failed: %v", err)
				}
			},
		})
		serverControl.GetJobControl().RegisterJob(&job.Job{
			Name:     "collect-docker-images",
			Interval: time.Second * 15,
			Task: func() {
				// 收集docker镜像信息
				collectDockerImageMsg := &libp2p.Message{
					Type:    libp2p.CollectionDockerImages,
					Content: []byte("collect all node docker image info..."),
				}
				if err := serverControl.GetLibp2pControl().BroadcastMessage(collectDockerImageMsg); err != nil {
					mainLogger.Errorf("broadcast collect docker image info message failed: %v", err)
				}
			},
		})
		serverControl.GetJobControl().RegisterJob(&job.Job{
			Name: "chat-message",
			Task: func() {
				chatMsg := &libp2p.Message{
					Type:    "chat_message",
					Content: []byte("hello libp2p"),
				}
				if err := serverControl.GetLibp2pControl().BroadcastMessage(chatMsg); err != nil {
					mainLogger.Errorf("broadcast chat message failed: %v", err)
				}

			},
			Interval: time.Second * 5,
		})
		serverControl.GetJobControl().RegisterJob(&job.Job{
			Name: "collect-node-info",
			Task: func() {
				collectInfoMsg := &libp2p.Message{
					Type:    libp2p.CollectionNode,
					Content: []byte("collect all node info"),
				}
				// 广播消息
				if err := serverControl.GetLibp2pControl().BroadcastMessage(collectInfoMsg); err != nil {
					mainLogger.Errorf("broadcast collect info message failed: %v", err)
				}
			},
			Interval: time.Second * 7,
		})
	}

	if !str.CompareIgnoreCase("windows", runtime.GOOS) {
		imageList, err := serverControl.GetDockerControl().ListImages()
		if err != nil {
			mainLogger.Errorf("list docker images failed: %v", err)
		} else {
			for _, img := range imageList {
				info := image.NewImageInfo()
				info.ImageName = img.RepoTags[0]
				info.IsDelete = sql.NullBool{Bool: false, Valid: true}
				info.ImageId = img.ID
				info.ImageCreated = img.Created
				labels, err := json.Marshal(img.Labels)
				if err != nil {
					mainLogger.Errorf("marshal image labels failed: %v", err)
					continue
				}
				info.ImageLabels = string(labels)
				info.ImageLocationPeerId = serverControl.GetLibp2pControl().GetLocalhostPeerID().String()
				if err := imageMapper.InsertOrUpdateOne(info); err != nil {
					mainLogger.Errorf("insert or update image info failed: %v", err)
					continue
				}
			}
		}
		networkList, err := serverControl.GetDockerControl().ListNetworks()
		if err != nil {
			mainLogger.Errorf("list docker networks failed: %v", err)
		} else {
			for _, net := range networkList {
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
				info.NetworkLocationPeerId = serverControl.GetLibp2pControl().GetLocalhostPeerID().String()
				info.IsDelete = sql.NullBool{Bool: false, Valid: true}
				if err := networkMapper.InsertOrUpdateOne(info); err != nil {
					mainLogger.Errorf("insert or update network info failed: %v", err)
				}
			}
		}

		// if err = dockerControl.PullImage("busybox:latest"); err != nil {
		// 	mainLogger.Errorf("pull image failed: %v", err)
		// }
	}

	mainLogger.Infof("starting http server agagin...")
	serverControl.GetHttpControl().StartUp(func(err error) {
		if !http.IsHttpErrServerClosed(err) {
			mainLogger.Errorf("start http server err: %v", err)
		}
		quit <- os.Interrupt
	})

	<-quit
	mainLogger.Infof("received shutdown signal...")
	time.Sleep(5 * time.Second)
}
