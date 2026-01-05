package iphelper

import (
	"errors"
	"net"
	"strings"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"

	"github.com/gin-gonic/gin"
)

// GetClientIP 获取客户端真实IP地址
// @description 优先从X-Forwarded-For头获取，其次是X-Real-IP，最后是RemoteAddr
// @param c *gin.Context Gin上下文
// @return string 客户端IP地址
func GetClientIP(c *gin.Context) string {
	// 从X-Forwarded-For头获取IP，通常由代理服务器添加
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For格式可能是多个IP，逗号分隔，第一个是原始客户端IP
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 从X-Real-IP头获取，通常由Nginx等代理服务器设置
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 直接从连接中获取RemoteAddr
	remoteAddr := c.Request.RemoteAddr
	// 移除端口部分
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		remoteAddr = remoteAddr[:idx]
	}
	// 移除可能的[]括号（IPv6地址）
	remoteAddr = strings.TrimPrefix(strings.TrimSuffix(remoteAddr, "]"), "[")

	return remoteAddr
}

// CIDRList 存储CIDR网段列表
// @description 用于高效存储和检查IP是否在多个CIDR网段中
// 通过在初始化时解析所有CIDR，避免重复解析提高性能
type CIDRList struct {
	networks []*net.IPNet
}

// NewCIDRList 从字符串列表创建CIDR列表
// @description 创建一个新的CIDRList，支持单个IP和CIDR格式
// @param cidrStrings []string IP或CIDR字符串列表
// @return *CIDRList CIDR列表对象
// @return error 错误信息
func NewCIDRList(cidrStrings []string) (*CIDRList, error) {
	cidrList := &CIDRList{
		networks: make([]*net.IPNet, 0, len(cidrStrings)),
	}

	for _, cidrStr := range cidrStrings {
		if stringer.IsBlank(cidrStr) {
			continue
		}

		// 支持单个IP和CIDR两种格式
		if !strings.Contains(cidrStr, "/") {
			// 单个IP转换为/32或/128的CIDR
			ip := net.ParseIP(cidrStr)
			if ip == nil {
				return nil, errors.New("invalid IP address: " + cidrStr)
			}
			if ip.To4() != nil {
				cidrStr = cidrStr + "/32"
			} else {
				cidrStr = cidrStr + "/128"
			}
		}

		// 解析CIDR
		_, ipNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return nil, err
		}
		cidrList.networks = append(cidrList.networks, ipNet)
	}

	return cidrList, nil
}

// Contains 检查IP是否在任何CIDR网段中
// @description 检查给定的IP地址是否在CIDR列表中的任何网段内
// @param ip net.IP 要检查的IP地址
// @return bool IP是否在列表中
func (cl *CIDRList) Contains(ip net.IP) bool {
	for _, network := range cl.networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// IsIPInList 检查IP是否在列表中（支持精确匹配和CIDR范围匹配）
// @description 检查给定的IP地址是否存在于IP列表中，支持精确匹配和CIDR格式的IP范围匹配
// @param ip string 要检查的IP地址
// @param ipList []string IP列表，可包含精确IP或CIDR格式的IP范围
// @return bool IP是否在列表中
func IsIPInList(ip string, ipList []string) bool {
	// 解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		// IP地址格式无效，返回不在列表中
		return false
	}

	for _, listEntry := range ipList {
		listEntry = strings.TrimSpace(listEntry)
		if listEntry == "" {
			continue
		}

		// 尝试解析为CIDR格式
		_, ipNet, err := net.ParseCIDR(listEntry)
		if err == nil {
			// 成功解析为CIDR，检查IP是否在范围内
			if ipNet.Contains(parsedIP) {
				return true
			}
		} else {
			// 不是CIDR格式，进行精确匹配
			if listEntry == ip {
				return true
			}
		}
	}
	return false
}
