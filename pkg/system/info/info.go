package info

import (
	"runtime"
	"time"

	"github.com/jianlu8023/golang-example/pkg/json"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

type Server struct {
	Os   Os     `json:"os"`
	Cpu  Cpu    `json:"cpu"`
	Ram  Ram    `json:"ram"`
	Disk []Disk `json:"disk"`
}

func (s Server) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

type Os struct {
	GOOS         string `json:"goos"`
	NumCPU       int    `json:"numCpu"`
	Compiler     string `json:"compiler"`
	GoVersion    string `json:"goVersion"`
	NumGoroutine int    `json:"numGoroutine"`
}

func (s Os) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

type Cpu struct {
	Cpus  []float64 `json:"cpus"`
	Cores int       `json:"cores"`
}

func (s Cpu) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

type Ram struct {
	UsedMB      int `json:"usedMb"`
	TotalMB     int `json:"totalMb"`
	UsedPercent int `json:"usedPercent"`
}

func (s Ram) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

type Disk struct {
	MountPoint  string `json:"mountPoint"`
	UsedMB      int    `json:"usedMb"`
	UsedGB      int    `json:"usedGb"`
	TotalMB     int    `json:"totalMb"`
	TotalGB     int    `json:"totalGb"`
	UsedPercent int    `json:"usedPercent"`
}

func (s Disk) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

// InitOS 初始化系统信息
// @author: [SliverHorn](https://github.com/SliverHorn)
// @function: InitCPU
// @description: OS信息
// @return: o Os, err error
func InitOS() (o Os) {
	o.GOOS = runtime.GOOS
	o.NumCPU = runtime.NumCPU()
	o.Compiler = runtime.Compiler
	o.GoVersion = runtime.Version()
	o.NumGoroutine = runtime.NumGoroutine()
	return o
}

// InitCPU 获取CPU信息
// @author: [SliverHorn](https://github.com/SliverHorn)
// @function: InitCPU
// @description: CPU信息
// @return: c Cpu, err error
func InitCPU() (c Cpu, err error) {
	if cores, err := cpu.Counts(false); err != nil {
		return c, err
	} else {
		c.Cores = cores
	}
	if cpus, err := cpu.Percent(time.Duration(200)*time.Millisecond, true); err != nil {
		return c, err
	} else {
		c.Cpus = cpus
	}
	return c, nil
}

// InitRAM RAM信息
// @author: [SliverHorn](https://github.com/SliverHorn)
// @function: InitRAM
// @description: RAM信息
// @return: r Ram, err error
func InitRAM() (r Ram, err error) {
	if u, err := mem.VirtualMemory(); err != nil {
		return r, err
	} else {
		r.UsedMB = int(u.Used) / MB
		r.TotalMB = int(u.Total) / MB
		r.UsedPercent = int(u.UsedPercent)
	}
	return r, nil
}

// InitDisk 硬盘信息
// @author: [SliverHorn](https://github.com/SliverHorn)
// @function: InitDisk
// @description: 硬盘信息
// @return: d Disk, err error
func InitDisk() (d []Disk, err error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return d, err
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			return d, err
		} else {
			d = append(d, Disk{
				MountPoint:  partition.Mountpoint,
				UsedMB:      int(usage.Used) / MB,
				UsedGB:      int(usage.Used) / GB,
				TotalMB:     int(usage.Total) / MB,
				TotalGB:     int(usage.Total) / GB,
				UsedPercent: int(usage.UsedPercent),
			})
		}
		// fmt.Printf("当前挂在 %v 总容量 %v 已使用 %v 剩余 %v\n", partition.Mountpoint, usage.Total, usage.Used, usage.Free)
	}

	return d, nil
}
