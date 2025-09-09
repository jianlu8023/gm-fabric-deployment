package main

import (
	"context"
	"fmt"
	mylogger "github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	configControl, err := config.NewConfigControl()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		return
	}

	loggerControl := mylogger.NewLoggerControl(configControl.GetConfig().LoggerConfig)

	mainLogger := loggerControl.GenLogger("main")

	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/10.10.10.40/tcp/2000"))
	libp2p.Ping(false)
	if err != nil {
		mainLogger.Errorf("create node failed: %v", err)
		return
	}
	defer func() {
		if err := node.Close(); err != nil {
			mainLogger.Errorf("close node failed: %v", err)
		}
	}()
	mainLogger.Infof("create node successfully")
	mainLogger.Infof("node info address %v", node.Addrs())
	mainLogger.Infof("node info id %v", node.ID())

	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	mainLogger.Infof("libp2p node address: %v", addrs)

	if len(os.Args) > 1 {
		addr, err := multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			panic(err)
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			panic(err)
		}
		if err := node.Connect(context.Background(), *peer); err != nil {
			panic(err)
		}
		fmt.Println("sending 5 ping messages to", addr)
		ch := pingService.Ping(context.Background(), peer.ID)
		for i := 0; i < 5; i++ {
			res := <-ch
			fmt.Println("pinged", addr, "in", res.RTT)
		}
	} else {
		// wait for a SIGINT or SIGTERM signal
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		fmt.Println("Received signal, shutting down...")
	}

	// <-quit
	// mainLogger.Infof("received signal, shutting down...")
}
