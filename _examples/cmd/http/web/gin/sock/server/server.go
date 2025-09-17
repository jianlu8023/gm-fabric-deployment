package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	socketFile = "/tmp/gin.sock" // Socket 文件路径
)

func main() {
	// 确保 Socket 文件不存在，如果存在则删除
	if _, err := os.Stat(socketFile); err == nil {
		if err := os.Remove(socketFile); err != nil {
			log.Fatalf("Error removing existing socket file: %v", err)
		}
	}

	// 创建 Gin 引擎
	router := gin.Default()

	// 定义一个简单的路由
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 创建 Unix Domain Socket 监听器
	listener, err := net.Listen("unix", socketFile)
	if err != nil {
		log.Fatalf("Error creating listener: %v", err)
	}
	defer listener.Close()

	// 设置 Socket 文件权限 (可选)
	if err := os.Chmod(socketFile, 0777); err != nil {
		log.Printf("Warning: Error setting socket file permissions: %v", err) // 权限设置失败不会导致程序退出
	}

	// 启动 Gin 服务
	fmt.Printf("Gin server listening on unix socket: %s\n", socketFile)
	if err := http.Serve(listener, router); err != nil {
		log.Fatalf("Error serving: %v", err)
	}
}
