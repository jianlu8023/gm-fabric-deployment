package main

import (
	"fmt"
	"github.com/jianlu8023/golang-example/pkg/mycommand"
	"os/exec"
)

func main() {

	cmd := exec.Command("ls", "-al")
	out, err := mycommand.Run(cmd)
	if err != nil {
		panic(err)
	}
	fmt.Println(out.StdoutPut)
	fmt.Println(out.StderrPut)

}
