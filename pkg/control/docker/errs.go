package docker

import (
	"errors"
)

var ErrNoAliveDockerClient = errors.New("no alive docker client")

// ErrNotFound 资源未找到哨兵错误
//
// @description 当查询的网络、镜像、容器等资源不存在时返回此错误，便于调用方使用 errors.Is 判断
var ErrNotFound = errors.New("docker resource not found")

// IsNoAliveDockerClient 判断错误是否为 ErrNoAliveDockerClient
//
// @description 用于调用方判断错误类型，当 Docker 客户端未初始化或已关闭时，方法会返回 ErrNoAliveDockerClient
// @param err error 待判断的错误
// @return bool 若为 ErrNoAliveDockerClient 则返回 true，否则返回 false
func IsNoAliveDockerClient(err error) bool {
	return errors.Is(err, ErrNoAliveDockerClient)
}

// IsNotFound 判断错误是否为 ErrNotFound
//
// @description 用于调用方判断资源是否未找到
// @param err error 待判断的错误
// @return bool 若为 ErrNotFound 则返回 true，否则返回 false
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
