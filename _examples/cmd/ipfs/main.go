package main

import (
	"fmt"

	shell "github.com/ipfs/go-ipfs-api"
)

func main() {
	newShell := shell.NewShell("http://localhost:5001")
	up := newShell.IsUp()
	fmt.Println("migration is up :", up)
	version, commit, err := newShell.Version()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("migration version :", fmt.Sprintf("%v+git%v", version, commit))
}
