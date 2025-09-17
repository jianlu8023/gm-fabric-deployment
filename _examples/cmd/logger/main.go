package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"_examples/internal/logger"
)

func main() {

	appLogger := logger.GetSDKLogger()
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for range ticker.C {

			appLogger.Debug("debug log")
			appLogger.Info("info log")
			appLogger.Warn("warn log")
			appLogger.Error("error log")
			// appLogger.Fatal("fatal log")
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-exit:
		ticker.Stop()
		appLogger.Infof("server stopped ...")
	}
}
