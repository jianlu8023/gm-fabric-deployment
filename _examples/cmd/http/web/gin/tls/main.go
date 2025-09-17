package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"_examples/internal/logger"
	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	router.Any("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
			"code":    http.StatusOK,
			"success": true,
		})
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServeTLS("./cert/rsa.crt", "./cert/rsa.key"); err != nil {
			logger.GetAppLogger().Errorf("启动gin web服务失败 %v", err)
			quit <- syscall.SIGINT
		}
	}()

	<-quit
	if err := srv.Shutdown(context.Background()); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		logger.GetAppLogger().Errorf("关闭gin web服务失败 %v", err)
	}

}
