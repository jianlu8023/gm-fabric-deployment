package main

import (
	"fmt"
	"os"

	"github.com/jianlu8023/golang-example/pkg/control/server"
)

func main() {

	serverControlFromFile, err := server.NewServerControlFromFile()
	if err != nil {
		fmt.Printf("generate server control from file error: %v\n", err)
		return
	}
	serverControlFromFile.StartUp(func(err error) {
		fmt.Printf("start up server control from file error: %v\n", err)
		os.Exit(1)
	})
	defer func() {
		if err := serverControlFromFile.Shutdown(); err != nil {
			fmt.Printf("shutdown server control from file error: %v\n", err)
		}
	}()
}
