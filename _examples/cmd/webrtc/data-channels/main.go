package main

import (
	"_examples/cmd/webrtc/data-channels/client"
	"_examples/cmd/webrtc/data-channels/server"
	"flag"
	"fmt"
)

func main() {

	s := flag.String("type", "", "server or client")
	flag.Parse()
	fmt.Println(*s)
	if *s == "server" {
		fmt.Println("start server ...")
		server.StartServer()
	} else if *s == "client" {
		fmt.Println("start client ...")
		client.StartClient()
	} else {
		fmt.Println("please input server or client")
		return
	}
}
