package docker

import (
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	str "github.com/jianlu8023/go-tools/v2/pkg/helper/stringer"
)

// WithNetworkQueryName 查看网络 filter name
func WithNetworkQueryName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(name) {
			*args = append(*args, filters.Arg("name", name))
		}
	}
}

// WithNetworkQueryID 查看网络 filter id
func WithNetworkQueryID(id string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(id) {
			*args = append(*args, filters.Arg("id", id))
		}
	}
}

// WithNetworkCreateIPAM 配置网络的IP地址管理
func WithNetworkCreateIPAM(ipamConfig *network.IPAM) func(options *network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.IPAM = ipamConfig
	}
}

// WithNetworkCreateLabels 为网络添加标签
func WithNetworkCreateLabels(labels map[string]string) func(options *network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.Labels = labels
	}
}

// WithNetworkCreateEnableIPv6 启用IPv6
func WithNetworkCreateEnableIPv6(enable bool) func(*network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.EnableIPv6 = &enable
	}
}

func WithNetworkCreateAttachable(enable bool) func(*network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.Attachable = enable
	}
}

// ------------------------------------------------------------------------------

// WithImagePullPlatform 配置镜像拉取平台
//
// @param platform 镜像平台
//
// @return func(options *image.PullOptions) 配置函数
func WithImagePullPlatform(platform string) func(options *image.PullOptions) {
	return func(options *image.PullOptions) {
		if !str.IsBlank(platform) {
			options.Platform = platform
		}
	}
}

// WithImagePullRegistryAuth 配置镜像拉取的认证信息
//
// @param registryAuth 认证信息
//
// @return func(options *image.PullOptions) 配置函数
func WithImagePullRegistryAuth(registryAuth string) func(options *image.PullOptions) {
	return func(options *image.PullOptions) {
		if !str.IsBlank(registryAuth) {
			options.RegistryAuth = registryAuth
		}
	}
}

// WithImagePullAll 配置镜像拉取是否拉取所有镜像
// @param all bool 是否拉取所有镜像
// @return func(options *image.PullOptions) 配置函数
func WithImagePullAll(all bool) func(options *image.PullOptions) {
	return func(options *image.PullOptions) {
		if all {
			options.All = true
		}
	}
}

// --------------------------------------------------------

func WithImageQueryName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(name) {
			*args = append(*args, filters.Arg("reference", name))
		}
	}
}

// --------------------------------------------------------

// WithRemoveImageForce 配置镜像删除是否强制删除
//
// @param force 是否强制删除
//
// @return func(options *image.RemoveOptions) 配置函数
func WithRemoveImageForce(force bool) func(options *image.RemoveOptions) {
	return func(options *image.RemoveOptions) {
		options.Force = force
	}
}

// WithRemoveImagePruneChildren 配置镜像删除是否删除子镜像
//
// @param pruneChildren 是否删除子镜像
//
// @return func(options *image.RemoveOptions) 配置函数
func WithRemoveImagePruneChildren(pruneChildren bool) func(options *image.RemoveOptions) {
	return func(options *image.RemoveOptions) {
		options.PruneChildren = pruneChildren
	}
}

// ----------------------------

func WithContainerQueryName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(name) {
			*args = append(*args, filters.Arg("name", name))
		}
	}
}
