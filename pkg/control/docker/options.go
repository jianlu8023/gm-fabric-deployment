package docker

import (
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// WithNetworkQueryName 查看网络 filter name
//
// @description 根据网络名称构造 ListNetworks/GetNetwork 的查询过滤器
// @param name string 网络名称
// @return func(args *[]filters.KeyValuePair) 过滤器配置函数
func WithNetworkQueryName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !stringer.IsBlank(name) {
			*args = append(*args, filters.Arg("name", name))
		}
	}
}

// WithNetworkQueryID 查看网络 filter id
//
// @description 根据网络ID构造 ListNetworks/GetNetwork 的查询过滤器
// @param id string 网络ID
// @return func(args *[]filters.KeyValuePair) 过滤器配置函数
func WithNetworkQueryID(id string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !stringer.IsBlank(id) {
			*args = append(*args, filters.Arg("id", id))
		}
	}
}

// WithNetworkCreateIPAM 配置网络的IP地址管理
//
// @description 为 CreateNetwork 设置 IPAM（IP Address Management）配置
// @param ipamConfig *network.IPAM IPAM配置
// @return func(options *network.CreateOptions) 创建选项配置函数
func WithNetworkCreateIPAM(ipamConfig *network.IPAM) func(options *network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.IPAM = ipamConfig
	}
}

// WithNetworkCreateLabels 为网络添加标签
//
// @description 为 CreateNetwork 设置自定义标签（Labels）
// @param labels map[string]string 标签键值对
// @return func(options *network.CreateOptions) 创建选项配置函数
func WithNetworkCreateLabels(labels map[string]string) func(options *network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.Labels = labels
	}
}

// WithNetworkCreateEnableIPv6 启用IPv6
//
// @description 为 CreateNetwork 开启或关闭 IPv6 支持
// @param enable bool 是否启用IPv6
// @return func(*network.CreateOptions) 创建选项配置函数
func WithNetworkCreateEnableIPv6(enable bool) func(*network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.EnableIPv6 = &enable
	}
}

// WithNetworkCreateAttachable 配置网络是否允许手动附加容器
//
// @description 为 CreateNetwork 设置 Attachable 属性，非 swarm 模式下通常配合 bridge 驱动使用
// @param enable bool 是否允许手动附加容器
// @return func(*network.CreateOptions) 创建选项配置函数
func WithNetworkCreateAttachable(enable bool) func(*network.CreateOptions) {
	return func(nc *network.CreateOptions) {
		nc.Attachable = enable
	}
}

// ------------------------------------------------------------------------------

// WithImagePullPlatform 配置镜像拉取平台
//
// @description 为 PullImage 指定拉取的目标平台（如 linux/amd64、linux/arm64）
// @param platform 镜像平台
// @return func(options *image.PullOptions) 配置函数
func WithImagePullPlatform(platform string) func(options *image.PullOptions) {
	return func(options *image.PullOptions) {
		if !stringer.IsBlank(platform) {
			options.Platform = platform
		}
	}
}

// WithImagePullRegistryAuth 配置镜像拉取的认证信息
//
// @description 为 PullImage 设置私有仓库的认证字符串（Base64 编码的 user:password）
// @param registryAuth 认证信息
// @return func(options *image.PullOptions) 配置函数
func WithImagePullRegistryAuth(registryAuth string) func(options *image.PullOptions) {
	return func(options *image.PullOptions) {
		if !stringer.IsBlank(registryAuth) {
			options.RegistryAuth = registryAuth
		}
	}
}

// WithImagePullAll 配置镜像拉取是否拉取所有镜像
//
// @description 为 PullImage 设置是否拉取仓库中所有 Tag 的镜像
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

// WithImageQueryName 根据镜像名称构造查询过滤器
//
// @description 为 ListImages/GetImage 按镜像 reference（仓库/标签）过滤
// @param name string 镜像名称（含 tag 或 digest）
// @return func(args *[]filters.KeyValuePair) 过滤器配置函数
func WithImageQueryName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !stringer.IsBlank(name) {
			*args = append(*args, filters.Arg("reference", name))
		}
	}
}

// --------------------------------------------------------

// WithRemoveImageForce 配置镜像删除是否强制删除
//
// @description 为 RemoveImage 设置是否强制删除（即使镜像被容器引用）
// @param force 是否强制删除
// @return func(options *image.RemoveOptions) 配置函数
func WithRemoveImageForce(force bool) func(options *image.RemoveOptions) {
	return func(options *image.RemoveOptions) {
		options.Force = force
	}
}

// WithRemoveImagePruneChildren 配置镜像删除是否删除子镜像
//
// @description 为 RemoveImage 设置是否在删除父镜像后一并清理无标签的子镜像
// @param pruneChildren 是否删除子镜像
// @return func(options *image.RemoveOptions) 配置函数
func WithRemoveImagePruneChildren(pruneChildren bool) func(options *image.RemoveOptions) {
	return func(options *image.RemoveOptions) {
		options.PruneChildren = pruneChildren
	}
}

// ----------------------------

// WithContainerQueryName 根据容器名称构造查询过滤器
//
// @description 为 ListContainers/GetContainer 按容器名称过滤
// @param name string 容器名称
// @return func(args *[]filters.KeyValuePair) 过滤器配置函数
func WithContainerQueryName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !stringer.IsBlank(name) {
			*args = append(*args, filters.Arg("name", name))
		}
	}
}
