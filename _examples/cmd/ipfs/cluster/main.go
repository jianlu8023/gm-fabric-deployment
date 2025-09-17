package main

import (
	"context"
	"fmt"

	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"

	"github.com/multiformats/go-multiaddr"
	// _ "github.com/ugorji/go"
	// _ "google.golang.org/genproto/googleapis/rpc"
	// _ "google.golang.org/genproto/googleapis/rpc/status"
)

func main() {
	newMultiaddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/9094")
	if err != nil {
		fmt.Println(err)
		return
	}

	cl, err := client.NewDefaultClient(&client.Config{
		DisableKeepAlives: true,
		Password:          "",
		Username:          "",
		APIAddr:           newMultiaddr,
	})
	a, err := cl.Version(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("migration-cluster version :", a.Version)

}
