package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/str"
)

// WithNetworkListName 查看网络 filter name
func WithNetworkListName(name string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(name) {
			*args = append(*args, filters.Arg("name", name))
		}
	}
}

// WithNetworkListID 查看网络 filter id
func WithNetworkListID(id string) func(args *[]filters.KeyValuePair) {
	return func(args *[]filters.KeyValuePair) {
		if !str.IsBlank(id) {
			*args = append(*args, filters.Arg("id", id))
		}
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
