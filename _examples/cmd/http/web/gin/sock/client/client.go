package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	socketFile = "/tmp/gin.sock" // Socket 文件路径 (与服务器端一致)
)

func main() {
	// 创建 Unix Domain Socket 连接
	conn, err := net.Dial("unix", socketFile)
	if err != nil {
		log.Fatalf("Error dialing: %v", err)
	}
	defer conn.Close()

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", "http://unix/ping", nil) // URL 中的 host 可以是任意值，这里使用 "unix"
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// 发送 HTTP 请求
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketFile)
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading body: %v", err)
	}

	fmt.Printf("Response: %s\n", string(body))

	// 删除 Socket 文件 (可选，如果客户端也创建了)
	if err := os.Remove(socketFile); err != nil && !os.IsNotExist(err) { // 忽略文件不存在的错误
		log.Printf("Warning: Error removing socket file: %v", err)
	}
}
