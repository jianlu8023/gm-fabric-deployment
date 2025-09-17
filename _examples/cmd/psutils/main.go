package main

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"

	"_examples/internal/logger"
)

var log = logger.GetAppLogger()

func main() {
	cupInfo()
	fmt.Println()
	diskInfo()
}

func diskInfo() {
	partitions, err := disk.Partitions(true)
	if err != nil {
		log.Errorf("get partitions failed, err: %v", err)
		return
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Errorf("get usage failed, err: %v", err)
			continue
		}
		fmt.Printf("当前挂在 %v 总容量 %v 已使用 %v 剩余 %v\n", partition.Mountpoint, usage.Total, usage.Used, usage.Free)

	}
}

func cupInfo() {
	infos, err := cpu.Info()
	if err != nil {
		log.Errorf("get cpu info failed, err: %v", err)
		return
	}
	for _, info := range infos {
		fmt.Printf("cpu info: %v\n", info)
	}
}
