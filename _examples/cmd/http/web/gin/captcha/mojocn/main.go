package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"

	"_examples/internal/logger"
)

var (
	store = base64Captcha.DefaultMemStore
)

func main() {

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/captcha", func(ctx *gin.Context) {
		driverDigit := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
		captcha := base64Captcha.NewCaptcha(driverDigit, store)
		id, b64s, answer, err := captcha.Generate()
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusInternalServerError,
				"message": err.Error(),
				"success": false,
			})
			return
		}
		body := map[string]interface{}{
			"code":    http.StatusOK,
			"message": "success",
			"success": true,
			"id":      id,
			"base64":  template.URL(b64s),
			"answer":  answer,
		}
		fmt.Println(body)
		ctx.HTML(http.StatusOK, "index.html", body)
		// ctx.JSON(http.StatusOK, gin.H{
		// 	"code":    http.StatusOK,
		// 	"message": "success",
		// 	"success": true,
		// 	"id":      id,
		// 	"base64":  b64s,
		// 	"answer":  answer,
		// })
	})
	router.POST("/captcha/verify", func(ctx *gin.Context) {

		id := ctx.PostForm("id")
		answer := ctx.PostForm("answer")
		if len(id) == 0 || len(answer) == 0 {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusBadRequest,
				"message": "id or answer is empty",
				"success": false,
			})
		}
		if store.Verify(id, answer, true) {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"message": "success",
				"success": true,
			})
		} else {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    http.StatusBadRequest,
				"message": "failed",
				"success": false,
			})
		}

	})

	srv := &http.Server{
		Handler: router,
		Addr:    ":8080",
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.GetAppLogger().Errorf("启动 gin web服务 失败 %v", err)
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
