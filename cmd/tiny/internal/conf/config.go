package conf

import (
	"io"
	"os"
	"runtime"

	"github.com/jianlu8023/go-tools/v2/pkg/bytes"
	"github.com/jianlu8023/go-tools/v2/pkg/check"
)

var Config = &config{
	Output: check.IF[io.Writer](runtime.GOOS == "windows", os.Stdout, os.Stdout),
	OS:     runtime.GOOS, // 程序所在的操作系统，默认值 linux
	IP:     IP,           // 本机局域网IP
	Pwd:    "/",

	RootPath:      RootPath,      // 共享目录的根路径，默认值：当前目录
	MaxLevel:      MaxLevel,      // 允许访问的最大层级，默认值  0
	Port:          Port,          // 指定的服务端口，默认值 9090
	IsAllowUpload: IsAllowUpload, // 是否允许上传，默认值：否
	IsSecure:      IsSecure,      // 是否开启访问登录，默认值：否

	SessionVal: "abcdefghijklmnopqrstuvwxyz",
	Username:   Username, // 访问登录的帐号，默认值：admin
	Password:   Password, // 访问登录的密码，默认值：admin
}

const (
	ROOT            string = "/"
	FileGroupPrefix string = "/file"
)

const (
	VersionListURL   = "https://api.github.com/repos/TCP404/OneTiny/tags"
	VersionLatestURL = "https://api.github.com/repos/TCP404/OneTiny/releases/latest"
	VersionByTagURL  = "https://api.github.com/repos/TCP404/OneTiny/releases/tags/"
)

var (
	ReleaseName = map[string]string{
		"linux":   "OneTiny",
		"windows": "OneTiny.exe",
		"darwin":  "OneTiny_mac",
	}
)

var BufferLimit = 512 * bytes.KiB
