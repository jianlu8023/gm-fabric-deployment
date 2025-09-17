package main

import (
	"_examples/internal/logger"
	mywebsocket "_examples/pkg/websocket"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// https://www.squash.io/implementing-real-time-features-websockets-in-gin-in-golang/

	router := gin.Default()

	router.GET("/ws", func(ctx *gin.Context) {

		conn, err := mywebsocket.Upgrade(ctx.Writer, ctx.Request)
		if err != nil {
			fmt.Printf("websocket连接失败, %v\n", err)
			ctx.SecureJSON(http.StatusInternalServerError, gin.H{
				"msg":     "websocket连接失败",
				"code":    http.StatusInternalServerError,
				"success": false,
			})
			return
		}

		mywebsocket.SetClient(mywebsocket.NewClientManager(ctx.RemoteIP(), conn))

		go func(addr string) {
			defer func() {
				mywebsocket.CloseClient(addr)
				fmt.Printf("websocket %s 连接关闭\n", addr)
			}()
			// 处理WebSocket消息
			fmt.Printf("websocket %s 连接成功\n", addr)
			ws := mywebsocket.GetClient(addr)
			for {
				messageType, p, err := ws.Conn.ReadMessage()
				if err != nil {
					fmt.Printf("read msg error %v\n", err)
					return
				}
				fmt.Printf("msg type %d\n", messageType)
				switch messageType {
				case websocket.TextMessage:
					fmt.Printf("处理文本消息, %s\n", string(p))
					ws.SendMsg(websocket.TextMessage, p)
				case websocket.BinaryMessage:
					fmt.Println("处理二进制消息")
				case websocket.CloseMessage:
					fmt.Println("关闭websocket连接")
					return
				case websocket.PingMessage:
					fmt.Println("处理ping消息")
					ws.SendMsg(websocket.PingMessage, nil)
				case websocket.PongMessage:
					fmt.Println("处理pong消息")
					ws.SendMsg(websocket.PongMessage, nil)
				default:
					fmt.Printf("未知消息类型: %d\n", messageType)
					return
				}
			}
		}(ctx.RemoteIP())
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.GetAppLogger().Errorf("starter gin web error %v", err)
			quit <- syscall.SIGINT
		}
	}()

	<-quit
	if err := srv.Shutdown(context.Background()); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.GetAppLogger().Info("gin web服务 正常关闭")
		} else {
			logger.GetAppLogger().Errorf("gin web服务 关闭失败 %v", err)
		}
	}
}
