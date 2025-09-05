package main

import (
	"fmt"

	"github.com/jianlu8023/gm-fabric-deployment/internal/middleware/libp2p"
)

func main() {
	fmt.Println("Starting libp2p control example...")

	// 运行libp2p示例
	if err := libp2p.RunLibp2pExample("/ip4/127.0.0.1/tcp/2000"); err != nil {
		fmt.Printf("Error running libp2p example: %v\n", err)
	}

	fmt.Println("Libp2p control example finished.")
}
