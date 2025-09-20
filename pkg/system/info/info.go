package info

import (
	"runtime"
	"time"

	humanbytes "github.com/jianlu8023/golang-example/pkg/human/bytehelper"
	"github.com/jianlu8023/golang-example/pkg/json"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type Server struct {
	Os   Os     `json:"os,omitempty" yaml:"os,omitempty"`
	Cpu  Cpu    `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Ram  Ram    `json:"ram,omitempty" yaml:"ram,omitempty"`
	Disk []Disk `json:"disk,omitempty" yaml:"disk,omitempty"`
}

func (s *Server) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

type Os struct {
	GOOS         string `json:"goos,omitempty" yaml:"goos,omitempty"`
	NumCPU       int    `json:"num_cpu,omitempty" yaml:"num_cpu,omitempty"`
	Compiler     string `json:"compiler,omitempty" yaml:"compiler,omitempty"`
	GoVersion    string `json:"go_version,omitempty" yaml:"go_version,omitempty"`
	NumGoroutine int    `json:"num_goroutine,omitempty" yaml:"num_goroutine,omitempty"`
}

func (s *Os) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

// InitOS 初始化系统信息
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

type Cpu struct {
	Cpus  []float64 `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	Cores int       `json:"cores,omitempty" yaml:"cores,omitempty"`
}

func (s *Cpu) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

// InitCPU 获取CPU信息
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

type Ram struct {
	Free           string  `json:"free,omitempty" yaml:"free,omitempty"`
	Available      string  `yaml:"available,omitempty" yaml:"available,omitempty"`
	Used           string  `json:"used,omitempty" yaml:"used,omitempty"`
	Total          string  `json:"total,omitempty" yaml:"total,omitempty"`
	UsedPercentage float64 `json:"used_percentage,omitempty" yaml:"used_percentage,omitempty"`
}

func (s *Ram) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

// InitRAM RAM信息
// @function: InitRAM
// @description: RAM信息
// @return: r Ram, err error
func InitRAM() (r Ram, err error) {
	if u, err := mem.VirtualMemory(); err != nil {
		return r, err
	} else {
		r.Free = humanbytes.HumanBytes1024(u.Free)
		r.Available = humanbytes.HumanBytes1024(u.Available)
		r.Used = humanbytes.HumanBytes1024(u.Used)
		r.Total = humanbytes.HumanBytes1024(u.Total)
		r.UsedPercentage = u.UsedPercent
	}
	return r, nil
}

type Disk struct {
	MountPoint        string  `json:"mount_point,omitempty" yaml:"mount_point,omitempty"`
	Free              string  `yaml:"free,omitempty" yaml:"free,omitempty"`
	Used              string  `json:"used,omitempty" yaml:"used,omitempty"`
	Total             string  `json:"total,omitempty" yaml:"total,omitempty"`
	UsedPercentage    float64 `json:"used_percentage,omitempty" yaml:"used_percentage,omitempty"`
	InodesFree        uint64  `json:"inodes_free,omitempty" yaml:"inodes_free,omitempty"`
	InodesUsed        uint64  `json:"inodes_used,omitempty" yaml:"inodes_used,omitempty"`
	InodesTotal       uint64  `json:"inodes_total,omitempty" yaml:"inodes_total,omitempty"`
	InodesUsedPercent float64 `json:"inodes_used_percent,omitempty" yaml:"inodes_used_percent,omitempty"`
}

func (s *Disk) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

// InitDisk 硬盘信息
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
				MountPoint:        partition.Mountpoint,
				Free:              humanbytes.HumanBytes1024(usage.Free),
				Used:              humanbytes.HumanBytes1024(usage.Used),
				Total:             humanbytes.HumanBytes1024(usage.Total),
				UsedPercentage:    usage.UsedPercent,
				InodesFree:        usage.InodesFree,
				InodesUsed:        usage.InodesUsed,
				InodesTotal:       usage.InodesTotal,
				InodesUsedPercent: usage.InodesUsedPercent,
			})
		}
		// fmt.Printf("当前挂在 %v 总容量 %v 已使用 %v 剩余 %v\n", partition.Mountpoint, usage.Total, usage.Used, usage.Free)
	}

	return d, nil
}
