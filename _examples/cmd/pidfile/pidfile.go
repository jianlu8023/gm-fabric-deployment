package main

import (
	"_examples/pkg/pidfile"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	pid := os.Getpid()
	log.Printf("pid from process %d\n", pid)
	if err := pidfile.CreateOrUpdatePIDFile(filepath.Join("app.pid")); err != nil {
		log.Fatalf("error creating or updating pid file: %v\n", err)
	}

	defer pidfile.ReleasePID()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-quit
	log.Printf("received a quit signal ...\n")

}
