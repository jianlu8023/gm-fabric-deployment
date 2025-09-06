package libp2p

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jianlu8023/gm-fabric-deployment/internal/logger"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/config"
	"github.com/libp2p/go-libp2p/core/peer"
)

// RunLibp2pExample 运行libp2p示例
func RunLibp2pExample(listenAddr string) error {
	// 创建配置控制器
	configControl, err := config.NewConfigControl()
	if err != nil {
		return fmt.Errorf("load config failed: %v", err)
	}

	// 创建logger控制器
	loggerControl := logger.NewLoggerControl(configControl.GetConfig().LoggerConfig)

	// 创建libp2p配置
	libp2pConfig := &config.Libp2pConfig{
		ListenAddr:    listenAddr,              // 监听所有接口的2000端口
		ProtocolID:    "/gm-fabric/chat/1.0.0", // 自定义协议ID
		ServiceTag:    "gm-fabric-deployment",  // mdns serviceTag
		BootstrapList: []string{
			// "/ip4/127.0.0.1/tcp/2000",
		},
	}

	// 创建libp2p控制器
	libp2pControl, err := NewLibp2pControl(libp2pConfig, loggerControl)
	if err != nil {
		return fmt.Errorf("create libp2p control failed: %v", err)
	}

	// 创建上下文，用于捕获中断信号
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册消息处理器
	libp2pControl.RegisterMessageHandler("chat_message", handleChatMessage)
	libp2pControl.RegisterMessageHandler("system_message", handleSystemMessage)

	// 设置节点发现回调
	libp2pControl.discoveryService.NotifyPeerFound(func(id peer.ID, info peer.AddrInfo) {
		fmt.Printf("New peer discovered and connected: %s\n", id)
	})

	// 启动libp2p服务
	if err := libp2pControl.Start(); err != nil {
		libp2pControl.Shutdown()
		return fmt.Errorf("start libp2p service failed: %v", err)
	}
	defer libp2pControl.Shutdown()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Libp2p node running. Press Ctrl+C to stop.")

	// 模拟发送一条消息
	go func() {
		// 等待一段时间，让节点有机会发现其他节点
		time.Sleep(5 * time.Second)
		ticker := time.NewTicker(time.Second * 5)

		for range ticker.C {
			// 创建一条聊天消息
			msg := &Message{
				Type:    "chat_message",
				Content: []byte("Hello from libp2p example!"),
			}

			// 广播消息
			if err := libp2pControl.BroadcastMessage(msg); err != nil {
				fmt.Printf("Failed to broadcast message: %v\n", err)
			} else {
				fmt.Println("Broadcasted chat message to network")
			}
		}
	}()

	// 等待中断信号
	select {
	case <-sigCh:
		fmt.Println("Received interrupt signal, shutting down...")
	case <-ctx.Done():
		fmt.Println("Context canceled, shutting down...")
	}

	return nil
}

// handleChatMessage 处理聊天消息
func handleChatMessage(msg *Message) {
	fmt.Printf("received chat message from %s: %s\n", msg.From, string(msg.Content))
}

// handleSystemMessage 处理系统消息
func handleSystemMessage(msg *Message) {
	fmt.Printf("received system message from %s: %s\n", msg.From, string(msg.Content))
}

// ExampleUsageInApplication 完整的使用示例 - 如何在实际应用中集成Libp2pControl
func ExampleUsageInApplication() {
	// 1. 初始化配置和logger
	configControl, _ := config.NewConfigControl()
	loggerControl := logger.NewLoggerControl(configControl.GetConfig().LoggerConfig)

	// 2. 创建libp2p配置
	libp2pConfig := &config.Libp2pConfig{
		ListenAddr: "/ip4/0.0.0.0/tcp/2000",
		ProtocolID: "/my-application/protocol/1.0.0",
	}

	// 3. 创建libp2p控制器
	libp2pControl, _ := NewLibp2pControl(libp2pConfig, loggerControl)

	// 4. 注册消息处理器
	libp2pControl.RegisterMessageHandler("data_request", func(msg *Message) {
		// 处理数据请求消息
	})

	libp2pControl.RegisterMessageHandler("data_response", func(msg *Message) {
		// 处理数据响应消息
	})

	// 5. 启动libp2p服务
	libp2pControl.Start()

	// 6. 在需要时发送消息
	// 广播消息
	broadcastMsg := &Message{
		Type:    "system_announcement",
		Content: []byte("System update notification"),
	}
	libp2pControl.BroadcastMessage(broadcastMsg)

	// 7. 应用退出时关闭libp2p服务
	defer libp2pControl.Shutdown()
}
