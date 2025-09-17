package main

import (
	"_examples/pkg/system/info"
	"fmt"
)

func main() {
	os := info.InitOS()

	disk, err := info.InitDisk()
	if err != nil {
		fmt.Println(err)
		return
	}

	ram, err := info.InitRAM()
	if err != nil {
		fmt.Println(err)
		return
	}

	cpu, err := info.InitCPU()
	if err != nil {
		fmt.Println(err)
		return
	}

	i := info.Server{
		Os:   os,
		Cpu:  cpu,
		Ram:  ram,
		Disk: disk,
	}
	fmt.Println(i)
}
