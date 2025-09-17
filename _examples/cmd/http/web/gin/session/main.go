package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"

	"_examples/internal/logger"
)

func main() {

	router := gin.Default()
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	srv := &http.Server{
		Handler: router,
		Addr:    ":8080",
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.GetAppLogger().Errorf("start gin web error %v", err)
			quit <- syscall.SIGINT
		}
	}()

	<-quit
	if err := srv.Shutdown(context.Background()); err != nil {
		logger.GetAppLogger().Errorf("shutdown gin web error %v", err)

	}

}
