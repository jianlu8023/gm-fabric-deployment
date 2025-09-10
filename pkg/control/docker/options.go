package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/str"
)

// WithNetworkName 查看网络 filter name
func WithNetworkName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(name) {
			*args = append(*args, filters.Arg("name", name))
		}
	}
}

// WithNetworkID 查看网络 filter id
func WithNetworkID(id string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(id) {
			*args = append(*args, filters.Arg("id", id))
		}
	}
}

// WithImagePullPlatform 配置镜像拉取平台
//
// @param platform 镜像平台
//
// @return func(options *types.ImagePullOptions) 配置函数
func WithImagePullPlatform(platform string) func(options *types.ImagePullOptions) {
	return func(options *types.ImagePullOptions) {
		if !str.IsBlank(platform) {
			options.Platform = platform
		}
	}
}

// WithImagePullRegistryAuth 配置镜像拉取的认证信息
//
// @param registryAuth 认证信息
//
// @return func(options *types.ImagePullOptions) 配置函数
func WithImagePullRegistryAuth(registryAuth string) func(options *types.ImagePullOptions) {
	return func(options *types.ImagePullOptions) {
		if !str.IsBlank(registryAuth) {
			options.RegistryAuth = registryAuth
		}
	}
}

// WithRemoveImageForce 配置镜像删除是否强制删除
//
// @param force 是否强制删除
//
// @return func(options *types.ImageRemoveOptions) 配置函数
func WithRemoveImageForce(force bool) func(options *types.ImageRemoveOptions) {
	return func(options *types.ImageRemoveOptions) {
		options.Force = force
	}
}

// WithRemoveImagePruneChildren 配置镜像删除是否删除子镜像
//
// @param pruneChildren 是否删除子镜像
//
// @return func(options *types.ImageRemoveOptions) 配置函数
func WithRemoveImagePruneChildren(pruneChildren bool) func(options *types.ImageRemoveOptions) {
	return func(options *types.ImageRemoveOptions) {
		options.PruneChildren = pruneChildren
	}
}

// WithNetworkIPAM 配置网络的IP地址管理
func WithNetworkIPAM(ipamConfig *network.IPAM) func(*types.NetworkCreate) {
	return func(nc *types.NetworkCreate) {
		nc.IPAM = ipamConfig
	}
}

// WithNetworkLabels 为网络添加标签
func WithNetworkLabels(labels map[string]string) func(*types.NetworkCreate) {
	return func(nc *types.NetworkCreate) {
		nc.Labels = labels
	}
}

// WithNetworkEnableIPv6 启用IPv6
func WithNetworkEnableIPv6(enable bool) func(*types.NetworkCreate) {
	return func(nc *types.NetworkCreate) {
		nc.EnableIPv6 = enable
	}
}
