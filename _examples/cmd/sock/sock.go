package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

const sock = "app.sock"

func main() {
	if err := os.Remove(sock); err != nil && !os.IsNotExist(err) {
		log.Printf("error removing socket file: %v", err)
		return
	}
	listen, err := net.Listen("unix", sock)
	if err != nil {
		log.Printf("error listening on socket: %v", err)
		return
	}
	defer listen.Close()
	log.Printf("listening on socket: %s", sock)
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 读取数据
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error:", err)
			return
		}

		// 打印接收到的数据
		fmt.Printf("Received: %s", buf[:n])

		// 发送回数据
		_, err = conn.Write([]byte("hello from server"))
		if err != nil {
			fmt.Println("write error:", err)
			return
		}
	}
}

func client() {
	// 连接到服务端
	conn, err := net.Dial("unix", sock)
	if err != nil {
		fmt.Println("dial error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 发送数据
	message := "Hello from client!"
	fmt.Println("Sending:", message)
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("write error:", err)
		return
	}

	// 读取服务端返回的数据
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("read error:", err)
		return
	}

	// 打印接收到的数据
	fmt.Printf("Received from server: %s\n", buf[:n])
}
