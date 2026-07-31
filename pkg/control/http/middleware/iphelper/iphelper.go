package iphelper

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/jianlu8023/go-tools/v2/pkg/stringer"

	"github.com/gin-gonic/gin"
)

// GetClientIP 获取客户端真实 IP 地址
//
// @description 基于"可信代理"机制解析客户端真实 IP，防止 X-Forwarded-For 头被伪造绕过 IP 黑白名单与限流。
// 解析顺序：
//  1. 从 RemoteAddr 提取直连 IP（使用 net.SplitHostPort，正确处理 IPv6）；
//  2. 若未提供可信代理列表（trustedProxies 为 nil 或空），直接返回 RemoteAddr 的 IP，不信任任何 XFF 头；
//  3. 若 RemoteAddr 命中可信代理网段，则从 X-Forwarded-For 中自右向左跳过可信代理 IP，返回第一个非可信 IP；
//  4. 若 X-Forwarded-For 全部为可信代理 IP，返回最左侧的 IP（即原始客户端 IP）；
//  5. 若 X-Forwarded-For 缺失，回退到 X-Real-IP；
//  6. 仍无法获取时，返回 RemoteAddr 的 IP。
//
// @param c *gin.Context Gin 上下文
// @param trustedProxies *CIDRList 可信代理网段列表，为 nil 表示不信任任何代理
// @return string 客户端 IP 地址
func GetClientIP(c *gin.Context, trustedProxies *CIDRList) string {
	remoteIP := getRemoteIP(c.Request)

	// 无可信代理配置，直接返回 RemoteAddr 的 IP（最安全的默认行为）
	if trustedProxies == nil || len(trustedProxies.networks) == 0 {
		return remoteIP
	}

	// RemoteAddr 不在可信代理列表中，客户端直连，忽略 XFF 头
	remoteIPParsed := net.ParseIP(remoteIP)
	if remoteIPParsed == nil || !trustedProxies.Contains(remoteIPParsed) {
		return remoteIP
	}

	// RemoteAddr 是可信代理，从 X-Forwarded-For 中解析真实客户端 IP
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		// 自右向左跳过可信代理 IP，找到第一个非可信 IP
		for i := len(ips) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(ips[i])
			if ip == "" {
				continue
			}
			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				continue
			}
			if !trustedProxies.Contains(parsedIP) {
				return ip
			}
		}
		// 所有 IP 都是可信代理，返回最左侧的 IP（原始客户端 IP）
		for i := 0; i < len(ips); i++ {
			ip := strings.TrimSpace(ips[i])
			if ip != "" && net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// X-Forwarded-For 缺失或全部非法，回退到 X-Real-IP
	if realIP := strings.TrimSpace(c.GetHeader("X-Real-IP")); realIP != "" && net.ParseIP(realIP) != nil {
		return realIP
	}

	return remoteIP
}

// getRemoteIP 从请求的 RemoteAddr 中提取 IP 地址
//
// @description 使用 net.SplitHostPort 解析，正确处理 IPv6 地址（如 [::1]:8080）
// @param req *http.Request HTTP 请求
// @return string IP 地址（不含端口）
func getRemoteIP(req *http.Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		// 没有 port 或者格式异常，直接返回 RemoteAddr
		return req.RemoteAddr
	}
	return host
}

// CIDRList 存储CIDR网段列表
//
// @description 用于高效存储和检查IP是否在多个CIDR网段中
// 通过在初始化时解析所有CIDR，避免重复解析提高性能
type CIDRList struct {
	networks []*net.IPNet
}

// NewCIDRList 从字符串列表创建CIDR列表
//
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

// Len 返回 CIDR 列表中的网段数量
//
// @description 用于判断列表是否为空
// @return int 网段数量
func (cl *CIDRList) Len() int {
	return len(cl.networks)
}

// Contains 检查IP是否在任何CIDR网段中
//
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
