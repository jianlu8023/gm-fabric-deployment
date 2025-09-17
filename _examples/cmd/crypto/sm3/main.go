package main

import (
	"fmt"

	"_examples/pkg/crypto/sm3"
)

func main() {
	hash := sm3.Hash("hello world")
	fmt.Println(string(hash))
}
